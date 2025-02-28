package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/thing.go

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

// A UnitAsset models an interface or API for a smaller part of a whole system,
// for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
//
// This unit asset represents a room's statistics cache, where all data from the
// other system are gathered, before being sent off to the InfluxDB service.
type unitAsset struct {
	// These fields are required by the interface (below) and must also be public
	// for being included in the systemconfig JSON file.
	Name        string              `json:"name"`    // Must be a unique name, ie. a sensor ID
	Owner       *components.System  `json:"-"`       // The parent system this UA is part of
	Details     map[string][]string `json:"details"` // Metadata or details about this UA
	ServicesMap components.Services `json:"-"`       // Services provided to consumers
	CervicesMap components.Cervices `json:"-"`       // Services being consumed

	InfluxDBHost         string   `json:"influxdb_host"`  // IP:port addr to the influxdb server
	InfluxDBToken        string   `json:"influxdb_token"` // Auth token
	InfluxDBOrganisation string   `json:"influxdb_organisation"`
	InfluxDBBucket       string   `json:"influxdb_bucket"`   // Data bucket
	CollectionPeriod     int      `json:"collection_period"` // Period (in seconds) between each data collection
	Samples              []Sample `json:"samples"`           // Arrowhead services we want to sample data from

	// Mockable function for getting states from the consumed services.
	apiGetState func(*components.Cervice, *components.System) (forms.Form, error)

	// internal things for talking with Influx
	influx       influxdb2.Client
	influxWriter api.WriteAPI
}

// A Sample is a struct that defines a service to be sampled.
// The service sampled is identified using the details map.
// Inspired from:
// https://github.com/vanDeventer/metalepsis/blob/9752ee11657a44fd701e3c3b4f75c592d001a5e5/Influxer/thing.go#L38
type Sample struct {
	Service string              `json:"service"`
	Details map[string][]string `json:"details"`
}

// Following methods are required by the interface components.UnitAsset.
// Enforce a compile-time check that the interface is implemented correctly.
var _ components.UnitAsset = (*unitAsset)(nil)

func (ua *unitAsset) GetName() string {
	return ua.Name
}

func (ua *unitAsset) GetDetails() map[string][]string {
	return ua.Details
}

func (ua *unitAsset) GetServices() components.Services {
	return ua.ServicesMap
}

func (ua *unitAsset) GetCervices() components.Cervices {
	return ua.CervicesMap
}

func (ua *unitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	// We don't provide any services!
	http.Error(w, "No services available", http.StatusNotImplemented)
}

////////////////////////////////////////////////////////////////////////////////

const uaName string = "Cache"

// initTemplate initializes a new UA and prefils it with some default values.
// The returned instance is used for generating the configuration file, whenever it's missing.
func initTemplate() *unitAsset {
	return &unitAsset{
		Name:                 uaName,
		InfluxDBHost:         "http://localhost:8086",
		InfluxDBToken:        "insert secret token here",
		InfluxDBOrganisation: "organisation",
		InfluxDBBucket:       "arrowhead",
		CollectionPeriod:     30,
		Samples: []Sample{
			{"temperature", map[string][]string{"Location": {"Kitchen"}}},
			{"SEKPrice", map[string][]string{"Location": {"Kitchen"}}},
			{"DesiredTemp", map[string][]string{"Location": {"Kitchen"}}},
			{"setpoint", map[string][]string{"Location": {"Kitchen"}}},
			{"consumption", map[string][]string{"Location": {"Kitchen"}}},
			{"state", map[string][]string{"Location": {"Kitchen"}}},
			{"power", map[string][]string{"Location": {"Kitchen"}}},
			{"current", map[string][]string{"Location": {"Kitchen"}}},
			{"voltage", map[string][]string{"Location": {"Kitchen"}}},
		},
	}
}

// newUnitAsset creates a new instance of UnitAsset, using settings and values
// loaded from an existing configuration file.
// Returns an UA instance that is ready to be published and used by others.
func newUnitAsset(uac unitAsset, sys *system) *unitAsset {
	client := influxdb2.NewClientWithOptions(
		uac.InfluxDBHost, uac.InfluxDBToken,
		influxdb2.DefaultOptions().SetHTTPClient(http.DefaultClient),
	)

	ua := &unitAsset{
		Name:        uac.Name,
		Owner:       &sys.System,
		Details:     uac.Details,
		CervicesMap: components.Cervices{},

		InfluxDBHost:         uac.InfluxDBHost,
		InfluxDBToken:        uac.InfluxDBToken,
		InfluxDBOrganisation: uac.InfluxDBOrganisation,
		InfluxDBBucket:       uac.InfluxDBBucket,
		CollectionPeriod:     uac.CollectionPeriod,
		Samples:              uac.Samples,

		// Default to using the API method, outside of tests.
		apiGetState: usecases.GetState,
		influx:      client,
		// "[The async] WriteAPI automatically logs write errors." Source:
		// https://pkg.go.dev/github.com/influxdata/influxdb-client-go/v2#readme-reading-async-errors
		influxWriter: client.WriteAPI(uac.InfluxDBOrganisation, uac.InfluxDBBucket),
	}

	// Maps the services we want to sample. The services will then be looked up
	// using the Orchestrator.
	// Again based on code from VanDeventer.
	for _, s := range ua.Samples {
		ua.CervicesMap[s.Service] = &components.Cervice{
			Name:    s.Service,
			Details: s.Details,
			Url:     make([]string, 0),
		}
	}

	return ua
}

////////////////////////////////////////////////////////////////////////////////

var errTooShortPeriod error = fmt.Errorf("collection period less than 1 second")

func (ua *unitAsset) startup() (err error) {
	if ua.CollectionPeriod < 1 {
		return errTooShortPeriod
	}

	// Make sure we can contact the influxdb server, before trying to do any thing else
	running, err := ua.influx.Ping(ua.Owner.Ctx)
	if err != nil {
		return fmt.Errorf("ping influxdb: %w", err)
	} else if !running {
		return fmt.Errorf("influxdb not running")
	}

	for {
		select {
		// Wait for a shutdown signal
		case <-ua.Owner.Ctx.Done():
			ua.cleanup()
			return

		// Wait until it's time to collect new data
		case <-time.Tick(time.Duration(ua.CollectionPeriod) * time.Second):
			if err = ua.collectAllServices(); err != nil {
				log.Println("Error: ", err)
			}
		}
	}
}

func (ua *unitAsset) cleanup() {
	ua.influx.Close()
}

func (ua *unitAsset) collectAllServices() (err error) {
	var wg sync.WaitGroup
	for _, sample := range ua.Samples {
		wg.Add(1)
		go func(s Sample) {
			if e := ua.collectService(s); e != nil {
				err = fmt.Errorf("collecting data from %s: %w", s, e)
			}
			wg.Done()
		}(sample)
	}

	// Errors from the writer are caught in another goroutine and logged there
	wg.Wait()
	ua.influxWriter.Flush()
	return
}

func (ua *unitAsset) collectService(sam Sample) (err error) {
	f, err := ua.apiGetState(ua.CervicesMap[sam.Service], ua.Owner)
	if err != nil {
		return fmt.Errorf("failed to get state: %w", err)
	}
	sig, ok := f.(*forms.SignalA_v1a)
	if !ok {
		err = fmt.Errorf("bad form version: %s", f.FormVersion())
		return
	}

	ua.influxWriter.WritePoint(influxdb2.NewPoint(
		sam.Service,
		map[string]string{
			"unit":     sig.Unit,
			"location": strings.Join(sam.Details["Location"], "-"),
		},
		map[string]interface{}{"value": sig.Value},
		sig.Timestamp.UTC(),
	))
	return nil
}
