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
	// Prepare for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background()) // Create a context that can be cancelled
	defer cancel()                                          // Make sure all paths cancel the context to avoid context leak

	// Instantiate the System
	sys := components.NewSystem("SunButton", ctx)

	// Instantiate the Capsule
	sys.Husk = &components.Husk{
		Description: "Is a controller for a consumed button based on a consumed time of day. Powered by SunriseSunset.io",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8770, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/SunButton",
	}

	// Instantiate a template unit asset
	assetTemplate := initTemplate()
	assetName := assetTemplate.GetName()
	sys.UAssets[assetName] = &assetTemplate

	// Configure the system
	rawResources, servsTemp, err := usecases.Configure(&sys)
	if err != nil {
		log.Fatalf("Configuration error: %v\n", err)
	}
	sys.UAssets = make(map[string]*components.UnitAsset) // Clear the unit asset map (from the template)
	for _, raw := range rawResources {
		var uac UnitAsset
		if err := json.Unmarshal(raw, &uac); err != nil {
			log.Fatalf("Resource configuration error: %+v\n", err)
		}
		ua, startup := newUnitAsset(uac, &sys, servsTemp)
		startup()
		sys.UAssets[ua.GetName()] = &ua
	}

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.RequestCertificate(&sys)

	// Register the (system) and its services
	usecases.RegisterServices(&sys)

	// Start the http handler and server
	go usecases.SetoutServers(&sys)

	// Wait for shutdown signal and gracefully close properly goroutines with context
	<-sys.Sigs // Wait for a SIGINT (Crtl+C) signal
	fmt.Println("\nShutting down system", sys.Name)
	cancel()                    // Cancel the context, signaling the goroutines to stop
	time.Sleep(2 * time.Second) // Allow the go routines to be executed, which might take more time then the main routine to end
}

// Serving handles the resource services. NOTE: It expects those names from the request URL path
func (t *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	case "ButtonStatus":
		t.httpSetButton(w, r)
	case "Latitude":
		t.httpSetLatitude(w, r)
	case "Longitude":
		t.httpSetLongitude(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configuration file]", http.StatusBadRequest)
	}
}

// All these functions below handles HTTP "PUT" or "GET" requests to modify or retrieve the latitude and longitude and the state of the button
// For the PUT case - the "HTTPProcessSetRequest(w, r)" is called to prosses the data given from the user and if no error,
// call the set functions in thing.go with the value witch updates the value in the struct
func (rsc *UnitAsset) httpSetButton(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setButtonStatus(sig)
	case "GET":
		signalErr := rsc.getButtonStatus()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

func (rsc *UnitAsset) httpSetLatitude(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setLatitude(sig)
	case "GET":
		signalErr := rsc.getLatitude()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

func (rsc *UnitAsset) httpSetLongitude(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setLongitude(sig)
	case "GET":
		signalErr := rsc.getLongitude()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
