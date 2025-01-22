package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/thing.go

import (
	"log"

	"github.com/sdoque/mbaigo/components"
)

// A UnitAsset models an interface or API for a smaller part of a whole system, for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
type UnitAsset struct {
	// Public fields
	// TODO: Why have these public and then provide getter methods? Might need refactor..
	Name        string              `json:"name"`    // Must be a unique name, ie. a sensor ID
	Owner       *components.System  `json:"-"`       // The parent system this UA is part of
	Details     map[string][]string `json:"details"` // Metadata or details about this UA
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`

	// Internal fields this UA might need to perform it's function, example:
	temperature float64
}

func (ua *UnitAsset) GetName() string {
	return ua.Name
}

func (ua *UnitAsset) GetDetails() map[string][]string {
	return ua.Details
}

// GetServices returns all services and capabilities this UnitAsset is providing to consumers.
func (ua *UnitAsset) GetServices() components.Services {
	return ua.ServicesMap
}

// GetCervices returns the list of services that is being consumed by this UnitAsset.
func (ua *UnitAsset) GetCervices() components.Cervices {
	return ua.CervicesMap
}

// ensure UnitAsset implements the components.UnitAsset interface (this check is done at compile time)
var _ components.UnitAsset = (*UnitAsset)(nil)

////////////////////////////////////////////////////////////////////////////////

// initTemplate initializes a new UA and prefils it with some default values.
// The returned instance is used for generating the configuration file, whenever it's missing.
func initTemplate() components.UnitAsset {
	// First predefine any exposed services
	// (see https://github.com/sdoque/mbaigo/blob/main/components/service.go for documentation)
	temperature := components.Service{
		Definition:  "temperature-def",                             // TODO: this get's incorrectly linked to the below subpath
		SubPath:     "temperature-sub",                             // TODO: this path needs to be setup in Serving() too
		Details:     map[string][]string{"Forms": {"SignalA_v1a"}}, // TODO: why this form here??
		RegPeriod:   30,
		Description: "provides the current temperature of this sensor (using a GET request)",
	}

	return &UnitAsset{
		// TODO: These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name: "temperature-UA",
		Details: map[string][]string{
			"Unit":     {"Celsius"},
			"Location": {"Kitchen"},
		},
		// Don't forget to map the provided services from above!
		ServicesMap: components.Services{
			temperature.SubPath: &temperature,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// newUnitAsset creates a new and proper instance of UnitAsset, using settings and
// values loaded from an existing configuration file.
// This function returns an UA instance that is ready to be published and used,
// aswell as a function that can perform any cleanup when the system is shutting down.
func newUnitAsset(uac UnitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func()) {
	ua := &UnitAsset{
		// Filling in public fields using the given data
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),

		// Setting the example variable
		temperature: 3.14,
	}

	// Optionally start background tasks here! Example:
	go func() {
		log.Println("Starting up " + ua.Name)
	}()

	// Returns the loaded unit asset and an function to handle optional cleanup at shutdown
	return ua, func() {
		log.Println("Cleaning up " + ua.Name)
	}
}
