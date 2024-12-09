package main

import (
	"log"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
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
	//
	Period    time.Duration `json:"samplingPeriod"`
	deviation float64
	//
	SEK_price float64 `json:"SEK_per_kWh"`
	Min_price float64 `json:"min_price"`
	Max_price float64 `json:"max_price"`
	Min_temp  float64 `json:"min_temp"`
	Max_temp  float64 `json:"max_temp"`
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

////////////////////////////////////////////////////////////////////////////////

// initTemplate initializes a new UA and prefils it with some default values.
// The returned instance is used for generating the configuration file, whenever it's missing.
func initTemplate() components.UnitAsset {
	// First predefine any exposed services
	// (see https://github.com/sdoque/mbaigo/blob/main/components/service.go for documentation)
	setSEK_price := components.Service{
		Definition:  "SEK_price",
		SubPath:     "SEK_price",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current electric hourly price (using a GET request)",
	}

	setMax_temp := components.Service{
		Definition:  "max_temperature",                                                  // TODO: this get's incorrectly linked to the below subpath
		SubPath:     "max_temperature",                                                  // TODO: this path needs to be setup in Serving() too
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}}, // TODO: why this form here??
		Description: "provides the maximum temp the user wants (using a GET request)",
	}
	setMin_temp := components.Service{
		Definition:  "min_temperature",
		SubPath:     "min_temperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum temp the user could tolerate (using a GET request)",
	}
	setMax_price := components.Service{
		Definition:  "max_price",
		SubPath:     "max_price",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the maximum price the user wants to pay (using a GET request)",
	}
	setMin_price := components.Service{
		Definition:  "min_price",
		SubPath:     "min_price",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum price the user wants to pay (using a GET request)",
	}

	return &UnitAsset{
		// TODO: These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		SEK_price: 0.0,  // Example electricity price in SEK per kWh
		Min_price: 0.0,  // Minimum price allowed
		Max_price: 1.0,  // Maximum price allowed
		Min_temp:  15.0, // Minimum temperature
		Max_temp:  20.0, // Maximum temprature allowed

		// Don't forget to map the provided services from above!
		ServicesMap: components.Services{
			setMax_temp.SubPath:  &setMax_temp,
			setMin_temp.SubPath:  &setMin_temp,
			setMax_price.SubPath: &setMax_price,
			setMin_price.SubPath: &setMin_price,
			setSEK_price.SubPath: &setSEK_price,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////

// newUnitAsset creates a new and proper instance of UnitAsset, using settings and
// values loaded from an existing configuration file.
// This function returns an UA instance that is ready to be published and used,
// aswell as a function that can perform any cleanup when the system is shutting down.
func newUnitAsset(uac UnitAsset, sys *components.System, servs []components.Service) (components.UnitAsset, func()) {

	sProtocol := components.SProtocols(sys.Husk.ProtoPort)

	// the Cervice that is to be consumed by zigbee, there fore the name with the C
	t := &components.Cervice{
		Name:   "setpoint",
		Protos: sProtocol,
		Url:    make([]string, 0),
	}
	ua := &UnitAsset{
		// Filling in public fields using the given data
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),
		CervicesMap: components.Cervices{
			t.Name: t,
		},
	}
	/*
		var ref components.Service
		for _, s := range servs {
			if s.Definition == "setpoint" {
				ref = s
			}
		}

		ua.CervicesMap["temperature"].Details = components.MergeDetails(ua.Details, ref.Details)
		ua.CervicesMap["rotation"].Details = components.MergeDetails(ua.Details, map[string][]string{"Unit": {"Percent"}, "Forms": {"SignalA_v1a"}})
	*/
	// Optionally start background tasks here! Example:
	go func() {
		log.Println("Starting up " + ua.Name)
	}()

	// Returns the loaded unit asset and an function to handle optional cleanup at shutdown
	return ua, func() {
		log.Println("Cleaning up " + ua.Name)
	}
}

// getSEK_price is used for reading the current hourly electric price
func (ua *UnitAsset) getSEK_price() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.SEK_price
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

// setSEK_price updates the current electric price with the new current electric hourly price
func (ua *UnitAsset) setSEK_price(f forms.SignalA_v1a) {
	ua.SEK_price = f.Value
	log.Printf("new electric price: %.1f", f.Value)
}

/////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////

// getMin_price is used for reading the current value of Min_price
func (ua *UnitAsset) getMin_price() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Min_price
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

// setMin_price updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMin_price(f forms.SignalA_v1a) {
	ua.Min_price = f.Value
	log.Printf("new minimum price: %.1f", f.Value)
}

// getMax_price is used for reading the current value of Max_price
func (ua *UnitAsset) getMax_price() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Max_price
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

// setMax_price updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMax_price(f forms.SignalA_v1a) {
	ua.Max_price = f.Value
	log.Printf("new maximum price: %.1f", f.Value)
}

// getMin_temp is used for reading the current minimum temerature value
func (ua *UnitAsset) getMin_temp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Min_temp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

// setMin_temp updates the current minimum temperature set by the user with a new value
func (ua *UnitAsset) setMin_temp(f forms.SignalA_v1a) {
	ua.Min_temp = f.Value
	log.Printf("new minimum temperature: %.1f", f.Value)
}

// getMax_temp is used for reading the current value of Min_price
func (ua *UnitAsset) getMax_temp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Max_temp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

// setMax_temp updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMax_temp(f forms.SignalA_v1a) {
	ua.Max_temp = f.Value
	log.Printf("new maximum temperature: %.1f", f.Value)
}

//TODO: This fuction is used for checking the electric price ones every x hours and so on
//TODO: Needs to be modified to match our needs, not using processFeedbacklopp
//TODO: So mayby the period is every hour, call the api to receive the current price ( could be every 24 hours)
//TODO: This function is may be better in the COMFORTSTAT MAIN
/*
// feedbackLoop is THE control loop (IPR of the system)
func (ua *UnitAsset) API_feedbackLoop(ctx context.Context) {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(ua.Period * time.Second)
	defer ticker.Stop()

	// start the control loop
	for {
		select {
		case <-ticker.C:
			ua.processFeedbackLoop() // either modifiy processFeedback loop or write a new one
		case <-ctx.Done():
			return
		}
	}
}

// feedbackLoop is THE control loop (IPR of the system)
func (ua *UnitAsset) feedbackLoop(ctx context.Context) {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(ua.Period * time.Second)
	defer ticker.Stop()

	// start the control loop
	for {
		select {
		case <-ticker.C:
			ua.processFeedbackLoop() // either modifiy processFeedback loop or write a new one
		case <-ctx.Done():
			return
		}
	}
}
*/
