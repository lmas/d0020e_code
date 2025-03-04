package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/usecases"
)

func main() {
	sys := newSystem()
	if err := sys.loadConfiguration(); err != nil {
		log.Fatalf("Error loading config: %s\n", err)
	}

	// Generate PKI keys and CSR to obtain a authentication certificate from the CA
	usecases.RequestCertificate(&sys.System)

	// Register the system and its services
	// WARN: this func runs a goroutine of it's own, which makes it hard to count
	// using the waitgroup (and I can't be arsed to do it properly...)
	usecases.RegisterServices(&sys.System)

	// Run forever
	if err := sys.listenAndServe(); err != nil {
		log.Fatalf("Error running system: %s\n", err)
	}
}

////////////////////////////////////////////////////////////////////////////////

// There's no interface to use, so have to encapsulate the base struct instead.
// This allows for access/storage of internal vars shared system-wide.
type system struct {
	components.System

	cancel   func()
	startups []func() error
}

const systemName string = "Collector"

// Creates a new system with a context and husk prepared for later use.
func newSystem() (sys *system) {
	// Handle graceful shutdowns using this context. It should always be canceled,
	// no matter the final execution path so all computer resources are freed up.
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new Eclipse Arrowhead application system and then wrap it with a
	// "husk" (aka a wrapper or shell), which then sets up various properties and
	// operations that's required of an Arrowhead system.
	// var sys system
	sys = &system{
		System: components.NewSystem(systemName, ctx),
		cancel: cancel,
	}
	sys.Husk = &components.Husk{
		Description: "pulls data from other Arrorhead systems and sends it to a InfluxDB server.",
		Details:     map[string][]string{"Developer": {"Alex"}},
		ProtoPort:   map[string]int{"https": 8666, "http": 8666, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/collector",
	}
	return
}

// Allows for mocking this extremely heavy function call
var configureSystem = usecases.Configure

// Try load configuration from the standard "systemconfig.json" file.
// Any unit assets will be prepared for later startup.
// WARN: An error is raised if the config file is missing!
func (sys *system) loadConfiguration() (err error) {
	// Try loading the config file (in JSON format) for this deployment,
	// by using a unit asset with default values.
	uat := components.UnitAsset(initTemplate())
	sys.UAssets[uat.GetName()] = &uat
	rawUAs, _, err := configureSystem(&sys.System)

	// If the file is missing, a new config will be created and an error is returned here.
	if err != nil {
		return
	}

	// Load the proper unit asset(s) using the user-defined settings from the config file.
	clear(sys.UAssets)
	for _, raw := range rawUAs {
		var uac unitAsset
		if err := json.Unmarshal(raw, &uac); err != nil {
			return fmt.Errorf("unmarshalling json config: %w", err)
		}
		ua := newUnitAsset(uac, sys)
		sys.startups = append(sys.startups, ua.startup)
		intf := components.UnitAsset(ua)
		sys.UAssets[ua.GetName()] = &intf
	}
	return
}

// Run the system and all the unit assets, blocking until user cancels or an
// error is raised in any background workers.
func (sys *system) listenAndServe() (err error) {
	var wg sync.WaitGroup // Used for counting all started goroutines

	// start a web server that serves basic documentation of the system
	wg.Add(1)
	go func() {
		if e := usecases.SetoutServers(&sys.System); e != nil {
			err = fmt.Errorf("web server: %w", e)
			sys.cancel()
		}
		wg.Done()
	}()

	// Run all the startups in separate goroutines and keep track of them
	for _, f := range sys.startups {
		wg.Add(1)
		go func(start func() error) {
			if e := start(); e != nil {
				err = fmt.Errorf("startup: %w", e)
				sys.cancel()
			}
			wg.Done()
		}(f)
	}

	// Block and wait for either a...
	select {
	case <-sys.Sigs: // user initiated shutdown signal (ctrl+c) or a...
		log.Println("Initiated shutdown, waiting for workers to terminate")
	case <-sys.Ctx.Done(): // shutdown request from a worker
	}

	// Gracefully terminate any leftover goroutines and wait for them to shutdown properly
	sys.cancel()
	wg.Wait()
	return
}
