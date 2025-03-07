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

type SunData struct {
	Date        string  `json:"date"`
	Sunrise     string  `json:"sunrise"`
	Sunset      string  `json:"sunset"`
	First_light string  `json:"first_light"`
	Last_light  string  `json:"last_light"`
	Dawn        string  `json:"dawn"`
	Dusk        string  `json:"dusk"`
	Solar_noon  string  `json:"solar_noon"`
	Golden_hour string  `json:"golden_hour"`
	Day_length  string  `json:"day_length"`
	Timezone    string  `json:"timezone"`
	Utc_offset  float64 `json:"utc_offset"`
}

type Data struct {
	Results SunData `json:"results"`
	Status  string  `json:"status"`
}

// A unitAsset models an interface or API for a smaller part of a whole system, for example a single temperature sensor.
// This type must implement the go interface of "components.UnitAsset"
type UnitAsset struct {
	Name        string              `json:"name"`
	Owner       *components.System  `json:"-"`
	Details     map[string][]string `json:"details"`
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`

	Period time.Duration `json:"samplingPeriod"`

	ButtonStatus float64 `json:"ButtonStatus"`
	Latitude     float64 `json:"Latitude"`
	oldLatitude  float64
	Longitude    float64 `json:"Longitude"`
	oldLongitude float64
	data         Data
	connError    float64
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

// Ensure UnitAsset implements components.UnitAsset (this check is done at during the compilation)
var _ components.UnitAsset = (*UnitAsset)(nil)

////////////////////////////////////////////////////////////////////////////////

// initTemplate initializes a new UA and prefils it with some default values.
// The returned instance is used for generating the configuration file, whenever it's missing.
// (see https://github.com/sdoque/mbaigo/blob/main/components/service.go for documentation)
func initTemplate() components.UnitAsset {
	setLatitude := components.Service{
		Definition:  "Latitude",
		SubPath:     "Latitude",
		Details:     map[string][]string{"Unit": {"Degrees"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current set latitude (using a GET request)",
	}
	setLongitude := components.Service{
		Definition:  "Longitude",
		SubPath:     "Longitude",
		Details:     map[string][]string{"Unit": {"Degrees"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current set longitude (using a GET request)",
	}
	setButtonStatus := components.Service{
		Definition:  "ButtonStatus",
		SubPath:     "ButtonStatus",
		Details:     map[string][]string{"Unit": {"bool"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the status of a button (using a GET request)",
	}

	return &UnitAsset{
		// These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name:         "Button",
		Details:      map[string][]string{"Location": {"Kitchen"}},
		Latitude:     65.584816, // Latitude for the button
		Longitude:    22.156704, // Longitude for the button
		ButtonStatus: 0.5,       // Status for the button (on/off) NOTE: This status is neither on or off as default, this is up for the system to decide.
		Period:       15,
		data:         Data{SunData{}, ""},

		// Maps the provided services from above
		ServicesMap: components.Services{
			setLatitude.SubPath:     &setLatitude,
			setLongitude.SubPath:    &setLongitude,
			setButtonStatus.SubPath: &setButtonStatus,
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

	// the Cervice that is to be consumed by the ZigBee, therefore the name with the C
	t := &components.Cervice{
		Name:   "state",
		Protos: sProtocol,
		Url:    make([]string, 0),
	}

	ua := &UnitAsset{
		// Filling in public fields using the given data
		Name:         uac.Name,
		Owner:        sys,
		Details:      uac.Details,
		ServicesMap:  components.CloneServices(servs),
		Latitude:     uac.Latitude,
		Longitude:    uac.Longitude,
		ButtonStatus: uac.ButtonStatus,
		Period:       uac.Period,
		data:         uac.data,
		CervicesMap: components.Cervices{
			t.Name: t,
		},
	}

	var ref components.Service
	for _, s := range servs {
		if s.Definition == "ButtonStatus" {
			ref = s
		}
	}

	ua.CervicesMap["state"].Details = components.MergeDetails(ua.Details, ref.Details)

	// Returns the loaded unit asset and a function to handle
	return ua, func() {
		// Start the unit asset(s)
		go ua.feedbackLoop(sys.Ctx)
	}
}

// getLatitude is used for reading the current latitude
func (ua *UnitAsset) getLatitude() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Latitude
	f.Unit = "Degrees"
	f.Timestamp = time.Now()
	return f
}

// setLatitude is used for updating the current latitude
func (ua *UnitAsset) setLatitude(f forms.SignalA_v1a) {
	ua.oldLatitude = ua.Latitude
	ua.Latitude = f.Value
}

// getLongitude is used for reading the current longitude
func (ua *UnitAsset) getLongitude() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Longitude
	f.Unit = "Degrees"
	f.Timestamp = time.Now()
	return f
}

// setLongitude is used for updating the current longitude
func (ua *UnitAsset) setLongitude(f forms.SignalA_v1a) {
	ua.oldLongitude = ua.Longitude
	ua.Longitude = f.Value
}

// getButtonStatus is used for reading the current button status
func (ua *UnitAsset) getButtonStatus() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.ButtonStatus
	f.Unit = "bool"
	f.Timestamp = time.Now()
	return f
}

// setButtonStatus is used for updating the current button status
func (ua *UnitAsset) setButtonStatus(f forms.SignalA_v1a) {
	ua.ButtonStatus = f.Value
}

// feedbackLoop is THE control loop (IPR of the system)
func (ua *UnitAsset) feedbackLoop(ctx context.Context) {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(ua.Period * time.Second)
	defer ticker.Stop()

	// Start the control loop
	for {
		select {
		case <-ticker.C:
			ua.processFeedbackLoop()
		case <-ctx.Done():
			return
		}
	}
}

// This function sends a new button status to the ZigBee system if needed
func (ua *UnitAsset) processFeedbackLoop() {
	date := time.Now().Format("2006-01-02") // Gets the current date in the defined format.
	apiURL := fmt.Sprintf(`http://api.sunrisesunset.io/json?lat=%06f&lng=%06f&timezone=CET&date=%d-%02d-%02d&time_format=24`, ua.Latitude, ua.Longitude, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())

	if !((ua.data.Results.Date == date) && ((ua.oldLatitude == ua.Latitude) && (ua.oldLongitude == ua.Longitude))) { // If there is a new day or latitude or longitude is changed new data is downloaded.
		log.Printf("Sun API has not been called today for this region, downloading sun data...")
		err := ua.getAPIData(apiURL)
		if err != nil {
			log.Printf("Cannot get sun API data: %s\n", err)
			return
		}
	}
	ua.oldLongitude = ua.Longitude
	ua.oldLatitude = ua.Latitude
	layout := "15:04:05"
	sunrise, _ := time.Parse(layout, ua.data.Results.Sunrise)                   // Saves the sunrise in the layout format.
	sunset, _ := time.Parse(layout, ua.data.Results.Sunset)                     // Saves the sunset in the layout format.
	currentTime, _ := time.Parse(layout, time.Now().Local().Format("15:04:05")) // Saves the current time in the layout format.
	if currentTime.After(sunrise) && !(currentTime.After(sunset)) {             // This checks if the time is between sunrise or sunset, if it is the switch is supposed to turn off.
		if ua.ButtonStatus == 0 && ua.connError == 0 { // If the button is already off there is no need to send a state again.
			log.Printf("The button is already off")
			return
		}
		ua.ButtonStatus = 0
		err := ua.sendStatus()
		if err != nil {
			ua.connError = 1
			return
		} else {
			ua.connError = 0
		}

	} else { // If the time is not between sunrise and sunset the button is supposed to be on.
		if ua.ButtonStatus == 1 && ua.connError == 0 { // If the button is already on there is no need to send a state again.
			log.Printf("The button is already on")
			return
		}
		ua.ButtonStatus = 1
		err := ua.sendStatus()
		if err != nil {
			ua.connError = 1
			return
		} else {
			ua.connError = 0
		}
	}
}

func (ua *UnitAsset) sendStatus() error {
	// Prepare the form to send
	var of forms.SignalA_v1a
	of.NewForm()
	of.Value = ua.ButtonStatus
	of.Unit = ua.CervicesMap["state"].Details["Unit"][0]
	of.Timestamp = time.Now()
	// Pack the new state form
	// Pack() converting the data in "of" into JSON format
	op, err := usecases.Pack(&of, "application/json")
	if err != nil {
		return err
	}
	// Send the new request
	err = usecases.SetState(ua.CervicesMap["state"], ua.Owner, op)
	if err != nil {
		log.Printf("Cannot update ZigBee state: %s\n", err)
		return err
	}
	return nil
}

var errStatuscode error = fmt.Errorf("bad status code")

func (ua *UnitAsset) getAPIData(apiURL string) error {
	//apiURL := fmt.Sprintf(`http://api.sunrisesunset.io/json?lat=%06f&lng=%06f&timezone=CET&date=%d-%02d-%02d&time_format=24`, ua.Latitude, ua.Longitude, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	parsedURL, err := url.Parse(apiURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("the url is invalid")
	}
	// End of validating the URL //
	res, err := http.Get(parsedURL.String())
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &ua.data)

	defer res.Body.Close()

	if res.StatusCode > 299 {
		return errStatuscode
	}
	if err != nil {
		return err
	}
	return nil
}
