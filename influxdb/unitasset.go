package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/thing.go

import (
	"log"
	"net/http"

	"github.com/sdoque/mbaigo/components"
)

// A UnitAsset models an interface or API for a smaller part of a whole system, for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
type unitAsset struct {
	// Public fields
	// TODO: Why have these public and then provide getter methods? Might need refactor..
	Name        string              `json:"name"`    // Must be a unique name, ie. a sensor ID
	Owner       *components.System  `json:"-"`       // The parent system this UA is part of
	Details     map[string][]string `json:"details"` // Metadata or details about this UA
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`

	// Internal fields this UA might need to perform it's function
}

func (ua *unitAsset) GetName() string {
	return ua.Name
}

func (ua *unitAsset) GetDetails() map[string][]string {
	return ua.Details
}

// GetServices returns all services and capabilities this UnitAsset is providing to consumers.
func (ua *unitAsset) GetServices() components.Services {
	return ua.ServicesMap
}

// GetCervices returns the list of services that is being consumed by this UnitAsset.
func (ua *unitAsset) GetCervices() components.Cervices {
	return ua.CervicesMap
}

// ensure UnitAsset implements the components.UnitAsset interface (this check is done at compile time)
var _ components.UnitAsset = (*unitAsset)(nil)

////////////////////////////////////////////////////////////////////////////////

// initTemplate initializes a new UA and prefils it with some default values.
// The returned instance is used for generating the configuration file, whenever it's missing.
func initTemplate() components.UnitAsset {
	return &unitAsset{
		Name: "InfluxDB collector",
	}
}

// newUnitAsset creates a new and proper instance of UnitAsset, using settings and
// values loaded from an existing configuration file.
// This function returns an UA instance that is ready to be published and used,
// aswell as a function that can perform any cleanup when the system is shutting down.
func newUnitAsset(uac unitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func()) {
	ua := &unitAsset{
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),
	}
	// Returns the loaded unit asset and an function to handle optional cleanup at shutdown
	return ua, func() {
		log.Println("Cleaning up " + ua.Name)
	}
}

////////////////////////////////////////////////////////////////////////////////

// Serving maps the requested service paths with any request handlers.
func (ua *unitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	default:
		// TODO: should instead tell the visitor that there's no services published?
		http.Error(w, "Invalid service request", http.StatusBadRequest)
	}
}
