package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/components"
)

type mockConfigure struct {
	createFile bool
	badUnit    bool
}

func (c *mockConfigure) load(sys *components.System) (raws []json.RawMessage, servs []components.Service, err error) {
	if c.createFile {
		err = fmt.Errorf("a new configuration file has been written")
		return
	}
	if c.badUnit {
		raws = []json.RawMessage{json.RawMessage("}")}
		return
	}
	b, err := json.Marshal(initTemplate())
	if err != nil {
		return
	}
	raws = []json.RawMessage{json.RawMessage(b)}
	return
}

func TestLoadConfig(t *testing.T) {
	sys := newSystem()

	// Good case: loads config
	conf := &mockConfigure{}
	configureSystem = conf.load
	if err := sys.loadConfiguration(); err != nil {
		t.Fatalf("Expected nil error, got: %s", err)
	}
	_, found := sys.UAssets[uaName]
	if !found {
		t.Fatalf("Expected to find loaded unitasset, got nil")
	}

	// Bad case: stop system startup if config file is missing
	conf = &mockConfigure{createFile: true}
	configureSystem = conf.load
	if err := sys.loadConfiguration(); err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Bad case: fails to unmarshal json for unit
	conf = &mockConfigure{badUnit: true}
	configureSystem = conf.load
	if err := sys.loadConfiguration(); err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

////////////////////////////////////////////////////////////////////////////////

var errShutdown = fmt.Errorf("test startup error")

func TestListenAndServe(t *testing.T) {
	// Bad case: startup returns error
	sys := newSystem()
	sys.startups = []func() error{
		func() error {
			return errShutdown
		},
	}

	c := make(chan bool)
	go func(logf func(string, ...any)) {
		if err := sys.listenAndServe(); !errors.Is(err, errShutdown) {
			logf("Expected startup error, got: %s", err)
		}
		close(c)
	}(t.Errorf)

	// Wait for graceful shutdown, fail if it times out.
	// The timeout might cause flaky testing here (if the shutdown takes longer
	// than usual). I'm averaging about 1s on a laptop.
	select {
	case <-c:
	case <-time.After(2000 * time.Millisecond):
		t.Fatalf("Expected startup to quit and call close(), but timed out")
	}

	// NOTE: Don't bother trying to test for errors from usecases.SetoutServers()
}
