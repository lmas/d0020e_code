/* In order to follow the structure of the other systems made before this one, most functions and structs are copied and slightly edited from:
 * https://github.com/sdoque/systems/blob/main/thermostat/thermostat.go */

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/usecases"
)

func main() {
	// prepare for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background()) // create a context that can be cancelled
	defer cancel()                                          // make sure all paths cancel the context to avoid context leak

	// instantiate the System
	sys := components.NewSystem("ZigBeeValve", ctx)

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: " is a controller for smart thermostats connected with a RaspBee II",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/sdoque/systems/tree/master/ZigBeeValve",
	}

	// instantiate a template unit asset
	assetTemplate := initTemplate()
	assetName := assetTemplate.GetName()
	sys.UAssets[assetName] = &assetTemplate

	// Configure the system
	rawResources, servsTemp, err := usecases.Configure(&sys)
	if err != nil {
		log.Fatalf("Configuration error: %v\n", err)
	}
	sys.UAssets = make(map[string]*components.UnitAsset) // clear the unit asset map (from the template)
	for _, raw := range rawResources {
		var uac UnitAsset
		if err := json.Unmarshal(raw, &uac); err != nil {
			log.Fatalf("Resource configuration error: %+v\n", err)
		}
		ua, cleanup := newResource(uac, &sys, servsTemp)
		defer cleanup()
		sys.UAssets[ua.GetName()] = &ua
	}

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.RequestCertificate(&sys)

	// Register the (system) and its services
	usecases.RegisterServices(&sys)

	// start the http handler and server
	go usecases.SetoutServers(&sys)

	// wait for shutdown signal, and gracefully close properly goroutines with context
	<-sys.Sigs // wait for a SIGINT (Ctrl+C) signal
	fmt.Println("\nshuting down system", sys.Name)
	cancel()                    // cancel the context, signaling the goroutines to stop
	time.Sleep(2 * time.Second) // allow the go routines to be executed, which might take more time than the main routine to end
}

// Serving handles the resources services. NOTE: it exepcts those names from the request URL path
func (t *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	case "setpoint":
		t.setpt(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configurration file]", http.StatusBadRequest)
	}
}

func (rsc *UnitAsset) setpt(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		setPointForm := rsc.getSetPoint()
		usecases.HTTPProcessGetRequest(w, r, &setPointForm)
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting desired temp ", err)
		}
		rsc.setSetPoint(sig)
		// API call to set desired temp in smart thermostat
		// PUT call should be sent to  URL/api/apikey/sensors/2/config
		apiURL := "http://" + rsc.gateway + "/api/" + rsc.Apikey + "/sensors/2/config"
		// URL is hardcoded for now, could use /sensors to get all sensors, and then go through all of 'em with a loop
		// looking for a specific keywords, like kitchen and save the id of all thermostats or w/e in the kitchen
		// in an array to then change them all one at a time with a loop
		/*
			// GET request, unmarshal to get the array of JSONS, containing all sensors/thermostats,
			// save the array in an array called sensors
			var changeArray []int
			for int j = 0; i in range sensors;j++ {
				if i.location == keyword {
					// save its id to a new array
					changeArray[j] = i.id

				}
			}
			// Then loop through the array containing ids, and set all their configs (heatsetpoint) to desired temp
		*/

		// Create http friendly payload
		s := fmt.Sprintf(`{"heatsetpoint":%f}`, rsc.Setpt*100) // Create payload
		data := []byte(s)                                      // Turned into byte array
		body := bytes.NewBuffer(data)                          // and put into buffer

		req, err := http.NewRequest(http.MethodPut, apiURL, body) // Put request is made
		if err != nil {
			log.Println("Error making new HTTP PUT request, error:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json") // Make sure it's JSON

		client := &http.Client{}    // Make a client
		resp, err := client.Do(req) // Perform the put request
		defer resp.Body.Close()
		if err != nil {
			log.Println("Error sending HTTP PUT request, error:", err)
		}

	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
