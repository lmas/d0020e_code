/* In order to follow the structure of the other systems made before this one, most functions and structs are copied and slightly edited from:
 * https://github.com/sdoque/systems/blob/main/thermostat/thing.go */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
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
	Setpt   float64 `json:"setpoint"`
	gateway string  `json:"-"`
	Apikey  string  `json:"APIkey"`
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
		Name:    "2",
		Details: map[string][]string{"Location": {"Kitchen"}},
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

	// intantiate the unit asset
	ua := &UnitAsset{
		Name:        uac.Name,
		Owner:       sys,
		Details:     uac.Details,
		ServicesMap: components.CloneServices(servs),
		Setpt:       uac.Setpt,
		gateway:     uac.gateway,
		Apikey:      uac.Apikey,
		CervicesMap: components.Cervices{},
	}

	findGateway(ua)
	ua.sendSetPoint()
	return ua, func() {
		log.Println("Shutting down zigbeevalve ", ua.Name)
	}
}

func findGateway(ua *UnitAsset) {
	// https://pkg.go.dev/net/http#Get
	// GET https://phoscon.de/discover	// to find gateways, array of JSONs is returned in http body, we'll only have one so take index 0
	// GET the gateway through phoscons built in discover tool, the get will return a response, and in its body an array with JSON elements
	// ours is index 0 since there's no other RaspBee/ZigBee gateways on the network
	res, err := http.Get("https://phoscon.de/discover")
	if err != nil {
		log.Println("Couldn't get gateway, error:", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d and\n", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body) // Read the payload into body variable
	if err != nil {
		log.Println("Something went wrong while reading the body during discovery, error:", err)
	}
	var gw []discoverJSON           // Create a list to hold the gateway json
	err = json.Unmarshal(body, &gw) // "unpack" body from []byte to []discoverJSON, save errors
	if err != nil {
		log.Println("Error during Unmarshal, error:", err)
	}
	if len(gw) < 1 {
		log.Println("No gateway was found")
		return
	}
	// Save the gateway to our unitasset
	s := fmt.Sprintf(`%s:%d`, gw[0].Internalipaddress, gw[0].Internalport)
	ua.gateway = s
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
	apiURL := "http://" + ua.gateway + "/api/" + ua.Apikey + "/sensors/" + ua.Name + "/config"

	// Create http friendly payload
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, ua.Setpt*100) // Create payload
	data := []byte(s)                                     // Turned into byte array
	body := bytes.NewBuffer(data)                         // and put into buffer

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
}
