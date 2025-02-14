package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
)

// mockTransport is used for replacing the default network Transport (used by
// http.DefaultClient) and it will intercept network requests.

type mockTransport struct {
	returnError bool
	resp        *http.Response
	hits        map[string]int
	err         error
}

func newMockTransport(resp *http.Response, retErr bool, err error) mockTransport {
	t := mockTransport{
		returnError: retErr,
		resp:        resp,
		hits:        make(map[string]int),
		err:         err,
	}
	// Hijack the default http client so no actual http requests are sent over the network
	http.DefaultClient.Transport = t
	return t
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

var errHTTP error = fmt.Errorf("bad http request")

func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.hits[req.URL.Hostname()] += 1
	if t.err != nil {
		return nil, t.err
	}
	if t.returnError != false {
		req.GetBody = func() (io.ReadCloser, error) {
			return nil, errHTTP
		}
	}
	t.resp.Request = req
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
	// Test GetName()
	name := ua.GetName()
	if name != "SmartThermostat1" {
		t.Errorf("Expected name to be SmartThermostat1, instead got %s", name)
	}
	// Test GetServices()
	services := ua.GetServices()
	if services == nil {
		t.Fatalf("Expected services not to be nil")
	}
	if services["setpoint"].Definition != "setpoint" {
		t.Errorf("Expected definition to be setpoint")
	}
	// Test GetDetails()
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
	// Test GetCervices()
	cervices := ua.GetCervices()
	if cervices != nil {
		t.Errorf("Expected no cervices")
	}
}

func TestNewResource(t *testing.T) {
	// Setup test context, system and unitasset
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
	setPointService := components.Service{
		Definition:  "setpoint",
		SubPath:     "setpoint",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current thermal setpoint (GET) or sets it (PUT)",
	}
	uac := UnitAsset{
		Name:    "SmartThermostat1",
		Details: map[string][]string{"Location": {"Kitchen"}},
		Model:   "ZHAThermostat",
		Period:  10,
		Setpt:   20,
		Apikey:  "1234",
		ServicesMap: components.Services{
			setPointService.SubPath: &setPointService,
		},
	}
	// Test newResource function
	ua, _ := newResource(uac, &sys, nil)
	// Happy test case:
	name := ua.GetName()
	if name != "SmartThermostat1" {
		t.Errorf("Expected name to be SmartThermostat1, but instead got: %v", name)
	}
}

type errReader int

var errBodyRead error = fmt.Errorf("bad body read")

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errBodyRead
}

func (errReader) Close() error {
	return nil
}

func TestFindGateway(t *testing.T) {
	// Create mock response for findGateway function
	fakeBody := fmt.Sprint(discoverExample)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	newMockTransport(resp, false, nil)

	// ---- All ok! ----
	err := findGateway()
	if err != nil {
		t.Fatal("Gateway not found", err)
	}
	if gateway != "localhost:8080" {
		t.Fatalf("Expected gateway to be localhost:8080, was %s", gateway)
	}

	// ---- Error cases ----
	// Unmarshall error
	newMockTransport(resp, false, fmt.Errorf("Test error"))
	err = findGateway()
	if err == nil {
		t.Error("Error expected during unmarshalling, got nil instead", err)
	}

	// Statuscode > 299, have to make changes to mockTransport to test this
	resp.StatusCode = 300
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err != errStatusCode {
		t.Error("Expected errStatusCode, got", err)
	}

	// Broken body - https://stackoverflow.com/questions/45126312/how-do-i-test-an-error-on-reading-from-a-request-body
	resp.StatusCode = 200
	resp.Body = errReader(0)
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err != errBodyRead {
		t.Error("Expected errBodyRead, got", err)
	}

	// Actual http body is unmarshaled incorrectly
	resp.Body = io.NopCloser(strings.NewReader(fakeBody + "123"))
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err == nil {
		t.Error("Expected error while unmarshalling body, error:", err)
	}

	// Empty list of gateways
	resp.Body = io.NopCloser(strings.NewReader("[]"))
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err != errMissingGateway {
		t.Error("Expected error when list of gateways is empty:", err)
	}
}

func TestToggleState(t *testing.T) {
	// Create mock response and unitasset for toggleState() function
	fakeBody := fmt.Sprint(`{"on":true, "Version": "SignalA_v1a"}`)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	newMockTransport(resp, false, nil)
	ua := initTemplate().(*UnitAsset)
	// All ok!
	ua.toggleState(true)
	// Error
	// change gateway to bad character/url, return gateway to original value
	gateway = brokenURL
	ua.toggleState(true)
	findGateway()
}

func TestSendSetPoint(t *testing.T) {
	// Create mock response and unitasset for sendSetPoint() function
	fakeBody := fmt.Sprint(`{"Value": 12.4, "Version": "SignalA_v1a}`)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	newMockTransport(resp, false, nil)
	ua := initTemplate().(*UnitAsset)
	// All ok!
	gateway = "localhost"
	err := ua.sendSetPoint()
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	// Error
	gateway = brokenURL
	ua.sendSetPoint()
	findGateway()
	gateway = "localhost"
}

type testJSON struct {
	FirstAttr string `json:"firstAttr"`
	Uniqueid  string `json:"uniqueid"`
	ThirdAttr string `json:"thirdAttr"`
}

func TestGetConnectedUnits(t *testing.T) {
	gateway = "localhost"
	// Set up standard response & catch http requests
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       nil,
	}
	ua := initTemplate().(*UnitAsset)
	ua.Uniqueid = "123test"

	// --- Broken body ---
	newMockTransport(resp, false, nil)
	resp.Body = errReader(0)
	err := ua.getConnectedUnits(ua.Model)

	if err == nil {
		t.Error("Expected error while unpacking body in getConnectedUnits()")
	}

	// --- All ok! ---
	// Make a map
	fakeBody := make(map[string]testJSON)
	test := testJSON{
		FirstAttr: "123",
		Uniqueid:  "123test",
		ThirdAttr: "456",
	}
	// Insert the JSON into the map with key="1"
	fakeBody["1"] = test
	// Marshal and create response
	jsonBody, _ := json.Marshal(fakeBody)
	resp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(jsonBody))),
	}
	// Start up a newMockTransport to capture HTTP requests before they leave
	newMockTransport(resp, false, nil)
	// Test function
	err = ua.getConnectedUnits(ua.Model)
	if err != nil {
		t.Error("Expected no errors, error occurred:", err)
	}

	// --- Bad statuscode ---
	resp.StatusCode = 300
	newMockTransport(resp, false, nil)
	err = ua.getConnectedUnits(ua.Model)
	if err == nil {
		t.Errorf("Expected status code > 299 in getConnectedUnits(), got %v", resp.StatusCode)
	}

	// --- Missing uniqueid ---
	// Make a map
	fakeBody = make(map[string]testJSON)
	test = testJSON{
		FirstAttr: "123",
		Uniqueid:  "missing",
		ThirdAttr: "456",
	}
	// Insert the JSON into the map with key="1"
	fakeBody["1"] = test
	// Marshal and create response
	jsonBody, _ = json.Marshal(fakeBody)
	resp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(jsonBody))),
	}
	// Start up a newMockTransport to capture HTTP requests before they leave
	newMockTransport(resp, false, nil)
	// Test function
	err = ua.getConnectedUnits(ua.Model)
	if err != errMissingUniqueID {
		t.Error("Expected uniqueid to be missing when running getConnectedUnits()")
	}

	// --- Unmarshall error ---
	resp.Body = io.NopCloser(strings.NewReader(string(jsonBody) + "123"))
	newMockTransport(resp, false, nil)
	err = ua.getConnectedUnits(ua.Model)
	if err == nil {
		t.Error("Error expected during unmarshalling, got nil instead", err)
	}

	// --- Error performing request ---
	newMockTransport(resp, false, fmt.Errorf("Test error"))
	err = ua.getConnectedUnits(ua.Model)
	if err == nil {
		t.Error("Error expected while performing http request, got nil instead")
	}
}

// func createRequest(data string, apiURL string) (req *http.Request, err error)
func TestCreateRequest(t *testing.T) {
	data := "test"
	apiURL := "http://localhost:8080/test"

	_, err := createRequest(data, apiURL)
	if err != nil {
		t.Error("Error occurred, expected none")
	}

	_, err = createRequest(data, brokenURL)
	if err == nil {
		t.Error("Expected error")
	}

}

var brokenURL string = string([]byte{0x7f})

func TestSendRequest(t *testing.T) {
	// Set up standard response & catch http requests
	fakeBody := fmt.Sprint(`Test`)

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}

	// All ok!
	newMockTransport(resp, false, nil)
	apiURL := "http://localhost:8080/test"
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, 25.0) // Create payload
	req, _ := createRequest(s, apiURL)
	err := sendRequest(req)
	if err != nil {
		t.Error("Expected no errors, error occurred:", err)
	}

	// Break defaultClient.Do()
	// --- Error performing request ---
	newMockTransport(resp, false, fmt.Errorf("Test error"))
	s = fmt.Sprintf(`{"heatsetpoint":%f}`, 25.0) // Create payload
	req, _ = createRequest(s, apiURL)
	err = sendRequest(req)
	if err == nil {
		t.Error("Error expected while performing http request, got nil instead")
	}

	// Error unpacking body
	resp.Body = errReader(0)
	newMockTransport(resp, false, nil)

	err = sendRequest(req)

	if err == nil {
		t.Error("Expected errors, no error occurred:")
	}

	// Error StatusCode
	resp.Body = io.NopCloser(strings.NewReader(fakeBody))
	resp.StatusCode = 300
	newMockTransport(resp, false, nil)
	err = sendRequest(req)
	if err != errStatusCode {
		t.Error("Expected errStatusCode, got", err)
	}

}
