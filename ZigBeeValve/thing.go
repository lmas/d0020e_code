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
	"strings"
	"time"

	"github.com/gorilla/websocket"
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
	Model    string            `json:"model"`
	Uniqueid string            `json:"uniqueid"`
	Period   time.Duration     `json:"period"`
	Setpt    float64           `json:"setpoint"`
	Slaves   map[string]string `json:"slaves"`
	Apikey   string            `json:"APIkey"`
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
	// This service will only be supported by Smart Thermostats and Smart Power plugs.
	setPointService := components.Service{
		Definition:  "setpoint",
		SubPath:     "setpoint",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current thermal setpoint (GET) or sets it (PUT)",
	}

	// This service will only be supported by Smart Power plugs (will be noted as sensors of type ZHAConsumption)
	consumptionService := components.Service{
		Definition:  "consumption",
		SubPath:     "consumption",
		Details:     map[string][]string{"Unit": {"Wh"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current consumption of the device in Wh (GET)",
	}

	// This service will only be supported by Smart Power plugs (will be noted as sensors of type ZHAPower)
	currentService := components.Service{
		Definition:  "current",
		SubPath:     "current",
		Details:     map[string][]string{"Unit": {"mA"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current going through the device in mA (GET)",
	}

	// This service will only be supported by Smart Power plugs (will be noted as sensors of type ZHAPower)
	powerService := components.Service{
		Definition:  "power",
		SubPath:     "power",
		Details:     map[string][]string{"Unit": {"W"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current consumption of the device in W (GET)",
	}

	// This service will only be supported by Smart Power plugs (Will be noted as sensors of type ZHAPower)
	voltageService := components.Service{
		Definition:  "voltage",
		SubPath:     "voltage",
		Details:     map[string][]string{"Unit": {"V"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current voltage of the device in V (GET)",
	}

	// This service will only be supported by Smart Power plugs (Will be noted as sensors of type ZHAPower)
	toggleService := components.Service{
		Definition:  "toggle",
		SubPath:     "toggle",
		Details:     map[string][]string{"Unit": {"Binary"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current state of the device (GET), or sets it (PUT) [0 = off, 1 = on]",
	}

	// var uat components.UnitAsset // this is an interface, which we then initialize
	uat := &UnitAsset{
		Name:     "SmartSwitch1",
		Details:  map[string][]string{"Location": {"Kitchen"}},
		Model:    "ZHASwitch",
		Uniqueid: "14:ef:14:10:00:6f:d0:d7-11-1201",
		Period:   10,
		Setpt:    20,
		// Only switches needs to manually add controlled power plug and light uniqueids, power plugs get their sensors added automatically
		Slaves: map[string]string{"Plug1": "14:ef:14:10:00:6f:d0:d7-XX-XXXX", "Plug2": "24:ef:24:20:00:6f:d0:d2-XX-XXXX"},
		Apikey: "1234",
		ServicesMap: components.Services{
			setPointService.SubPath:    &setPointService,
			consumptionService.SubPath: &consumptionService,
			currentService.SubPath:     &currentService,
			powerService.SubPath:       &powerService,
			voltageService.SubPath:     &voltageService,
			toggleService.SubPath:      &toggleService,
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
		Period:      uac.Period,
		Setpt:       uac.Setpt,
		Slaves:      uac.Slaves,
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
		if websocketport == "startup" {
			err := ua.getWebsocketPort()
			if err != nil {
				log.Println("Error occured during startup, while calling getWebsocketPort():", err)
				// TODO: Check if we need to kill program if this doesn't pass?
			}

		}
		switch ua.Model {
		case "ZHAThermostat":
			err := ua.sendSetPoint()
			if err != nil {
				log.Println("Error occured during startup, while calling sendSetPoint():", err)
			}
		case "Smart plug":
			// Find all sensors belonging to the smart plug and put them in the slaves array with
			// their type as the key
			err := ua.getSensors()
			if err != nil {
				log.Println("Error occured during startup, while calling getSensors():", err)
			}
			// Not all smart plugs should be handled by the feedbackloop, some should be handled by a switch
			if ua.Period > 0 {
				// start the unit assets feedbackloop, this fetches the temperature from ds18b20 and and toggles
				// between on/off depending on temperature in the room and a set temperature in the unitasset
				go ua.feedbackLoop(ua.Owner.Ctx)
			}
		case "ZHASwitch":
			// Starts listening to the websocket to find buttonevents (button presses) and then
			// turns its controlled devices (slaves) on/off
			go ua.initWebsocketClient(ua.Owner.Ctx)
		default:
			return
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

// Function to find the gateway and save its ip and port (assuming there's only one) and return the error if one occurs
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
	// If the returned list is empty, return a missing gateway error
	if len(gw) < 1 {
		return errMissingGateway
	}
	// Save the gateway
	s := fmt.Sprintf(`%s:%d`, gw[0].Internalipaddress, gw[0].Internalport)
	gateway = s
	return
}

//-------------------------------------Thing's resource methods

// Function to get sensors connected to a smart plug and place them in the "slaves" array
type sensorJSON struct {
	UniqueID string `json:"uniqueid"`
	Type     string `json:"type"`
}

func (ua *UnitAsset) getSensors() (err error) {
	// Create and send a get request to get all sensors connected to deConz gateway
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors"
	req, err := createGetRequest(apiURL)
	if err != nil {
		return err
	}
	data, err := sendGetRequest(req)
	if err != nil {
		return err
	}
	// Unmarshal data from get request into an easy to use JSON format
	var sensors map[string]sensorJSON
	err = json.Unmarshal(data, &sensors)
	if err != nil {
		return err
	}
	// Take only the part of the mac address that is present in both the smart plug and the sensors
	macAddr := ua.Uniqueid[0:23]
	for _, sensor := range sensors {
		uniqueid := sensor.UniqueID
		check := strings.Contains(uniqueid, macAddr)
		if check == true {
			if sensor.Type == "ZHAConsumption" {
				ua.Slaves["ZHAConsumption"] = sensor.UniqueID
			}
			if sensor.Type == "ZHAPower" {
				ua.Slaves["ZHAPower"] = sensor.UniqueID
			}
		}
	}
	return
}

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

// Function to send a new setpoint ot a device that has the "heatsetpoint" in its config (smart plug or smart thermostat)
func (ua *UnitAsset) sendSetPoint() (err error) {
	// API call to set desired temp in smart thermostat, PUT call should be sent to  URL/api/apikey/sensors/sensor_id/config
	// --- Send setpoint to specific unit ---
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Uniqueid + "/config"
	// Create http friendly payload
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, ua.Setpt*100) // Create payload
	req, err := createPutRequest(s, apiURL)
	if err != nil {
		return
	}
	return sendPutRequest(req)
}

// Functions and structs to get and set current state of a smart plug
type plugJSON struct {
	State struct {
		On bool `json:"on"`
	} `json:"state"`
}

func (ua *UnitAsset) getState() (f forms.SignalA_v1a, err error) {
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/lights/" + ua.Uniqueid
	req, err := createGetRequest(apiURL)
	if err != nil {
		return f, err
	}
	data, err := sendGetRequest(req)
	var plug plugJSON
	err = json.Unmarshal(data, &plug)
	if err != nil {
		return f, err
	}
	// Return a form containing current state in binary form (1 = on, 0 = off)
	if plug.State.On == true {
		f := getForm(1, "Binary")
		return f, nil
	}
	if plug.State.On == false {
		f := getForm(0, "Binary")
		return f, nil
	}
	return
}

func (ua *UnitAsset) setState(f forms.SignalA_v1a) (err error) {
	if f.Value == 0 {
		return ua.toggleState(false)
	}
	if f.Value == 1 {
		return ua.toggleState(true)
	}
	return
}

// Function to toggle the state of a specific device (power plug or light) on/off and return an error if it occurs
func (ua *UnitAsset) toggleState(state bool) (err error) {
	// API call to toggle light/smart plug on/off, PUT call should be sent to URL/api/apikey/lights/[light_id or plug_id]/state
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/lights/" + ua.Uniqueid + "/state"
	// Create http friendly payload
	s := fmt.Sprintf(`{"on":%t}`, state) // Create payload
	req, err := createPutRequest(s, apiURL)
	if err != nil {
		return
	}
	return sendPutRequest(req)
}

// Functions to create put or get reques and return the *http.request and/or error if one occurs
func createPutRequest(data string, apiURL string) (req *http.Request, err error) {
	body := bytes.NewReader([]byte(data))                    // Put data into buffer
	req, err = http.NewRequest(http.MethodPut, apiURL, body) // Put request is made
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json") // Make sure it's JSON
	return req, nil
}

func createGetRequest(apiURL string) (req *http.Request, err error) {
	req, err = http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json") // Make sure it's JSON
	return req, nil
}

// A function to send a put request that returns the error if one occurs
func sendPutRequest(req *http.Request) (err error) {
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

// A function to send get requests and return the data received in the response body as a []byte and/or error if it happens
func sendGetRequest(req *http.Request) (data []byte, err error) {
	resp, err := http.DefaultClient.Do(req) // Perform the http request
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body) // Read the response body, and check for errors/bad statuscodes
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		return nil, errStatusCode
	}
	return data, nil
}

// Creates a form that fills the fields of forms.SignalA_v1a with values from arguments and current time
func getForm(value float64, unit string) (f forms.SignalA_v1a) {
	f.NewForm()
	f.Value = value
	f.Unit = fmt.Sprint(unit)
	f.Timestamp = time.Now()
	return f
}

// ------------------------------------------------------------------------------------------------------------
// IMPORTANT: lumi.plug.maeu01 HAS BEEN KNOWN TO GIVE BAD READINGS, BASICALLY STOP RESPONDING OR RESPOND WITH 0
// 	      They also don't appear for a long time after re-pairing devices to deConz
// ------------------------------------------------------------------------------------------------------------

// Struct and method to get and return a form containing current consumption (in Wh)
type consumptionJSON struct {
	State struct {
		Consumption uint64 `json:"consumption"`
	} `json:"state"`
	Name     string `json:"name"`
	UniqueID string `json:"uniqueid"`
	Type     string `json:"type"`
}

func (ua *UnitAsset) getConsumption() (f forms.SignalA_v1a, err error) {
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Slaves["ZHAConsumption"]
	// Create a get request
	req, err := createGetRequest(apiURL)
	if err != nil {
		return f, err
	}
	// Perform get request to sensor, expecting a body containing json data to be returned
	body, err := sendGetRequest(req)
	if err != nil {
		return f, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()
	// Unmarshal the body into usable json data
	var data consumptionJSON
	err = json.Unmarshal(body, &data)
	if err != nil {
		return f, err
	}
	// Set form value to sensors value
	value := float64(data.State.Consumption)
	f = getForm(value, "Wh")
	return f, nil
}

// Struct and method to get and return a form containing current power (in W)
type powerJSON struct {
	State struct {
		Power int16 `json:"power"`
	} `json:"state"`
	Name     string `json:"name"`
	UniqueID string `json:"uniqueid"`
	Type     string `json:"type"`
}

func (ua *UnitAsset) getPower() (f forms.SignalA_v1a, err error) {
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Slaves["ZHAPower"]
	// Create a get request
	req, err := createGetRequest(apiURL)
	if err != nil {
		return f, err
	}
	// Perform get request to sensor, expecting a body containing json data to be returned
	body, err := sendGetRequest(req)
	if err != nil {
		return f, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()
	// Unmarshal the body into usable json data
	var data powerJSON
	err = json.Unmarshal(body, &data)
	if err != nil {
		return f, err
	}
	// Set form value to sensors value
	value := float64(data.State.Power)
	f = getForm(value, "W")
	return f, nil
}

// Struct and method to get and return a form containing current (in mA)
type currentJSON struct {
	State struct {
		Current uint16 `json:"current"`
	} `json:"state"`
	Name     string `json:"name"`
	UniqueID string `json:"uniqueid"`
	Type     string `json:"type"`
}

func (ua *UnitAsset) getCurrent() (f forms.SignalA_v1a, err error) {
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Slaves["ZHAPower"]
	// Create a get request
	req, err := createGetRequest(apiURL)
	if err != nil {
		return f, err
	}
	// Perform get request to sensor, expecting a body containing json data to be returned
	body, err := sendGetRequest(req)
	if err != nil {
		return f, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()
	// Unmarshal the body into usable json data
	var data currentJSON
	err = json.Unmarshal(body, &data)
	if err != nil {
		return f, err
	}
	// Set form value to sensors value
	value := float64(data.State.Current)
	f = getForm(value, "mA")
	return f, nil
}

// Struct and method to get and return a form containing current voltage (in V)
type voltageJSON struct {
	State struct {
		Voltage uint16 `json:"voltage"`
	} `json:"state"`
	Name     string `json:"name"`
	UniqueID string `json:"uniqueid"`
	Type     string `json:"type"`
}

func (ua *UnitAsset) getVoltage() (f forms.SignalA_v1a, err error) {
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Slaves["ZHAPower"]
	// Create a get request
	req, err := createGetRequest(apiURL)
	if err != nil {
		return f, err
	}
	// Perform get request to power plug sensor, expecting a body containing json data to be returned
	body, err := sendGetRequest(req)
	if err != nil {
		return f, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()
	// Unmarshal the body into usable json data
	var data voltageJSON
	err = json.Unmarshal(body, &data)
	if err != nil {
		return f, err
	}
	// Set form value to sensors value
	value := float64(data.State.Voltage)
	f = getForm(value, "V")
	return f, nil
}

// --- HOW TO CONNECT AND LISTEN TO A WEBSOCKET ---
// Port 443, can be found by curl -v "http://localhost:8080/api/[apikey]/config", and getting the "websocketport".
// https://dresden-elektronik.github.io/deconz-rest-doc/endpoints/websocket/
// https://stackoverflow.com/questions/32745716/i-need-to-connect-to-an-existing-websocket-server-using-go-lang
// https://github.com/gorilla/websocket

// In order for websocketport to run at startup i gave it something to check against and update
var websocketport = "startup"

type eventJSON struct {
	State struct {
		Buttonevent int `json:"buttonevent"`
	} `json:"state"`
	UniqueID string `json:"uniqueid"`
}

// This function sends a request for the config of the gateway, and saves the websocket port
// If an error occurs it will return that error
func (ua *UnitAsset) getWebsocketPort() (err error) {
	// --- Get config ---
	apiURL := fmt.Sprintf("http://%s/api/%s/config", gateway, ua.Apikey)
	// Create a new request (Get)
	req, err := http.NewRequest(http.MethodGet, apiURL, nil) // Put request is made
	if err != nil {
		return err
	}
	// Make sure it's JSON
	req.Header.Set("Content-Type", "application/json")
	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Read the response body, and check for errors/bad statuscodes
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return errStatusCode
	}
	// How to access maps inside of maps below!
	// https://stackoverflow.com/questions/28806951/accessing-nested-map-of-type-mapstringinterface-in-golang
	var configMap map[string]interface{}
	err = json.Unmarshal([]byte(resBody), &configMap)
	if err != nil {
		return err
	}
	websocketport = fmt.Sprint(configMap["websocketport"])
	return
}

// STRETCH GOAL: Below can also be done with groups, could look into makeing groups for each device, and then delete them on shutdown
//		 doing it with groups would make it so we don't have to keep track of a global variable and i think if unlucky only change
//		 one light or smart plug depending on reachability

// This function loops through the "slaves" of a unit asset, and sets them to either true (for on) and false (off), returning an error if it occurs
func (ua *UnitAsset) toggleSlaves(currentState bool) (err error) {
	for i := range ua.Slaves {
		log.Printf("Toggling: %s to %v", ua.Slaves[i], currentState)
		// API call to toggle smart plug or lights on/off, PUT call should be sent to URL/api/apikey/[sensors or lights]/sensor_id/config
		apiURL := fmt.Sprintf("http://%s/api/%s/lights/%v/state", gateway, ua.Apikey, ua.Slaves[i])
		// Create http friendly payload
		s := fmt.Sprintf(`{"on":%t}`, currentState) // Create payload
		req, err := createPutRequest(s, apiURL)
		if err != nil {
			return err
		}
		err = sendPutRequest(req)
	}
	return err
}

// Function starts listening to a websocket, every message received through websocket is read, and checked if it's what we're looking for
// The uniqueid (UniqueID in systemconfig.json file) from the connected switch is used to filter out messages
func (ua *UnitAsset) initWebsocketClient(ctx context.Context) error {
	gateway = "192.168.10.122:8080" // For testing purposes
	dialer := websocket.Dialer{}
	wsURL := fmt.Sprintf("ws://192.168.10.122:%s", websocketport) // For testing purposes
	//wsURL := fmt.Sprintf("ws://localhost:%s", websocketport)
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Error occured while dialing:", err)
	}
	log.Println("Connected to websocket")
	defer conn.Close()
	currentState := false
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Read the message
			_, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error occured while reading message:", err)
				return err
			}
			// Put it inot a message variable of type eventJSON with "buttonevent" easily accessible
			var message eventJSON
			err = json.Unmarshal(p, &message)
			if err != nil {
				log.Println("Error unmarshalling message:", err)
				return err
			}
			// Depending on what buttonevent occured, either turn the slaves on, or off
			if message.UniqueID == ua.Uniqueid && (message.State.Buttonevent == 1002 || message.State.Buttonevent == 2002) {
				bEvent := message.State.Buttonevent
				if currentState == true {
					currentState = false
				} else {
					currentState = true
				}
				if bEvent == 1002 {
					// Turn on the smart plugs (lights)
					err = ua.toggleSlaves(currentState)
					if err != nil {
						return err
					}
				}
				if bEvent == 2002 {
					// Turn on the philips hue light
					err = ua.toggleSlaves(currentState)
					if err != nil {
						return err
					}
					// TODO: Find out how "long presses" works and if it can be used through websocket
				}
			}
		}
	}
}
