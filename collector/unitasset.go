package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/thing.go

import (
	"fmt"
	"log"
	"net/http"
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

	InfluxDBHost         string `json:"influxdb_host"`  // IP:port addr to the influxdb server
	InfluxDBToken        string `json:"influxdb_token"` // Auth token
	InfluxDBOrganisation string `json:"influxdb_organisation"`
	InfluxDBBucket       string `json:"influxdb_bucket"`   // Data bucket
	CollectionPeriod     int    `json:"collection_period"` // Period (in seconds) between each data collection

	// Mockable function for getting states from the consumed services.
	apiGetState func(*components.Cervice, *components.System) (forms.Form, error)

	//
	influx       influxdb2.Client
	influxWriter api.WriteAPI
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
// func initTemplate() components.UnitAsset {
func initTemplate() *unitAsset {
	return &unitAsset{
		Name:    uaName,
		Details: map[string][]string{"Location": {"Kitchen"}},

		InfluxDBHost:         "http://localhost:8086",
		InfluxDBToken:        "insert secret token here",
		InfluxDBOrganisation: "organisation",
		InfluxDBBucket:       "arrowhead",
		CollectionPeriod:     30,
	}
}

var consumeServices []string = []string{
	"temperature",
	"SEKPrice",
	"DesiredTemp",
	"setpoint",
}

// newUnitAsset creates a new and proper instance of UnitAsset, using settings and
// values loaded from an existing configuration file.
// This function returns an UA instance that is ready to be published and used,
// aswell as a function that can ...
// TODO: complete doc and remove servs here and in the system file
// func newUnitAsset(uac unitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func() error) {
// func newUnitAsset(uac unitAsset, sys *components.System, servs []components.Service) *unitAsset {
func newUnitAsset(uac unitAsset, sys *system, servs []components.Service) *unitAsset {
	client := influxdb2.NewClientWithOptions(
		uac.InfluxDBHost, uac.InfluxDBToken,
		influxdb2.DefaultOptions().SetHTTPClient(http.DefaultClient),
	)

	ua := &unitAsset{
		Name:    uac.Name,
		Owner:   &sys.System,
		Details: uac.Details,
		// ServicesMap: components.CloneServices(servs), // TODO: not required?
		CervicesMap: components.Cervices{},

		InfluxDBHost:         uac.InfluxDBHost,
		InfluxDBToken:        uac.InfluxDBToken,
		InfluxDBOrganisation: uac.InfluxDBOrganisation,
		InfluxDBBucket:       uac.InfluxDBBucket,
		CollectionPeriod:     uac.CollectionPeriod,

		apiGetState:  usecases.GetState,
		influx:       client,
		influxWriter: client.WriteAPI(uac.InfluxDBOrganisation, uac.InfluxDBBucket),
	}

	// TODO: handle influx write errors or don't care?

	// Prep all the consumed services
	protos := components.SProtocols(sys.Husk.ProtoPort)
	for _, service := range consumeServices {
		ua.CervicesMap[service] = &components.Cervice{
			Name:   service,
			Protos: protos,
			Url:    make([]string, 0),
		}
	}

	// TODO: required for matching values with locations?
	// ua.CervicesMap["temperature"].Details = components.MergeDetails(ua.Details, nil)
	// for _, cs := range ua.CervicesMap {
	// TODO: or merge it with an empty map if this doesn't work...
	// cs.Details = ua.Details
	// }

	// Returns the loaded unit asset and an function to handle optional cleanup at shutdown
	// return ua, ua.startup
	return ua
}

////////////////////////////////////////////////////////////////////////////////

var errTooShortPeriod error = fmt.Errorf("collection period less than 1 second")

func (ua *unitAsset) startup() (err error) {
	if ua.CollectionPeriod < 1 {
		return errTooShortPeriod
	}

	// TODO: try connecting to influx, check if need to call Health()/Ping()/Ready()/Setup()?

	for {
		select {
		// Wait for a shutdown signal
		case <-ua.Owner.Ctx.Done():
			ua.cleanup()
			return

			// Wait until it's time to collect new data
		case <-time.Tick(time.Duration(ua.CollectionPeriod) * time.Second):
			if err = ua.collectAllServices(); err != nil {
				return
			}
		}
	}
}

func (ua *unitAsset) cleanup() {
	ua.influx.Close()
}

func (ua *unitAsset) collectAllServices() (err error) {
	// log.Println("tick") // TODO
	var wg sync.WaitGroup

	for _, service := range consumeServices {
		wg.Add(1)
		go func(s string) {
			if err := ua.collectService(s); err != nil {
				log.Printf("Error collecting data from %s: %s", s, err)
			}
			wg.Done()
		}(service)
	}

	wg.Wait()
	ua.influxWriter.Flush()
	return nil
}

func (ua *unitAsset) collectService(service string) (err error) {
	f, err := ua.apiGetState(ua.CervicesMap[service], ua.Owner)
	if err != nil {
		return // TODO: use a better error?
	}
	// fmt.Println(f)
	s, ok := f.(*forms.SignalA_v1a)
	if !ok {
		err = fmt.Errorf("bad form version: %s", f.FormVersion())
		return
	}
	// fmt.Println(s) // TODO

	p := influxdb2.NewPoint(
		service,
		map[string]string{"unit": s.Unit},
		map[string]interface{}{"value": s.Value},
		s.Timestamp.UTC(),
	)
	// fmt.Println(p)

	ua.influxWriter.WritePoint(p)
	return nil
}
