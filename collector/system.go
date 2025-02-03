package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/usecases"
)

func main() {
	sys := newSystem()
	sys.loadConfiguration()

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.RequestCertificate(&sys.System)

	// Register the system and its services
	// WARN: this func runs a goroutine of it's own, which makes it hard to count
	// using the waitgroup (and I can't be arsed to do it properly...)
	usecases.RegisterServices(&sys.System)

	// Run forever
	sys.listenAndServe()
}

////////////////////////////////////////////////////////////////////////////////

// There's no interface to use, so have to encapsulate the base struct instead
type system struct {
	components.System

	cancel   func()
	startups []func() error
}

func newSystem() (sys *system) {
	// Handle graceful shutdowns using this context. It should always be canceled,
	// no matter the final execution path so all computer resources are freed up.
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new Eclipse Arrowhead application system and then wrap it with a
	// "husk" (aka a wrapper or shell), which then sets up various properties and
	// operations that's required of an Arrowhead system.
	// var sys system
	sys = &system{
		System: components.NewSystem("Collector", ctx),
		cancel: cancel,
	}
	sys.Husk = &components.Husk{
		Description: "pulls data from other Arrorhead systems and sends it to a InfluxDB server.",
		Details:     map[string][]string{"Developer": {"Alex"}},
		ProtoPort:   map[string]int{"https": 8691, "http": 8690, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/collector",
	}
	return
}

func (sys *system) loadConfiguration() {
	// Try loading the config file (in JSON format) for this deployment,
	// by using a unit asset with default values.
	uat := initTemplate()
	sys.UAssets[uat.GetName()] = &uat
	rawUAs, servsTemp, err := usecases.Configure(&sys.System)
	// If the file is missing, a new config will be created and an error is returned here.
	if err != nil {
		// TODO: it would had been nice to catch the exact error for "created config.."
		// and not display it as an actuall error, per se.
		log.Fatalf("Error while reading configuration: %v\n", err)
	}

	// Load the proper unit asset(s) using the user-defined settings from the config file.
	clear(sys.UAssets)
	for _, raw := range rawUAs {
		var uac unitAsset
		if err := json.Unmarshal(raw, &uac); err != nil {
			log.Fatalf("Error while unmarshalling configuration: %+v\n", err)
		}
		ua, startup := newUnitAsset(uac, &sys.System, servsTemp)
		sys.UAssets[ua.GetName()] = &ua
		sys.startups = append(sys.startups, startup)
	}
}

func (sys *system) listenAndServe() {
	var wg sync.WaitGroup // Used for counting all started goroutines

	// start a web server that serves basic documentation of the system
	wg.Add(1)
	go func() {
		if err := usecases.SetoutServers(&sys.System); err != nil {
			log.Println("Error while running web server:", err)
			sys.cancel()
		}
		wg.Done()
	}()

	// Run all the startups in separate goroutines and keep track of them
	for _, f := range sys.startups {
		wg.Add(1)
		go func(start func() error) {
			if err := start(); err != nil {
				log.Printf("Error while running collector: %s\n", err)
				sys.cancel()
			}
			wg.Done()
		}(f)
	}

	// Block and wait for either a...
	select {
	case <-sys.Sigs: // user initiated shutdown signal (ctrl+c) or a...
	case <-sys.Ctx.Done(): // shutdown request from a worker
	}

	// Gracefully terminate any leftover goroutines and wait for them to shutdown properly
	log.Println("Initiated shutdown, waiting for workers to terminate")
	sys.cancel()
	wg.Wait()
}
