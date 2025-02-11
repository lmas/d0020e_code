/* In order to follow the structure of the other systems made before this one, most functions and structs are copied and slightly edited from:
 * https://github.com/sdoque/systems/blob/main/thermostat/thing.go */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

// ------------------------------------ Used when discovering the gateway
type discoverJSON struct {
	Id                string `json:"id"`
	Internalipaddress string `json:"internalipaddress"`
	Macaddress        string `json:"macaddress"`
	Internalport      int    `json:"internalport"`
	Name              string `json:"name"`
	Publicipaddress   string `json:"publicipaddress"`
}

//-------------------------------------Define the unit asset

// UnitAsset type models the unit asset (interface) of the system
type UnitAsset struct {
	Name        string              `json:"name"`
	Owner       *components.System  `json:"-"`
	Details     map[string][]string `json:"details"`
	ServicesMap components.Services `json:"-"`
	CervicesMap components.Cervices `json:"-"`
	//
	Model       string `json:"model"`
	Uniqueid    string `json:"uniqueid"`
	deviceIndex string
	Period      time.Duration `json:"period"`
	Setpt       float64       `json:"setpoint"`
	Apikey      string        `json:"APIkey"`
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

	// var uat components.UnitAsset // this is an interface, which we then initialize
	uat := &UnitAsset{
		Name:        "SmartThermostat1",
		Details:     map[string][]string{"Location": {"Kitchen"}},
		Model:       "ZHAThermostat",
		Uniqueid:    "14:ef:14:10:00:6f:d0:d7-11-1201",
		deviceIndex: "",
		Period:      10,
		Setpt:       20,
		Apikey:      "1234",
		ServicesMap: components.Services{
			setPointService.SubPath: &setPointService,
		},
	}
	return uat
}

//-------------------------------------Instatiate the unit assets based on configuration

// newResource creates the resource with its pointers and channels based on the configuration using the tConfig structs
// This is a startup function that's used to initiate the unit assets declared in the systemconfig.json, the function
// that is returned is later used to send a setpoint/start a goroutine depending on model of the unitasset
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
		Model:       uac.Model,
		Uniqueid:    uac.Uniqueid,
		deviceIndex: uac.deviceIndex,
		Period:      uac.Period,
		Setpt:       uac.Setpt,
		Apikey:      uac.Apikey,
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

		if ua.Model == "ZHAThermostat" {
			// Get correct index in list returned by api/sensors to make sure we always change correct device
			err := ua.getConnectedUnits("sensors")
			if err != nil {
				log.Println("Error occured during startup, while calling getConnectedUnits:", err)
			}
			err = ua.sendSetPoint()
			if err != nil {
				log.Println("Error occured during startup, while calling sendSetPoint():", err)
				// TODO: Turn off system if this startup() fails?
			}
		} else if ua.Model == "Smart plug" {
			// Get correct index in list returned by api/lights to make sure we always change correct device
			err := ua.getConnectedUnits("lights")
			if err != nil {
				log.Println("Error occured during startup, while calling getConnectedUnits:", err)
			}
			// start the unit assets feedbackloop, this fetches the temperature from ds18b20 and and toggles
			// between on/off depending on temperature in the room and a set temperature in the unitasset
			go ua.feedbackLoop(ua.Owner.Ctx)
		}
	}
}

func (ua *UnitAsset) feedbackLoop(ctx context.Context) {
	// Initialize a ticker for periodic execution
	ticker := time.NewTicker(ua.Period * time.Second)
	defer ticker.Stop()
	// start the control loop
	for {
		select {
		case <-ticker.C:
			ua.processFeedbackLoop()
		case <-ctx.Done():
			return
		}
	}
}

func (ua *UnitAsset) processFeedbackLoop() {
	// get the current temperature
	tf, err := usecases.GetState(ua.CervicesMap["temperature"], ua.Owner)
	if err != nil {
		log.Printf("\n unable to obtain a temperature reading error: %s\n", err)
		return
	}
	// Perform a type assertion to convert the returned Form to SignalA_v1a
	tup, ok := tf.(*forms.SignalA_v1a)
	if !ok {
		log.Println("problem unpacking the temperature signal form")
		return
	}
	// TODO: Check diff instead of a hard over/under value? meaning it'll only turn on/off if diff is over 0.5 degrees
	if tup.Value < ua.Setpt {
		err = ua.toggleState(true)
		if err != nil {
			log.Println("Error occured while toggling state to true: ", err)
		}
	} else {
		err = ua.toggleState(false)
		if err != nil {
			log.Println("Error occured while toggling state to false: ", err)
		}
	}
}

var gateway string

const discoveryURL string = "https://phoscon.de/discover"

var errStatusCode error = fmt.Errorf("bad status code")
var errMissingGateway error = fmt.Errorf("missing gateway")
var errMissingUniqueID error = fmt.Errorf("uniqueid not found")

func findGateway() (err error) {
	// https://pkg.go.dev/net/http#Get
	// GET https://phoscon.de/discover	// to find gateways, array of JSONs is returned in http body, we'll only have one so take index 0
	// GET the gateway through phoscons built in discover tool, the get will return a response, and in its body an array with JSON elements
	// ours is index 0 since there's no other RaspBee/ZigBee gateways on the network
	res, err := http.Get(discoveryURL)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		//log.Printf("Response failed with status code: %d and\n", res.StatusCode)
		return errStatusCode
	}
	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		return
	}
	var gw []discoverJSON           // Create a list to hold the gateway json
	err = json.Unmarshal(body, &gw) // "unpack" body from []byte to []discoverJSON, save errors
	if err != nil {
		return
	}

	if len(gw) < 1 {
		return errMissingGateway
	}
	// Save the gateway
	s := fmt.Sprintf(`%s:%d`, gw[0].Internalipaddress, gw[0].Internalport)
	gateway = s
	return
}

//-------------------------------------Thing's resource methods

// getSetPoint fills out a signal form with the current thermal setpoint
func (ua *UnitAsset) getSetPoint() (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = ua.Setpt
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	return f
}

// setSetPoint updates the thermal setpoint
func (ua *UnitAsset) setSetPoint(f forms.SignalA_v1a) {
	ua.Setpt = f.Value
}

func (ua *UnitAsset) sendSetPoint() (err error) {
	// API call to set desired temp in smart thermostat, PUT call should be sent to  URL/api/apikey/sensors/sensor_id/config

	// --- Send setpoint to specific unit ---
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.deviceIndex + "/config"
	// Create http friendly payload
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, ua.Setpt*100) // Create payload
	req, err := createRequest(s, apiURL)
	if err != nil {
		return
	}
	return sendRequest(req)
}

func (ua *UnitAsset) toggleState(state bool) (err error) {
	// API call turn smart plug on/off, PUT call should be sent to  URL/api/apikey/lights/sensor_id/config
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/lights/" + ua.deviceIndex + "/state"
	// Create http friendly payload
	s := fmt.Sprintf(`{"on":%t}`, state) // Create payload
	req, err := createRequest(s, apiURL)
	if err != nil {
		return
	}
	return sendRequest(req)
}

func (ua *UnitAsset) getConnectedUnits(unitType string) (err error) {
	// Get all devices
	apiURL := fmt.Sprintf("http://%s/api/%s/%s", gateway, ua.Apikey, unitType)
	// Create a new request (Get)
	// Put data into buffer
	req, err := http.NewRequest(http.MethodGet, apiURL, nil) // Put request is made
	req.Header.Set("Content-Type", "application/json")       // Make sure it's JSON
	// Send the request
	resp, err := http.DefaultClient.Do(req) // Perform the http request
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body) // Read the response body, and check for errors/bad statuscodes
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		return errStatusCode
	}

	// How to access maps inside of maps below!
	// https://stackoverflow.com/questions/28806951/accessing-nested-map-of-type-mapstringinterface-in-golang
	var deviceMap map[string]interface{}
	err = json.Unmarshal([]byte(resBody), &deviceMap)
	if err != nil {
		return
	}
	for i := range deviceMap {
		if deviceMap[i].(map[string]interface{})["uniqueid"] == ua.Uniqueid {
			ua.deviceIndex = i
			return
		}
	}
	return errMissingUniqueID
}

func createRequest(data string, apiURL string) (req *http.Request, err error) {
	body := bytes.NewReader([]byte(data))                    // Put data into buffer
	req, err = http.NewRequest(http.MethodPut, apiURL, body) // Put request is made
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json") // Make sure it's JSON
	return req, err
}

func sendRequest(req *http.Request) (err error) {
	resp, err := http.DefaultClient.Do(req) // Perform the http request
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body) // Read the response body, and check for errors/bad statuscodes
	if err != nil {
		return
	}
	if resp.StatusCode > 299 {
		return errStatusCode
	}
	return
}

// Create a group, add all lights/power plugs from e.g. kitchen to said group
// Create rule, on button.event toggle power plugs
