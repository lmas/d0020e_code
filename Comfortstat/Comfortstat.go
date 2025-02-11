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
	sys := components.NewSystem("Comfortstat", ctx)

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: " is a controller for a consumed servo motor position based on a consumed temperature",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/Comfortstat",
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
		ua, cleanup := newUnitAsset(uac, &sys, servsTemp)
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

// TODO: change the namne, will get one function for each of the four cases
// Serving handles the resources services. NOTE: it exepcts those names from the request URL path
func (t *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	case "min_temperature":
		t.set_minTemp(w, r)
	case "max_temperature":
		t.set_maxTemp(w, r)
	case "max_price":
		t.set_maxPrice(w, r)
	case "min_price":
		t.set_minPrice(w, r)
	case "SEK_price":
		t.set_SEKprice(w, r)
	case "desired_temp":
		t.set_desiredTemp(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configurration file]", http.StatusBadRequest)
	}
}

func (rsc *UnitAsset) set_SEKprice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		signalErr := rsc.getSEK_price()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

// TODO: split up this function to two sepreate function that sets on max and min price.
func (rsc *UnitAsset) set_minTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.setMin_temp(sig)
	case "GET":
		signalErr := rsc.getMin_temp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
func (rsc *UnitAsset) set_maxTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.setMax_temp(sig)
	case "GET":
		signalErr := rsc.getMax_temp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

// LOOK AT: I guess that we probable only need to if there is a PUT from user?
// LOOK AT: so not the GET!
// For PUT - the "HTTPProcessSetRequest(w, r)" is called to prosses the data given from the user and if no error, call set_minMaxprice with the value
// wich updates the value in thge struct
func (rsc *UnitAsset) set_minPrice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.setMin_price(sig)
	case "GET":
		signalErr := rsc.getMin_price()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)

	}
}

func (rsc *UnitAsset) set_maxPrice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.setMax_price(sig)
	case "GET":
		signalErr := rsc.getMax_price()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)

	}
}

func (rsc *UnitAsset) set_desiredTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			log.Println("Error with the setting request of the position ", err)
		}
		rsc.setDesired_temp(sig)
	case "GET":
		signalErr := rsc.getDesired_temp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
