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

//------------------------------------ Used when discovering the gateway

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
	Model   string        `json:"model"`
	Period  time.Duration `json:"period"`
	Setpt   float64       `json:"setpoint"`
	gateway string
	Apikey  string `json:"APIkey"`
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
		Name:    "Template",
		Details: map[string][]string{"Location": {"Kitchen"}},
		Model:   "",
		Period:  10,
		Setpt:   20,
		gateway: "",
		Apikey:  "",
		ServicesMap: components.Services{
			setPointService.SubPath: &setPointService,
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
		Model:       uac.Model,
		Period:      uac.Period,
		Setpt:       uac.Setpt,
		gateway:     uac.gateway,
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
		if ua.Model == "SmartThermostat" {
			ua.sendSetPoint()
		} else if ua.Model == "SmartPlug" {
			// start the unit asset(s)
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
		ua.toggleState(true)
	} else {
		ua.toggleState(false)
	}
	//log.Println("Feedback loop done.")

}

var gateway string

func findGateway() {
	// https://pkg.go.dev/net/http#Get
	// GET https://phoscon.de/discover	// to find gateways, array of JSONs is returned in http body, we'll only have one so take index 0
	// GET the gateway through phoscons built in discover tool, the get will return a response, and in its body an array with JSON elements
	// ours is index 0 since there's no other RaspBee/ZigBee gateways on the network
	res, err := http.Get("https://phoscon.de/discover")
	if err != nil {
		log.Println("Couldn't get gateway, error:", err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and\n", res.StatusCode)
		return
	}
	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		log.Println("Something went wrong while reading the body during discovery, error:", err)
		return
	}
	var gw []discoverJSON           // Create a list to hold the gateway json
	err = json.Unmarshal(body, &gw) // "unpack" body from []byte to []discoverJSON, save errors
	if err != nil {
		log.Println("Error during Unmarshal, error:", err)
		return
	}
	if len(gw) < 1 {
		log.Println("No gateway was found")
		return
	}
	// Save the gateway to our unitasset
	s := fmt.Sprintf(`%s:%d`, gw[0].Internalipaddress, gw[0].Internalport)
	gateway = s
	//log.Println("Gateway found:", s)
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
	log.Println("*---------------------*")
	log.Printf("New set point: %.1f\n", f.Value)
	log.Println("*---------------------*")
}

func (ua *UnitAsset) sendSetPoint() {
	// API call to set desired temp in smart thermostat, PUT call should be sent to  URL/api/apikey/sensors/sensor_id/config
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Name + "/config"

	// Create http friendly payload
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, ua.Setpt*100) // Create payload
	data := []byte(s)                                     // Turned into byte array
	sendRequest(data, apiURL)
}

func (ua *UnitAsset) toggleState(state bool) {
	// API call to set desired temp in smart thermostat, PUT call should be sent to  URL/api/apikey/sensors/sensor_id/config
	apiURL := "http://" + gateway + "/api/" + ua.Apikey + "/lights/" + ua.Name + "/state"

	// Create http friendly payload
	s := fmt.Sprintf(`{"on":%t}`, state) // Create payload
	data := []byte(s)                    // Turned into byte array
	sendRequest(data, apiURL)
}

func sendRequest(data []byte, apiURL string) {
	body := bytes.NewBuffer(data) // Put data into buffer

	req, err := http.NewRequest(http.MethodPut, apiURL, body) // Put request is made
	if err != nil {
		log.Println("Error making new HTTP PUT request, error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json") // Make sure it's JSON

	client := &http.Client{}    // Make a client
	resp, err := client.Do(req) // Perform the put request
	if err != nil {
		log.Println("Error sending HTTP PUT request, error:", err)
		return
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body) // Read the payload into body variable
	if err != nil {
		log.Println("Something went wrong while reading the body during discovery, error:", err)
		return
	}
	if resp.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, string(b))
		return
	}
}
