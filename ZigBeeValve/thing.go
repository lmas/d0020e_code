/* In order to follow the structure of the other systems made before this one, most functions and structs are copied and slightly edited from:
 * https://github.com/sdoque/systems/blob/main/thermostat/thing.go */

package main

import (
	"log"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
)

//-------------------------------------Define the unit asset

// UnitAsset type models the unit asset (interface) of the system
type UnitAsset struct {
	Name        string              `json:"name"`
	Owner       *components.System  `json:"-"`
	Details     map[string][]string `json:"details"`
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`
	//
	Setpt float64 `json:"setpoint"`
}

// GetName returns the name of the Resource.
func (ua *UnitAsset) GetName() string {
	return ua.Name
}

// GetServices returns the services of the Resource.
func (ua *UnitAsset) GetServices() components.Services {
	return ua.ServicesMap
}

// GetCervices returns the list of consumed services by the Resource.
func (ua *UnitAsset) GetCervices() components.Cervices {
	return ua.CervicesMap
}

// GetDetails returns the details of the Resource.
func (ua *UnitAsset) GetDetails() map[string][]string {
	return ua.Details
}

// ensure UnitAsset implements components.UnitAsset (this check is done at during the compilation)
var _ components.UnitAsset = (*UnitAsset)(nil)

//-------------------------------------Instatiate a unit asset template

// initTemplate initializes a UnitAsset with default values.
func initTemplate() components.UnitAsset {
	setPointService := components.Service{
		Definition:  "setpoint",
		SubPath:     "setpoint",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current thermal setpoint (GET) or sets it (PUT)",
	}

	getAmpereUsage := components.Service{ //current use of Ampere (change name)
		Definition:  "usage",
		SubPath:     "usage",
		Details:     map[string][]string{"Unit": {"Ampere"}, "Forms": {"SignalA_v1a"}},
		CUnit:       "Eur/kWh",
		Description: "provides current use of Amperes",
	}
	// add more shit, like Wh, W, V etc

	// var uat components.UnitAsset // this is an interface, which we then initialize
	uat := &UnitAsset{
		Name:    "ZigBeeValve",
		Details: map[string][]string{"Location": {"Kitchen"}},
		Setpt:   20,
		ServicesMap: components.Services{
			setPointService.SubPath: &setPointService,
			measure.SubPath:         &getAmpereUsage,
		},
	}
	return uat
}

//-------------------------------------Instatiate the unit assets based on configuration

// newResource creates the Resource resource with its pointers and channels based on the configuration using the tConig structs
func newResource(uac UnitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func()) {
	// deterimine the protocols that the system supports
	sProtocols := components.SProtocols(sys.Husk.ProtoPort)
	// instantiate the consumed services
	t := &components.Cervice{
		Name:   "temperature",
		Protos: sProtocols,
		Url:    make([]string, 0),
	}

	// intantiate the unit asset
	ua := &UnitAsset{
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),
		Setpt:       uac.Setpt,
		CervicesMap: components.Cervices{
			t.Name: t,
		},
	}

	var ref components.Service
	for _, s := range servs {
		if s.Definition == "setpoint" {
			ref = s
		}
	}

	ua.CervicesMap["temperature"].Details = components.MergeDetails(ua.Details, ref.Details)

	return ua, func() {
		log.Println("Shutting down thermostat ", ua.Name)
	}
}

//-------------------------------------Thing's resource methods

// getSetPoint fills out a signal form with the current thermal setpoint
func (ua *UnitAsset) getSetPoint() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Setpt
	f.Unit = "Celcius"
	f.Timestamp = time.Now()
	return f
}

// setSetPoint updates the thermal setpoint
func (ua *UnitAsset) setSetPoint(f forms.SignalA_v1a) {
	ua.Setpt = f.Value
	log.Printf("new set point: %.1f", f.Value)
}
