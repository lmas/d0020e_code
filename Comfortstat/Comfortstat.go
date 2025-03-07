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

	// Instantiate the Capsule
	sys.Husk = &components.Husk{
		Description: " is a controller for a consumed servo motor position based on a consumed temperature",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/Comfortstat",
	}

	// instantiate a template unit asset
	assetTemplate := initTemplate()
	// Calling initAPI() starts the pricefeedbackloop that fetches the current electricity price for the particular hour
	initAPI()
	time.Sleep(1 * time.Second)
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
		ua, startup := newUnitAsset(uac, &sys, servsTemp)
		startup()
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

// Serving handles the resources services. NOTE: it expects those names from the request URL path
func (t *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	case "MinTemperature":
		t.httpSetMinTemp(w, r)
	case "MaxTemperature":
		t.httpSetMaxTemp(w, r)
	case "MaxPrice":
		t.httpSetMaxPrice(w, r)
	case "MinPrice":
		t.httpSetMinPrice(w, r)
	case "SEKPrice":
		t.httpSetSEKPrice(w, r)
	case "DesiredTemp":
		t.httpSetDesiredTemp(w, r)
	case "UserTemp":
		t.httpSetUserTemp(w, r)
	case "Region":
		t.httpSetRegion(w, r)
	default:
		http.Error(w, "Invalid service request [Do not modify the services subpath in the configurration file]", http.StatusBadRequest)
	}
}

func (rsc *UnitAsset) httpSetSEKPrice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		signalErr := rsc.getSEKPrice()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

// All these functions below handles HTTP "PUT" or "GET" requests to modefy or retrieve the MAX/MIN temprature/price and desierd temperature
// For the PUT case - the "HTTPProcessSetRequest(w, r)" is called to prosses the data given from the user and if no error,
// call the set functions in things.go with the value witch updates the value in the struct
func (rsc *UnitAsset) httpSetMinTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			//log.Println("Error with the setting request of the position ", err)
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return

		}
		rsc.setMinTemp(sig)
	case "GET":
		signalErr := rsc.getMinTemp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
func (rsc *UnitAsset) httpSetMaxTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			//log.Println("Error with the setting request of the position ", err)
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setMaxTemp(sig)
	case "GET":
		signalErr := rsc.getMaxTemp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

func (rsc *UnitAsset) httpSetMinPrice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			//log.Println("Error with the setting request of the position ", err)
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setMinPrice(sig)
	case "GET":
		signalErr := rsc.getMinPrice()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)

	}
}

func (rsc *UnitAsset) httpSetMaxPrice(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			//log.Println("Error with the setting request of the position ", err)
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setMaxPrice(sig)
	case "GET":
		signalErr := rsc.getMaxPrice()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)

	}
}

func (rsc *UnitAsset) httpSetDesiredTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			//log.Println("Error with the setting request of the position ", err)
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setDesiredTemp(sig)
	case "GET":
		signalErr := rsc.getDesiredTemp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}

}

func (rsc *UnitAsset) httpSetUserTemp(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setUserTemp(sig)
	case "GET":
		signalErr := rsc.getUserTemp()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}

func (rsc *UnitAsset) httpSetRegion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		sig, err := usecases.HTTPProcessSetRequest(w, r)
		if err != nil {
			http.Error(w, "request incorrectly formatted", http.StatusBadRequest)
			return
		}
		rsc.setRegion(sig)
	case "GET":
		signalErr := rsc.getRegion()
		usecases.HTTPProcessGetRequest(w, r, &signalErr)
	default:
		http.Error(w, "Method is not supported.", http.StatusNotFound)
	}
}
