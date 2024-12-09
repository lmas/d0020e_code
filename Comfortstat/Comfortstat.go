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
	sys := components.NewSystem("thermostat", ctx)

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: " is a controller for a consumed servo motor position based on a consumed temperature",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/sdoque/systems/tree/master/thermostat",
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
	case "min_temperature":
		t.set_temp(w, r)
	case "max_temperature":
		t.set_temp(w, r)
	case "max_price":
		t.set_price(w, r)
	case "min_price":
		t.set_price(w, r)
	case "SEK_price":
		t.set_SEKprice(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configurration file]", http.StatusBadRequest)
	}
}

/*
	func (rsc *UnitAsset) set_SEKprice(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			sig, err := usecases.HTTPProcessSetRequest(w, r)
			if err != nil {
				log.Println("Error with the setting request of the position ", err)
			}
			rsc.set_SEKprice(sig)
		default:
			http.Error(w, "Method is not supported.", http.StatusNotFound)
		}
	}
*/
func (rsc *UnitAsset) set_temp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.set_minMaxtemp(sig)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

// LOOK AT: I guess that we probable only need to if there is a PUT from user?
// LOOK AT: so not the GET!
// For PUT - the "HTTPProcessSetRequest(w, r)" is called to prosses the data given from the user and if no error, call set_minMaxprice with the value
// wich updates the value in thge struct
func (rsc *UnitAsset) set_price(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.set_minMaxprice(sig)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)

	}
}