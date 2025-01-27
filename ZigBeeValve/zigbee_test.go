package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

// mockTransport is used for replacing the default network Transport (used by
// http.DefaultClient) and it will intercept network requests.

type mockTransport struct {
	hits map[string]int
}

func newMockTransport() mockTransport {
	t := mockTransport{
		hits: make(map[string]int),
	}
	// Highjack the default http client so no actuall http requests are sent over the network
	http.DefaultClient.Transport = t
	return t
}

// domainHits returns the number of requests to a domain (or -1 if domain wasn't found).

func (t mockTransport) domainHits(domain string) int {
	for u, hits := range t.hits {
		if u == domain {
			return hits
		}
	}
	return -1
}

// TODO: this might need to be expanded to a full JSON array?

const priceExample string = `[{
	"SEK_per_kWh": 0.26673,
	"EUR_per_kWh": 0.02328,
	"EXR": 11.457574,
	"time_start": "2025-01-06T%02d:00:00+01:00",
	"time_end": "2025-01-06T%02d:00:00+01:00"
}]`

// RoundTrip method is required to fulfil the RoundTripper interface (as required by the DefaultClient).
// It prevents the request from being sent over the network and count how many times
// a domain was requested.

func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	hour := time.Now().Local().Hour()
	fakeBody := fmt.Sprintf(priceExample, hour, hour+1)
	// TODO: should be able to adjust these return values for the error cases
	resp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Request:    req,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	t.hits[req.URL.Hostname()] += 1
	return
}

////////////////////////////////////////////////////////////////////////////////

const thermostatDomain string = "http://localhost:8870/api/B3AFB6415A/sensors/2/config"
const plugDomain string = "http://localhost:8870/api/B3AFB6415A/lights/1/config"

func TestUnitAsset(t *testing.T) {

	// Don't understand how to check my own deConz API calls, will extend the test with this once i understand
	trans := newMockTransport()

	// Create a form
	f := forms.SignalA_v1a{
		Value: 27.0,
	}

	// Creates a single UnitAsset and assert it changes
	ua := initTemplate().(*UnitAsset)

	// Change Setpt
	ua.setSetPoint(f)

	if ua.Setpt != 27.0 {
		t.Errorf("Expected Setpt to be 27.0, instead got %f", ua.Setpt)
	}

	// TODO: Add api call to make sure it only sends update to HW once.
	hits := trans.domainHits(thermostatDomain)
	if hits > 1 {
		t.Errorf("Expected number of api requests = 1, got %d requests", hits)
	}
}

func TestGetters(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	name := ua.GetName()
	if name != "Template" {
		t.Errorf("Expected name to be 2, instead got %s", name)
	}

	services := ua.GetServices()
	if services == nil {
		t.Fatalf("Expected services not to be nil")
	}
	if services["setpoint"].Definition != "setpoint" {
		t.Errorf("Expected definition to be setpoint")
	}

	details := ua.GetDetails()
	if details == nil {
		t.Fatalf("Details was nil, expected map")
	}
	if len(details["Location"]) == 0 {
		t.Fatalf("Location was nil, expected kitchen")
	}

	if details["Location"][0] != "Kitchen" {
		t.Errorf("Expected location to be Kitchen")
	}

	cervices := ua.GetCervices()
	if cervices != nil {
		t.Errorf("Expected no cervices")
	}
}

func TestNewResource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sys := components.NewSystem("testsys", ctx)

	sys.Husk = &components.Husk{
		Description: " is a controller for smart thermostats connected with a RaspBee II",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8870, "coap": 0},
		InfoLink:    "https://github.com/sdoque/systems/tree/master/ZigBeeValve",
	}

	assetTemplate := initTemplate()
	assetName := assetTemplate.GetName()
	sys.UAssets[assetName] = &assetTemplate

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
		ua, _ := newResource(uac, &sys, servsTemp)
		//startup()
		sys.UAssets[ua.GetName()] = &ua
	}
}

func TestFeedbackLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sys := components.NewSystem("testsys", ctx)

	sys.Husk = &components.Husk{
		Description: " is a controller for smart thermostats connected with a RaspBee II",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8870, "coap": 0},
		InfoLink:    "https://github.com/sdoque/systems/tree/master/ZigBeeValve",
	}

	assetTemplate := initTemplate()
	assetName := assetTemplate.GetName()
	sys.UAssets[assetName] = &assetTemplate

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
		ua, _ := newResource(uac, &sys, servsTemp)
		//startup()
		sys.UAssets[ua.GetName()] = &ua
	}

	// TODO: Test feedbackloop and processfeedbackloop
	/*
		ua.feedbackLoop(ctx)
		cancel()
		time.Sleep(2 * time.Second)
	*/
}
