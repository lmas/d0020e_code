package main

import (
	"context"
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
	// Highjack the default http client so no actuall http requests are sent over the network
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

	setPointService := components.Service{
		Definition:  "setpoint",
		SubPath:     "setpoint",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current thermal setpoint (GET) or sets it (PUT)",
	}

	uac := UnitAsset{
		Name:    "Template",
		Details: map[string][]string{"Location": {"Kitchen"}},
		Model:   "SmartThermostat",
		Period:  10,
		Setpt:   20,
		Apikey:  "1234",
		ServicesMap: components.Services{
			setPointService.SubPath: &setPointService,
		},
	}

	ua, _ := newResource(uac, &sys, nil)
	// Happy test case:
	//Get a unitasset with correct values
	name := ua.GetName()
	if name != "Template" {
		t.Errorf("Expected name to be Template, but instead got: %v", name)
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

func TestProcessFeedbackLoop(t *testing.T) {
	// TODO: Test as much of the code as possible?
}

func TestFindGateway(t *testing.T) {
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
		t.Fatal("Gatewayn not found", err)
	}
	if gateway != "localhost:8080" {
		t.Fatalf("Expected gateway to be localhost:8080, was %s", gateway)
	}

	// ---- Error cases ----

	// Unmarshall error
	newMockTransport(resp, false, fmt.Errorf("Test error"))
	err = findGateway()
	if err == nil {
		t.Error("Error expcted, got nil instead", err)
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
		t.Error("Expected error")
	}

	// Actual http body is unmarshaled correctly
	resp.Body = io.NopCloser(strings.NewReader(fakeBody + "123"))
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err == nil {
		t.Error("Expected error")
	}

	// Empty list of gateways
	resp.Body = io.NopCloser(strings.NewReader("[]"))
	newMockTransport(resp, false, nil)
	err = findGateway()
	if err != errMissingGateway {
		t.Error("Expected error", err)
	}
}

func TestToggleState(t *testing.T) {
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

// func createRequest(data string, apiURL string) (req *http.Request, err error)
func TestCreateRequest(t *testing.T) {
	data := "test"
	apiURL := "http://localhost:8080/test"

	_, err := createRequest(data, apiURL)
	if err != nil {
		t.Error("Error occured, expected none")
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
		t.Error("Expected no errors, error occured:", err)
	}

	// Break defaultClient.Do()

	// Error unpacking body
	resp.Body = errReader(0)
	newMockTransport(resp, false, nil)

	err = sendRequest(req)

	if err == nil {
		t.Error("Expected errors, no error occured:")
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
