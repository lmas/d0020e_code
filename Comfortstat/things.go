package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

type GlobalPriceData struct {
	SEK_price  float64 `json:"SEK_per_kWh"`
	EUR_price  float64 `json:"EUR_per_kWh"`
	EXR        float64 `json:"EXR"`
	Time_start string  `json:"time_start"`
	Time_end   string  `json:"time_end"`
}

var globalPrice = GlobalPriceData{
	SEK_price:  0,
	EUR_price:  0,
	EXR:        0,
	Time_start: "0",
	Time_end:   "0",
}

// A UnitAsset models an interface or API for a smaller part of a whole system, for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
type UnitAsset struct {
	Name        string              `json:"name"`    // Must be a unique name, ie. a sensor ID
	Owner       *components.System  `json:"-"`       // The parent system this UA is part of
	Details     map[string][]string `json:"details"` // Metadata or details about this UA
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`
	//
	Period time.Duration `json:"samplingPeriod"`
	//
	Desired_temp     float64 `json:"desired_temp"`
	old_desired_temp float64 // keep this field private!
	SEK_price        float64 `json:"SEK_per_kWh"`
	Min_price        float64 `json:"min_price"`
	Max_price        float64 `json:"max_price"`
	Min_temp         float64 `json:"min_temp"`
	Max_temp         float64 `json:"max_temp"`
}

func initAPI() {
	go priceFeedbackLoop()
}

func priceFeedbackLoop() {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(time.Duration(apiFetchPeriod) * time.Second)
	defer ticker.Stop()

	url := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE1.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	// start the control loop
	for {
		getAPIPriceData(url)
		select {
		case <-ticker.C:
			// Block the loop until the next period
		}
	}
}

var err_statuscode error = fmt.Errorf("bad status code")

func getAPIPriceData(url string) error {

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		return err
	}

	var data []GlobalPriceData        // Create a list to hold the gateway json
	err = json.Unmarshal(body, &data) // "unpack" body from []byte to []GlobalPriceData, save errors
	res.Body.Close()                  // defer res.Body.Close()

	if res.StatusCode > 299 {
		return err_statuscode
	}
	if err != nil {
		return err
	}

	/////////
	now := fmt.Sprintf(`%d-%02d-%02dT%02d:00:00+01:00`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(), time.Now().Local().Hour())
	for _, i := range data {
		if i.Time_start == now {
			globalPrice.SEK_price = i.SEK_price
		}
	}
	return nil
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
		Definition:  "max_temperature",
		SubPath:     "max_temperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
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
	setDesired_temp := components.Service{
		Definition:  "desired_temp",
		SubPath:     "desired_temp",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the desired temperature the system calculates based on user inputs (using a GET request)",
	}

	return &UnitAsset{
		// TODO: These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name:         "Set Values",
		Details:      map[string][]string{"Location": {"Kitchen"}},
		SEK_price:    1.5,  // Example electricity price in SEK per kWh
		Min_price:    1.0,  // Minimum price allowed
		Max_price:    2.0,  // Maximum price allowed
		Min_temp:     20.0, // Minimum temperature
		Max_temp:     25.0, // Maximum temprature allowed
		Desired_temp: 0,    // Desired temp calculated by system
		Period:       15,

		// maps the provided services from above
		ServicesMap: components.Services{
			setMax_temp.SubPath:     &setMax_temp,
			setMin_temp.SubPath:     &setMin_temp,
			setMax_price.SubPath:    &setMax_price,
			setMin_price.SubPath:    &setMin_price,
			setSEK_price.SubPath:    &setSEK_price,
			setDesired_temp.SubPath: &setDesired_temp,
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

	// the Cervice that is to be consumed by zigbee, therefore the name with the C

	t := &components.Cervice{
		Name:   "setpoint",
		Protos: sProtocol,
		Url:    make([]string, 0),
	}

	ua := &UnitAsset{
		// Filling in public fields using the given data
		Name:         uac.Name,
		Owner:        sys,
		Details:      uac.Details,
		ServicesMap:  components.CloneServices(servs),
		SEK_price:    uac.SEK_price,
		Min_price:    uac.Min_price,
		Max_price:    uac.Max_price,
		Min_temp:     uac.Min_temp,
		Max_temp:     uac.Max_temp,
		Desired_temp: uac.Desired_temp,
		Period:       uac.Period,
		CervicesMap: components.Cervices{
			t.Name: t,
		},
	}

	var ref components.Service
	for _, s := range servs {
		if s.Definition == "desired_temp" {
			ref = s
		}
	}

	ua.CervicesMap["setpoint"].Details = components.MergeDetails(ua.Details, ref.Details)

	// start the unit asset(s)
	go ua.feedbackLoop(sys.Ctx)
	go ua.API_feedbackLoop(sys.Ctx)

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

/*
// setSEK_price updates the current electric price with the new current electric hourly price
func (ua *UnitAsset) setSEK_price(f forms.SignalA_v1a) {
	ua.SEK_price = f.Value
	//log.Printf("new electric price: %.1f", f.Value)
}
*/
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
}

func (ua *UnitAsset) getDesired_temp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Desired_temp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

func (ua *UnitAsset) setDesired_temp(f forms.SignalA_v1a) {
	ua.Desired_temp = f.Value
	log.Printf("new desired temperature: %.1f", f.Value)
}

//TODO: This fuction is used for checking the electric price ones every x hours and so on
//TODO: Needs to be modified to match our needs, not using processFeedbacklopp
//TODO: So mayby the period is every hour, call the api to receive the current price ( could be every 24 hours)
//TODO: This function is may be better in the COMFORTSTAT MAIN

// It's _strongly_ encouraged to not send requests to the API for more than once per hour.
// Making this period a private constant prevents a user from changing this value
// in the config file.
const apiFetchPeriod int = 3600

// feedbackLoop is THE control loop (IPR of the system)
func (ua *UnitAsset) API_feedbackLoop(ctx context.Context) {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(time.Duration(apiFetchPeriod) * time.Second)
	defer ticker.Stop()

	// start the control loop
	for {
		retrieveAPI_price(ua)
		select {
		case <-ticker.C:
			// Block the loop until the next period
		case <-ctx.Done():
			return
		}
	}
}

func retrieveAPI_price(ua *UnitAsset) {
	if globalPrice.SEK_price == 0 {
		time.Sleep(1 * time.Second)
	}
	ua.SEK_price = globalPrice.SEK_price
	// Don't send temperature updates if the difference is too low
	// (this could potentially save on battery!)
	new_temp := ua.calculateDesiredTemp()
	if math.Abs(ua.Desired_temp-new_temp) < 0.5 {
		return
	}
	ua.Desired_temp = new_temp
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

//

func (ua *UnitAsset) processFeedbackLoop() {
	// get the current best temperature

	//ua.Desired_temp = ua.calculateDesiredTemp(miT, maT, miP, maP, ua.getSEK_price().Value)
	ua.Desired_temp = ua.calculateDesiredTemp()
	// Only send temperature update when we have a new value.
	if ua.Desired_temp == ua.old_desired_temp {
		return
	}
	// Keep track of previous value
	ua.old_desired_temp = ua.Desired_temp

	// prepare the form to send
	var of forms.SignalA_v1a
	of.NewForm()
	of.Value = ua.Desired_temp
	of.Unit = ua.CervicesMap["setpoint"].Details["Unit"][0]
	of.Timestamp = time.Now()

	// pack the new valve state form
	// Pack() converting the data in "of" into JSON format
	op, err := usecases.Pack(&of, "application/json")
	if err != nil {
		return
	}
	// send the new valve state request
	err = usecases.SetState(ua.CervicesMap["setpoint"], ua.Owner, op)
	if err != nil {
		log.Printf("cannot update zigbee setpoint: %s\n", err)
		return
	}
}

// Calculates the new most optimal temprature (desierdTemp) based on the price/temprature intervalls
// and the current electricity price
func (ua *UnitAsset) calculateDesiredTemp() float64 {

	if ua.SEK_price <= ua.Min_price {
		return ua.Max_temp
	}
	if ua.SEK_price >= ua.Max_price {
		return ua.Min_temp
	}

	k := (ua.Min_temp - ua.Max_temp) / (ua.Max_price - ua.Min_price)
	m := ua.Max_temp - (k * ua.Min_price)
	desired_temp := k*(ua.SEK_price) + m

	return desired_temp
}
