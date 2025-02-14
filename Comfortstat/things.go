package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

type GlobalPriceData struct {
	SEKPrice  float64 `json:"SEK_per_kWh"`
	EURPrice  float64 `json:"EUR_per_kWh"`
	EXR       float64 `json:"EXR"`
	TimeStart string  `json:"time_start"`
	TimeEnd   string  `json:"time_end"`
}

// initiate "globalPrice" with default values
var globalPrice = GlobalPriceData{
	SEKPrice:  0,
	EURPrice:  0,
	EXR:       0,
	TimeStart: "0",
	TimeEnd:   "0",
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
	Period time.Duration `json:"SamplingPeriod"`
	//
	DesiredTemp    float64 `json:"DesiredTemp"`
	oldDesiredTemp float64 // keep this field private!
	SEKPrice       float64 `json:"SEK_per_kWh"`
	MinPrice       float64 `json:"MinPrice"`
	MaxPrice       float64 `json:"MaxPrice"`
	MinTemp        float64 `json:"MinTemp"`
	MaxTemp        float64 `json:"MaxTemp"`
	UserTemp       float64 `json:"UserTemp"`
	Region         float64 `json:"Region"` // the user can choose from what region the SEKPrice is taken from
}

// SE1: Norra Sverige/Luleå   		(value = 1)
// SE2: Norra MellanSverige/Sundsvall 	(value = 2)
// SE3: Södra MellanSverige/Stockholm   (value = 3)
// SE4: Södra Sverige/Kalmar 		(value = 4)

func initAPI() {
	go priceFeedbackLoop()
}

const apiFetchPeriod int = 3600

var GlobalRegion float64 = 1

// defines the URL for the electricity price and starts the getAPIPriceData function once every hour
func priceFeedbackLoop() {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(time.Duration(apiFetchPeriod) * time.Second)
	defer ticker.Stop()

	url := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE%d.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(), int(GlobalRegion))
	// start the control loop
	for {
		err := getAPIPriceData(url)
		if err != nil {
			return
		}

		select {
		case <-ticker.C:
			// blocks the execution until the ticker fires
		}
	}
}

// This function checks if the user has changed price-region and then calls the getAPIPriceData function which gets the right pricedata
func switchRegion() {
	urlSE1 := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE1.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	urlSE2 := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE2.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	urlSE3 := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE3.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	urlSE4 := fmt.Sprintf(`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE4.json`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())

	// SE1: Norra Sverige/Luleå   		(value = 1)
	if GlobalRegion == 1 {
		err := getAPIPriceData(urlSE1)
		if err != nil {
			return
		}
	}
	// SE2: Norra MellanSverige/Sundsvall 	(value = 2)
	if GlobalRegion == 2 {
		err := getAPIPriceData(urlSE2)
		if err != nil {
			return
		}
	}
	// SE3: Södra MellanSverige/Stockholm   (value = 3)
	if GlobalRegion == 3 {
		err := getAPIPriceData(urlSE3)
		if err != nil {
			return
		}
	}
	// SE4: Södra Sverige/Kalmar 		(value = 4)
	if GlobalRegion == 4 {
		err := getAPIPriceData(urlSE4)
		if err != nil {
			return
		}
	}
}

var errStatuscode error = fmt.Errorf("bad status code")
var data []GlobalPriceData // Create a list to hold the data json

// This function fetches the current electricity price from "https://www.elprisetjustnu.se/elpris-api", then process it and updates globalPrice
func getAPIPriceData(apiURL string) error {
	//Validate the URL//
	parsedURL, err := url.Parse(apiURL) // ensures the string is a valid URL, .schema and .Host checks prevent empty or altered URL
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("The URL is invalid")
	}
	// end of validating the URL//
	res, err := http.Get(parsedURL.String())
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &data) // "unpack" body from []byte to []GlobalPriceData, save errors

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return errStatuscode
	}
	if err != nil {
		return err
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
// (see https://github.com/sdoque/mbaigo/blob/main/components/service.go for documentation)
func initTemplate() components.UnitAsset {

	setSEKPrice := components.Service{
		Definition:  "SEKPrice",
		SubPath:     "SEKPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current electric hourly price (using a GET request)",
	}
	setMaxTemp := components.Service{
		Definition:  "MaxTemperature",
		SubPath:     "MaxTemperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the maximum temp the user wants (using a GET request)",
	}
	setMinTemp := components.Service{
		Definition:  "MinTemperature",
		SubPath:     "MinTemperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum temp the user could tolerate (using a GET request)",
	}
	setMaxPrice := components.Service{
		Definition:  "MaxPrice",
		SubPath:     "MaxPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the maximum price the user wants to pay (using a GET request)",
	}
	setMinPrice := components.Service{
		Definition:  "MinPrice",
		SubPath:     "MinPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum price the user wants to pay (using a GET request)",
	}
	setDesiredTemp := components.Service{
		Definition:  "DesiredTemp",
		SubPath:     "DesiredTemp",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the desired temperature the system calculates based on user inputs (using a GET request)",
	}
	setUserTemp := components.Service{
		Definition:  "UserTemp",
		SubPath:     "UserTemp",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the temperature the user wants regardless of prices (using a GET request)",
	}
	setRegion := components.Service{
		Definition:  "Region",
		SubPath:     "Region",
		Details:     map[string][]string{"Forms": {"SignalA_v1a"}},
		Description: "provides the temperature the user wants regardless of prices (using a GET request)",
	}

	return &UnitAsset{
		//These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name:        "Set_Values",
		Details:     map[string][]string{"Location": {"Kitchen"}},
		SEKPrice:    1.5,  // Example electricity price in SEK per kWh
		MinPrice:    1.0,  // Minimum price allowed
		MaxPrice:    2.0,  // Maximum price allowed
		MinTemp:     20.0, // Minimum temperature
		MaxTemp:     25.0, // Maximum temperature allowed
		DesiredTemp: 0,    // Desired temp calculated by system
		Period:      15,
		UserTemp:    0,
		Region:      1,

		// maps the provided services from above
		ServicesMap: components.Services{
			setMaxTemp.SubPath:     &setMaxTemp,
			setMinTemp.SubPath:     &setMinTemp,
			setMaxPrice.SubPath:    &setMaxPrice,
			setMinPrice.SubPath:    &setMinPrice,
			setSEKPrice.SubPath:    &setSEKPrice,
			setDesiredTemp.SubPath: &setDesiredTemp,
			setUserTemp.SubPath:    &setUserTemp,
			setRegion.SubPath:      &setRegion,
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
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),
		SEKPrice:    uac.SEKPrice,
		MinPrice:    uac.MinPrice,
		MaxPrice:    uac.MaxPrice,
		MinTemp:     uac.MinTemp,
		MaxTemp:     uac.MaxTemp,
		DesiredTemp: uac.DesiredTemp,
		Period:      uac.Period,
		UserTemp:    uac.UserTemp,
		Region:      uac.Region,
		CervicesMap: components.Cervices{
			t.Name: t,
		},
	}

	var ref components.Service
	for _, s := range servs {
		if s.Definition == "DesiredTemp" {
			ref = s
		}
	}

	ua.CervicesMap["setpoint"].Details = components.MergeDetails(ua.Details, ref.Details)

	// Returns the loaded unit asset and an function to handle
	return ua, func() {
		// start the unit asset(s)
		go ua.feedbackLoop(sys.Ctx)
	}
}

// getSEKPrice is used for reading the current hourly electric price
func (ua *UnitAsset) getSEKPrice() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.SEKPrice
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

//Get and set- methods for MIN/MAX price/temp and desierdTemp

// getMinPrice is used for reading the current value of MinPrice
func (ua *UnitAsset) getMinPrice() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.MinPrice
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

// setMinPrice updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMinPrice(f forms.SignalA_v1a) {
	ua.MinPrice = f.Value
}

// getMaxPrice is used for reading the current value of MaxPrice
func (ua *UnitAsset) getMaxPrice() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.MaxPrice
	f.Unit = "SEK"
	f.Timestamp = time.Now()
	return f
}

// setMaxPrice updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMaxPrice(f forms.SignalA_v1a) {
	ua.MaxPrice = f.Value
}

// getMinTemp is used for reading the current minimum temperature value
func (ua *UnitAsset) getMinTemp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.MinTemp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

// setMinTemp updates the current minimum temperature set by the user with a new value
func (ua *UnitAsset) setMinTemp(f forms.SignalA_v1a) {
	ua.MinTemp = f.Value
}

// getMaxTemp is used for reading the current value of MinPrice
func (ua *UnitAsset) getMaxTemp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.MaxTemp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

// setMaxTemp updates the current minimum price set by the user with a new value
func (ua *UnitAsset) setMaxTemp(f forms.SignalA_v1a) {
	ua.MaxTemp = f.Value
}

func (ua *UnitAsset) getDesiredTemp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.DesiredTemp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

func (ua *UnitAsset) setDesiredTemp(f forms.SignalA_v1a) {
	ua.DesiredTemp = f.Value
}

func (ua *UnitAsset) setUserTemp(f forms.SignalA_v1a) {
	ua.UserTemp = f.Value
	if ua.UserTemp != 0 {
		ua.sendUserTemp()
	}
}

func (ua *UnitAsset) getUserTemp() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.UserTemp
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}
func (ua *UnitAsset) setRegion(f forms.SignalA_v1a) {
	ua.Region = f.Value
	GlobalRegion = ua.Region
	switchRegion()
}

func (ua *UnitAsset) getRegion() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Region
	f.Timestamp = time.Now()
	return f
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
			ua.processFeedbackLoop() // either modify processFeedback loop or write a new one
		case <-ctx.Done():
			return
		}
	}
}

// this function adjust and sends a new desierd temperature to the zigbee system
// get the current best temperature
func (ua *UnitAsset) processFeedbackLoop() {
	ua.Region = GlobalRegion
	// extracts the electricity price depending on the current time and updates globalPrice
	now := fmt.Sprintf(`%d-%02d-%02dT%02d:00:00+01:00`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(), time.Now().Local().Hour())
	log.Println("TIME:", now)
	for _, i := range data {
		if i.TimeStart == now {
			globalPrice.SEKPrice = i.SEKPrice
		}
	}

	ua.SEKPrice = globalPrice.SEKPrice

	//ua.DesiredTemp = ua.calculateDesiredTemp(miT, maT, miP, maP, ua.getSEKPrice().Value)
	ua.DesiredTemp = ua.calculateDesiredTemp()
	// Only send temperature update when we have a new value.
	if (ua.DesiredTemp == ua.oldDesiredTemp) || (ua.UserTemp != 0) {
		if ua.UserTemp != 0 {
			ua.oldDesiredTemp = ua.UserTemp
			return
		}
		return
	}
	// Keep track of previous value
	ua.oldDesiredTemp = ua.DesiredTemp

	// prepare the form to send
	var of forms.SignalA_v1a
	of.NewForm()
	of.Value = ua.DesiredTemp
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

// Calculates the new most optimal temperature (desierdTemp) based on the price/temprature intervals
// and the current electricity price
func (ua *UnitAsset) calculateDesiredTemp() float64 {

	if ua.SEKPrice <= ua.MinPrice {
		return ua.MaxTemp
	}
	if ua.SEKPrice >= ua.MaxPrice {
		return ua.MinTemp
	}

	k := (ua.MinTemp - ua.MaxTemp) / (ua.MaxPrice - ua.MinPrice)
	m := ua.MaxTemp - (k * ua.MinPrice)
	DesiredTemp := k*(ua.SEKPrice) + m

	return DesiredTemp
}

func (ua *UnitAsset) sendUserTemp() {
	var of forms.SignalA_v1a
	of.NewForm()
	of.Value = ua.UserTemp
	of.Unit = ua.CervicesMap["setpoint"].Details["Unit"][0]
	of.Timestamp = time.Now()

	op, err := usecases.Pack(&of, "application/json")
	if err != nil {
		return
	}
	err = usecases.SetState(ua.CervicesMap["setpoint"], ua.Owner, op)
	if err != nil {
		log.Printf("cannot update zigbee setpoint: %s\n", err)
		return
	}
}
