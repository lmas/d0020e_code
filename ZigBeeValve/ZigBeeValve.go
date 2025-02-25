/* In order to follow the structure of the other systems made before this one, most functions and structs are copied and slightly edited from:
 * https://github.com/sdoque/systems/blob/main/thermostat/thermostat.go */

package main

import (
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
	sys := components.NewSystem("ZigBeeHandler", ctx)

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: " is a controller for smart devices connected with a RaspBee II",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8870, "coap": 0},
		InfoLink:    "https://github.com/sdoque/systems/tree/master",
	}

	// instantiate a template unit asset
	assetTemplate := initTemplate()
	assetName := assetTemplate.GetName()
	sys.UAssets[assetName] = &assetTemplate

	// Find zigbee gateway and store it in a global variable for reuse
	err := findGateway()
	if err != nil {
		log.Fatal("Error getting gateway, shutting down: ", err)
	}

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
		ua, startup := newResource(uac, &sys, servsTemp)
		if err := startup(); err != nil {
			log.Fatalf("Error during startup: %s\n", err)
		}
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
	case "consumption":
		t.consumption(w, r)
	case "current":
		t.current(w, r)
	case "power":
		t.power(w, r)
	case "voltage":
		t.voltage(w, r)
	case "state":
		t.state(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configuration file]", http.StatusBadRequest)
	}
}

// TODO: Add webhandler for power plug controller (sun up/down) and/or schedule later on.
// STRETCH GOAL: Instead of looking for specific models types, add a list of supported devices that we can check against

// Function used by webhandler to either get or set the setpoint of a specific device
func (rsc *UnitAsset) setpt(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Make sure only devices with setpoints actually support the http get method
		if rsc.Model == "ZHAThermostat" || rsc.Model == "Smart plug" {
			setPointForm := rsc.getSetPoint()
			usecases.HTTPProcessGetRequest(w, r, &setPointForm)
			return
		}
		http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
		return
	case "PUT":
		// Make sure only devices with setpoints actually support the http put method
		if rsc.Model == "ZHAThermostat" || rsc.Model == "Smart plug" {
			sig, err := usecases.HTTPProcessSetRequest(w, r)
			if err != nil {
				http.Error(w, "Request incorrectly formated", http.StatusBadRequest)
				return
			}
			rsc.setSetPoint(sig)
			return
		}
		http.Error(w, "This device doesn't support that method.", http.StatusInternalServerError)
		return
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

// Function used by the webhandler to get the consumption of a device
func (rsc *UnitAsset) consumption(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Make sure only devices with consumption sensors actually support the http get method
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		consumptionForm, err := rsc.getConsumption()
		if err != nil {
			http.Error(w, "Failed getting data, or data not present", http.StatusInternalServerError)
			return
		}
		usecases.HTTPProcessGetRequest(w, r, &consumptionForm)
	default:
		http.Error(w, "Method is not supported", http.StatusNotFound)
	}
}

// Function used by the webhandler to get the power of a device
func (rsc *UnitAsset) power(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Make sure only devices with power sensors actually support the http get method
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		powerForm, err := rsc.getPower()
		if err != nil {
			http.Error(w, "Failed getting data, or data not present", http.StatusInternalServerError)
			return
		}
		usecases.HTTPProcessGetRequest(w, r, &powerForm)
	default:
		http.Error(w, "Method is not supported", http.StatusNotFound)
	}
}

// Function used by the webhandler to get the current of a device
func (rsc *UnitAsset) current(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Make sure only devices with current sensors actually support the http get method
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		currentForm, err := rsc.getCurrent()
		if err != nil {
			http.Error(w, "Failed getting data, or data not present", http.StatusInternalServerError)
			return
		}
		usecases.HTTPProcessGetRequest(w, r, &currentForm)
	default:
		http.Error(w, "Method is not supported", http.StatusNotFound)
	}
}

// Function used by the webhandler to get the voltage of a device
func (rsc *UnitAsset) voltage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Make sure only devices with voltage sensors actually support the http get method
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		voltageForm, err := rsc.getVoltage()
		if err != nil {
			http.Error(w, "Failed getting data, or data not present", http.StatusInternalServerError)
			return
		}
		usecases.HTTPProcessGetRequest(w, r, &voltageForm)
	default:
		http.Error(w, "Method is not supported", http.StatusNotFound)
	}
}

func (rsc *UnitAsset) state(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		stateForm, err := rsc.getState()
		if err != nil {
			http.Error(w, "Failed getting data, or data not present", http.StatusInternalServerError)
			return
		}
		usecases.HTTPProcessGetRequest(w, r, &stateForm)
	case "PUT":
		if rsc.Model != "Smart plug" {
			http.Error(w, "That device doesn't support that method.", http.StatusInternalServerError)
			return
		}
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "Request incorrectly formated", http.StatusBadRequest)
			return
		}
		err = rsc.setState(sig)
		if err != nil {
			http.Error(w, "Something went wrong when setting state", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "Method is not supported", http.StatusNotFound)
	}
}
