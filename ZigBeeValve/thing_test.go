package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

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
	// Highjack the default http client so no actuall http requests are sent over the network
	http.DefaultClient.Transport = t
	return t
}

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
		t.Error("Error expcted during unmarshalling, got nil instead", err)
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

var brokenURL string = string([]byte{0x7f})

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

func TestCreateRequests(t *testing.T) {
	// --- Good test case: createPutRequest() ---
	data := "test"
	apiURL := "http://localhost:8080/test"

	_, err := createPutRequest(data, apiURL)
	if err != nil {
		t.Error("Error occured, expected none")
	}

	// --- Bad test case: Error in createPutRequest() because of broken URL---
	_, err = createPutRequest(data, brokenURL)
	if err == nil {
		t.Error("Expected error because of broken URL")
	}

	// --- Good test case: createGetRequest() ---
	_, err = createGetRequest(apiURL)
	if err != nil {
		t.Error("Error occured, expected none")
	}

	// --- Bad test case: Error in createGetRequest() because of broken URL---
	_, err = createGetRequest(brokenURL)
	if err == nil {
		t.Error("Expected error because of broken URL")
	}

}

func TestSendRequests(t *testing.T) {
	// Set up standard response & catch http requests
	fakeBody := fmt.Sprint(`Test`)

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}

	// --- Good test case: sendPutRequest ---
	newMockTransport(resp, false, nil)
	apiURL := "http://localhost:8080/test"
	s := fmt.Sprintf(`{"heatsetpoint":%f}`, 25.0) // Create payload
	req, _ := createPutRequest(s, apiURL)
	err := sendPutRequest(req)
	if err != nil {
		t.Error("Expected no errors, error occured:", err)
	}

	// Break defaultClient.Do()
	// --- Error performing request ---
	newMockTransport(resp, false, fmt.Errorf("Test error"))
	s = fmt.Sprintf(`{"heatsetpoint":%f}`, 25.0) // Create payload
	req, _ = createPutRequest(s, apiURL)
	err = sendPutRequest(req)
	if err == nil {
		t.Error("Error expected while performing http request, got nil instead")
	}

	// Error unpacking body
	resp.Body = errReader(0)
	newMockTransport(resp, false, nil)

	err = sendPutRequest(req)

	if err == nil {
		t.Error("Expected errors, no error occured:")
	}

	// Error StatusCode
	resp.Body = io.NopCloser(strings.NewReader(fakeBody))
	resp.StatusCode = 300
	newMockTransport(resp, false, nil)
	err = sendPutRequest(req)
	if err != errStatusCode {
		t.Error("Expected errStatusCode, got", err)
	}
	// TODO: test sendGetRequest()
}

func TestGetSensors(t *testing.T) {
	// Setup for test
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Uniqueid = "54:ef:44:10:00:d8:82:8d-01"

	zBeeResponse := `{
		"1": {
		"state": {"consumption": 1},
		"name": "test consumption",
		"uniqueid": "54:ef:44:10:00:d8:82:8d-02-000c",
		"type": "ZHAConsumption"
		}, 
		"2": {
		"state": {"power": 1},
		"name": "test consumption",
		"uniqueid": "54:ef:44:10:00:d8:82:8d-03-000c",
		"type": "ZHAPower"
		}}`

	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}

	// --- Good case test---
	newMockTransport(zResp, false, nil)
	ua.getSensors()
	if ua.Slaves["ZHAConsumption"] != "54:ef:44:10:00:d8:82:8d-02-000c" {
		t.Errorf("Error with ZHAConsumption, wrong mac addr.")
	}
	if ua.Slaves["ZHAPower"] != "54:ef:44:10:00:d8:82:8d-03-000c" {
		t.Errorf("Error with ZHAPower, wrong mac addr.")
	}

	// --- Bad case: Error on createGetRequest() using brokenURL (bad character) ---
	gateway = brokenURL
	newMockTransport(zResp, false, nil)
	err := ua.getSensors()
	if err == nil {
		t.Errorf("Expected an error during createGetRequest() because gateway is an invalid control char")
	}

	// --- Bad case: Error while unmarshalling data ---
	gateway = "localhost:8080"
	FaultyzBeeResponse := `{
		"1": {
		"state": {"consumption": 1},
		"name": "test consumption",
		"uniqueid": "54:ef:44:10:00:d8:82:8d-02-000c"+123,
		"type": "ZHAConsumption"
		}, 
		"2": {
		"state": {"power": 1},
		"name": "test consumption",
		"uniqueid": "54:ef:44:10:00:d8:82:8d-03-000c"+123,
		"type": "ZHAPower"
		}}`

	zResp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(FaultyzBeeResponse)),
	}
	newMockTransport(zResp, false, nil)
	err = ua.getSensors()
	if err == nil {
		t.Errorf("Expected error while unmarshalling data because of broken uniqueid field")
	}

	// --- Bad case: Error while sending request ---
	newMockTransport(zResp, false, fmt.Errorf("Test error"))
	err = ua.getSensors()
	if err == nil {
		t.Errorf("Expected error during sendGetRequest()")
	}
}

func TestGetState(t *testing.T) {
	// Setup for test
	ua := initTemplate().(*UnitAsset)
	gateway = "localhost:8080"
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Uniqueid = "54:ef:44:10:00:d8:82:8d-01"
	zBeeResponseTrue := `{"state": {"on": true}}`
	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponseTrue)),
	}
	// --- Good test case: plug.State.On = true ---
	newMockTransport(zResp, false, nil)
	f, err := ua.getState()
	if f.Value != 1 {
		t.Errorf("Expected value to be 1, was %f", f.Value)
	}
	if err != nil {
		t.Errorf("Expected no errors got: %v", err)
	}

	// --- Good test case: plug.State.On = false ---
	zBeeResponseFalse := `{"state": {"on": false}}`
	zResp.Body = io.NopCloser(strings.NewReader(zBeeResponseFalse))
	newMockTransport(zResp, false, nil)
	f, err = ua.getState()

	if f.Value != 0 {
		t.Errorf("Expected value to be 0, was %f", f.Value)
	}

	if err != nil {
		t.Errorf("Expected no errors got: %v", err)
	}

	// --- Bad test case: Error on createGetRequest() ---
	gateway = brokenURL
	zResp.Body = io.NopCloser(strings.NewReader(zBeeResponseTrue))
	newMockTransport(zResp, false, nil)
	f, err = ua.getState()

	if err == nil {
		t.Errorf("Expected an error during createGetRequest() because gateway is an invalid control char")
	}

	gateway = "localhost:8080"

	// --- Bad test case: Error on unmarshal ---
	zResp.Body = errReader(0)
	newMockTransport(zResp, false, nil)
	f, err = ua.getState()

	if err == nil {
		t.Errorf("Expected an error while unmarshalling data")
	}
}

func TestSetState(t *testing.T) {
	// Setup
	gateway = "localhost:8080"
	var f forms.SignalA_v1a
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Uniqueid = "54:ef:44:10:00:d8:82:8d-01"
	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
	}
	// --- Good test case: f.Value = 1 ---
	newMockTransport(zResp, false, nil)
	f.NewForm()
	f.Value = 1
	f.Timestamp = time.Now()
	err := ua.setState(f)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err)
	}
	// --- Good test case: f.Value = 0 ---
	newMockTransport(zResp, false, nil)
	f.NewForm()
	f.Value = 0
	f.Timestamp = time.Now()
	err = ua.setState(f)
	if err != nil {
		t.Errorf("Expected no errors, got: %v", err)
	}
	// --- Bad test case: f.value is not 1 or 0
	newMockTransport(zResp, false, nil)
	f.NewForm()
	f.Value = 3
	f.Timestamp = time.Now()
	err = ua.setState(f)
	if err != errBadFormValue {
		t.Errorf("Expected error because of f.Value not being 0 or 1")
	}
}
