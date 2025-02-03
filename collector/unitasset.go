package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/thing.go

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
)

// A UnitAsset models an interface or API for a smaller part of a whole system,
// for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
//
// This unit asset represents a room's statistics cache, where all data from the
// other system are gathered, before being sent off to the InfluxDB service.
type unitAsset struct {
	// Public fields
	// TODO: Why have these public and then provide getter methods? Might need refactor..
	Name        string              `json:"name"`    // Must be a unique name, ie. a sensor ID
	Owner       *components.System  `json:"-"`       // The parent system this UA is part of
	Details     map[string][]string `json:"details"` // Metadata or details about this UA
	ServicesMap components.Services `json:"-"`       // Services provided to consumers
	CervicesMap components.Cervices `json:"-"`       // Services being consumed

	// Internal fields this UA might need to perform it's function
	InfluxDBHost     string `json:"influxdb_host"`
	InfluxDBToken    string `json:"influxdb_token"`
	CollectionPeriod int    `json:"collection_period"`
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
func initTemplate() components.UnitAsset {
	return &unitAsset{
		Name:    uaName,
		Details: map[string][]string{"Location": {"Kitchen"}},

		InfluxDBHost:     "localhost:", // TODO: add port
		InfluxDBToken:    "insert secret token here",
		CollectionPeriod: 30,
	}
}

// newUnitAsset creates a new and proper instance of UnitAsset, using settings and
// values loaded from an existing configuration file.
// This function returns an UA instance that is ready to be published and used,
// aswell as a function that can ...
// TODO: complete doc and remove servs here and in the system file
func newUnitAsset(uac unitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func() error) {
	// Lost of consumed services
	sProtocols := components.SProtocols(sys.Husk.ProtoPort)
	temp := &components.Cervice{
		Name:   "temperature",
		Protos: sProtocols,
		Url:    make([]string, 0),
	}
	price := &components.Cervice{
		Name:   "SEK_price",
		Protos: sProtocols,
		Url:    make([]string, 0),
	}
	desired := &components.Cervice{
		Name:   "desired_temp",
		Protos: sProtocols,
		Url:    make([]string, 0),
	}
	setpoint := &components.Cervice{
		Name:   "setpoint",
		Protos: sProtocols,
		Url:    make([]string, 0),
	}

	ua := &unitAsset{
		Name:    uac.Name,
		Owner:   sys,
		Details: uac.Details,
		// ServicesMap: components.CloneServices(servs), // TODO: not required?
		CervicesMap: components.Cervices{
			temp.Name:     temp,
			price.Name:    price,
			desired.Name:  desired,
			setpoint.Name: setpoint,
		},

		InfluxDBHost:     uac.InfluxDBHost,
		InfluxDBToken:    uac.InfluxDBToken,
		CollectionPeriod: uac.CollectionPeriod,
		// TODO: other internal fields
	}

	// TODO: required for matching values with locations?
	// ua.CervicesMap["temperature"].Details = components.MergeDetails(ua.Details, ref.Details)
	// for _, cs := range ua.CervicesMap {
	// TODO: or merge it with an empty map if this doesn't work...
	// cs.Details = ua.Details
	// }

	// Returns the loaded unit asset and an function to handle optional cleanup at shutdown
	return ua, ua.startup
}

////////////////////////////////////////////////////////////////////////////////

var errTooShortPeriod error = fmt.Errorf("collection period less than 1 second")

func (ua *unitAsset) startup() (err error) {
	if ua.CollectionPeriod < 1 {
		return errTooShortPeriod
	}

	for {
		select {
		// Wait for a shutdown signal
		case <-ua.Owner.Ctx.Done():
			ua.cleanup()
			return

			// Wait until it's time to collect new data
		case <-time.Tick(time.Duration(ua.CollectionPeriod) * time.Second):
			if err = ua.collect(); err != nil {
				return
			}
		}
	}
}

func (ua *unitAsset) cleanup() {
	// TODO: remove this func altogether, if it's not required later on
}

func (ua *unitAsset) collect() (err error) {
	log.Println("tick")
	return nil
}
