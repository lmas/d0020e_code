package main

// This file was originally copied from:
// https://github.com/sdoque/systems/blob/main/ds18b20/ds18b20.go

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

func main() {
	// Handle graceful shutdowns using this context. It should always be canceled,
	// no matter the final execution path so all computer resources are freed up.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new Eclipse Arrowhead application system and then wrap it with a
	// "husk" (aka a wrapper or shell), which then sets up various properties and
	// operations that's required of an Arrowhead system.
	sys := components.NewSystem("ethermostat", ctx)
	sys.Husk = &components.Husk{
		Description: "reads the temperature from sensors",
		Details:     map[string][]string{"Developer": {"Group10"}},
		ProtoPort:   map[string]int{"https": 8691, "http": 8690, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/ethermostat",
	}

	// Try loading the config file (in JSON format) for this deployment,
	// by using a unit asset with default values.
	uat := initTemplate()
	sys.UAssets[uat.GetName()] = &uat
	rawUAs, servsTemp, err := usecases.Configure(&sys)
	// If the file is missing, a new config will be created and an error is returned here.
	if err != nil {
		log.Fatalf("Configuration error: %v\n", err)
	}

	// Load the proper unit asset(s) using the user-defined settings from the config file.
	clear(sys.UAssets)
	for _, raw := range rawUAs {
		var uac UnitAsset
		if err := json.Unmarshal(raw, &uac); err != nil {
			log.Fatalf("UnitAsset configuration error: %+v\n", err)
		}
		ua, cleanup := newUnitAsset(uac, &sys, servsTemp)
		sys.UAssets[ua.GetName()] = &ua
		defer cleanup()
	}

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.RequestCertificate(&sys)

	// Register the (system) and its services
	usecases.RegisterServices(&sys)

	// start the requests handlers and servers
	go usecases.SetoutServers(&sys)

	// Wait for the shutdown signal (ctrl+c) and gracefully terminate any goroutines by cancelling the context.
	<-sys.Sigs
	log.Println("Shuting down system: " + sys.Name)
	cancel()

	// Allow goroutines to finish execution (might take more time than main to end)
	time.Sleep(2 * time.Second)
}

////////////////////////////////////////////////////////////////////////////////

// Serving maps the requested service paths with any request handlers.
func (ua *UnitAsset) Serving(w http.ResponseWriter, r *http.Request, servicePath string) {
	switch servicePath {
	// TODO: match this subpath in a better way with the subpath defined in thing.go, ie. without relying on magic values
	case "temperature-sub":
		ua.getTemp(w, r)
	default:
		http.Error(w, "Invalid service request", http.StatusBadRequest)
	}
}

// getTemp returns the temperature of this sensor, using an analog signal form.
func (ua *UnitAsset) getTemp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	// Create and fill out the return form
	var f forms.SignalA_v1a
	f.NewForm()
	f.Value = ua.temperature
	f.Unit = "Celsius"
	f.Timestamp = time.Now()
	usecases.HTTPProcessGetRequest(w, r, &f)
}
