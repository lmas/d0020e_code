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

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

// mockTransport is used for replacing the default network Transport (used by
// http.DefaultClient) and it will intercept network requests.

type mockTransport struct {
	resp *http.Response
	hits map[string]int
	err  error
}

func newMockTransport(resp *http.Response, err error) mockTransport {
	t := mockTransport{
		resp: resp,
		hits: make(map[string]int),
		err:  err,
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

const discoverExample string = `[{
		"Id": "123",
		"Internalipaddress": "localhost",
		"Macaddress": "test",
		"Internalport": 8080,
		"Name": "My gateway",
		"Publicipaddress": "test"
	}]`

// RoundTrip method is required to fulfil the RoundTripper interface (as required by the DefaultClient).
// It prevents the request from being sent over the network and count how many times
// a domain was requested.

func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.resp.Request = req
	t.hits[req.URL.Hostname()] += 1
	if t.err != nil {
		return nil, t.err
	}
	return t.resp, nil
}

////////////////////////////////////////////////////////////////////////////////

func TestUnitAsset(t *testing.T) {
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

	// Fetch Setpt w/ a form
	f2 := ua.getSetPoint()

	if f2.Value != f.Value {
		t.Errorf("Expected %f, instead got %f", f.Value, f2.Value)
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

const thermostatDomain string = "http://localhost:8870/api/B3AFB6415A/sensors/2/config"
const plugDomain string = "http://localhost:8870/api/B3AFB6415A/lights/1/config"

type errReader int

var errBodyRead error = fmt.Errorf("bad body read")

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errBodyRead
}

func (errReader) Close() error {
	return nil
}

func TestFindGateway(t *testing.T) {
	// New mocktransport
	//gatewayDomain := "https://phoscon.de/discover"

	fakeBody := fmt.Sprint(discoverExample)

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	newMockTransport(resp, nil)

	// ---- All ok! ----
	err := findGateway()

	if err != nil {
		t.Fatal("Gatewayn not found", err)
	}
	if gateway != "localhost:8080" {
		t.Fatalf("Expected gateway to be localhost:8080, was %s", gateway)
	}

	// Have to make changes to mockTransport to test this?
	// ---- Error cases ----

	// Unmarshall error
	newMockTransport(resp, fmt.Errorf("Test error"))
	err = findGateway()
	if err == nil {
		t.Error("Error expcted, got nil instead", err)
	}

	// Statuscode > 299, have to make changes to mockTransport to test this
	resp.StatusCode = 300
	newMockTransport(resp, nil)
	err = findGateway()
	if err != errStatusCode {
		t.Error("Expected errStatusCode, got", err)
	}

	// Broken body - https://stackoverflow.com/questions/45126312/how-do-i-test-an-error-on-reading-from-a-request-body
	resp.StatusCode = 200
	resp.Body = errReader(0)
	newMockTransport(resp, nil)
	err = findGateway()
	if err != errBodyRead {
		t.Error("Expected error")
	}

	// Actual http body is unmarshaled correctly
	resp.Body = io.NopCloser(strings.NewReader(fakeBody + "123"))
	newMockTransport(resp, nil)
	err = findGateway()
	if err == nil {
		t.Error("Expected error")
	}

	// Empty list of gateways
	resp.Body = io.NopCloser(strings.NewReader("[]"))
	newMockTransport(resp, nil)
	err = findGateway()
	if err != errMissingGateway {
		t.Error("Expected error", err)
	}
}

func TestToggleState(t *testing.T) {
	fakeBody := fmt.Sprint("")

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}

	newMockTransport(resp, nil)

	ua := initTemplate().(*UnitAsset)

	ua.toggleState(true)
}

func TestSendSetPoint(t *testing.T) {
	fakeBody := fmt.Sprint(`{"Value": 12.4, "Version": "SignalA_v1a}`)

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}

	newMockTransport(resp, nil)

	ua := initTemplate().(*UnitAsset)

	ua.sendSetPoint()

}
