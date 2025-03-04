package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	//"io"
	"net/http"
	//"strings"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
)

// mockTransport is used for replacing the default network Transport (used by
// http.DefaultClient) and it will intercept network requests.

type mockTransport struct {
	resp *http.Response
	hits map[string]int
}

func newMockTransport(resp *http.Response) mockTransport {
	t := mockTransport{
		resp: resp,
		hits: make(map[string]int),
	}
	// Hijack the default http client so no actual http requests are sent over the network
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

var sunDataExample string = fmt.Sprintf(`[{
			"results": {
				"date": "%d-%02d-%02d",
				"sunrise": "08:00:00",
				"sunset": "20:00:00",
				"first_light": "07:00:00",
				"last_light": "21:00:00",
				"dawn": "07:30:00",
				"dusk": "20:30:00",
				"solar_noon": "16:00:00",
				"golden_hour": "19:00:00",
				"day_length": "12:00:00"
				"timezone": "CET",
				"utc_offset": "1"
			},
			"status": "OK"
}]`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())

// RoundTrip method is required to fulfil the RoundTripper interface (as required by the DefaultClient).
// It prevents the request from being sent over the network and count how many times
// a domain was requested.
func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.hits[req.URL.Hostname()] += 1
	t.resp.Request = req
	return t.resp, nil
}

// //////////////////////////////////////////////////////////////////////////////
const apiDomain string = "https://sunrisesunset.io/"

func TestSingleUnitAssetOneAPICall(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//maby better to use a getter method
	url := fmt.Sprintf(`http://api.sunrisesunset.io/json?lat=%06f&lng=%06f&timezone=CET&date=%d-%02d-%02d&time_format=24`, ua.Latitude, ua.Longitude, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		//Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	trans := newMockTransport(resp)
	// Creates a single UnitAsset and assert it only sends a single API request

	//retrieveAPIPrice(ua)
	// better to use a getter method??
	ua.getAPIData(url)

	// TEST CASE: cause a single API request
	hits := trans.domainHits(apiDomain)
	if hits > 1 {
		t.Errorf("expected number of api requests = 1, got %d requests", hits)
	}
}

func TestSetMethods(t *testing.T) {
	asset := initTemplate().(*UnitAsset)

	// Simulate the input signals
	buttonStatus := forms.SignalA_v1a{
		Value: 0,
	}
	//call and test ButtonStatus
	asset.setButtonStatus(buttonStatus)
	if asset.ButtonStatus != 0 {
		t.Errorf("expected ButtonStatus to be 0, got %f", asset.ButtonStatus)
	}
	// Simulate the input signals
	latitude := forms.SignalA_v1a{
		Value: 65.584816,
	}
	// call and test Latitude
	asset.setLatitude(latitude)
	if asset.Latitude != 65.584816 {
		t.Errorf("expected Latitude to be 65.584816, got %f", asset.Latitude)
	}
	// Simulate the input signals
	longitude := forms.SignalA_v1a{
		Value: 22.156704,
	}
	//call and test MinPrice
	asset.setLongitude(longitude)
	if asset.Longitude != 22.156704 {
		t.Errorf("expected Longitude to be 22.156704, got %f", asset.Longitude)
	}
}

func TestGetMethods(t *testing.T) {
	uasset := initTemplate().(*UnitAsset)

	// ButtonStatus
	// check if the value from the struct is the actual value that the func is getting
	result1 := uasset.getButtonStatus()
	if result1.Value != uasset.ButtonStatus {
		t.Errorf("expected Value of the ButtonStatus is to be %v, got %v", uasset.ButtonStatus, result1.Value)
	}
	//check that the Unit is correct
	if result1.Unit != "bool" {
		t.Errorf("expected Unit to be 'bool', got %v", result1.Unit)
	}
	// Latitude
	// check if the value from the struct is the actual value that the func is getting
	result2 := uasset.getLatitude()
	if result2.Value != uasset.Latitude {
		t.Errorf("expected Value of the Latitude is to be %v, got %v", uasset.Latitude, result2.Value)
	}
	//check that the Unit is correct
	if result2.Unit != "Degrees" {
		t.Errorf("expected Unit to be 'Degrees', got %v", result2.Unit)
	}
	// Longitude
	// check if the value from the struct is the actual value that the func is getting
	result3 := uasset.getLongitude()
	if result3.Value != uasset.Longitude {
		t.Errorf("expected Value of the Longitude is to be %v, got %v", uasset.Longitude, result3.Value)
	}
	//check that the Unit is correct
	if result3.Unit != "Degrees" {
		t.Errorf("expected Unit to be 'Degrees', got %v", result3.Unit)
	}
}

func TestInitTemplate(t *testing.T) {
	uasset := initTemplate().(*UnitAsset)

	//// unnecessary test, but good for practicing
	name := uasset.GetName()
	if name != "Button" {
		t.Errorf("expected name of the resource is %v, got %v", uasset.Name, name)
	}
	Services := uasset.GetServices()
	if Services == nil {
		t.Fatalf("If Services is nil, not worth to continue testing")
	}
	// Services
	if Services["ButtonStatus"].Definition != "ButtonStatus" {
		t.Errorf("expected service definition to be ButtonStatus")
	}
	if Services["Latitude"].Definition != "Latitude" {
		t.Errorf("expected service definition to be Latitude")
	}
	if Services["Longitude"].Definition != "Longitude" {
		t.Errorf("expected service definition to be Longitude")
	}
	//GetCervice//
	Cervices := uasset.GetCervices()
	if Cervices != nil {
		t.Fatalf("If cervises not nil, not worth to continue testing")
	}
	//Testing Details//
	Details := uasset.GetDetails()
	if Details == nil {
		t.Errorf("expected a map, but Details was nil, ")
	}
}

func TestNewUnitAsset(t *testing.T) {
	// prepare for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background()) // create a context that can be cancelled
	defer cancel()                                          // make sure all paths cancel the context to avoid context leak
	// instantiate the System
	sys := components.NewSystem("SunButton", ctx)
	// Instantiate the Capsule
	sys.Husk = &components.Husk{
		Description: " is a controller for a consumed smart plug based on status depending on the sun",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/Comfortstat/SunButton",
	}
	setButtonStatus := components.Service{
		Definition:  "ButtonStatus",
		SubPath:     "ButtonStatus",
		Details:     map[string][]string{"Unit": {"bool"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current button status (using a GET request)",
	}
	setLatitude := components.Service{
		Definition:  "Latitude",
		SubPath:     "Latitude",
		Details:     map[string][]string{"Unit": {"Degrees"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the latitude (using a GET request)",
	}
	setLongitude := components.Service{
		Definition:  "Longitude",
		SubPath:     "Longitude",
		Details:     map[string][]string{"Unit": {"Degrees"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the longitude (using a GET request)",
	}
	// New UnitAsset struct init
	uac := UnitAsset{
		//These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name:         "Button",
		Details:      map[string][]string{"Location": {"Kitchen"}},
		ButtonStatus: 0.5,
		Latitude:     65.584816,
		Longitude:    22.156704,
		Period:       15,
		data:         Data{SunData{}, ""},

		// Maps the provided services from above
		ServicesMap: components.Services{
			setButtonStatus.SubPath: &setButtonStatus,
			setLatitude.SubPath:     &setLatitude,
			setLongitude.SubPath:    &setLongitude,
		},
	}

	ua, _ := newUnitAsset(uac, &sys, nil)
	// Calls the method that gets the name of the new UnitAsset
	name := ua.GetName()
	if name != "Button" {
		t.Errorf("expected name to be Button, but got: %v", name)
	}
}

// Functions that help creating bad body
type errReader int

var errBodyRead error = fmt.Errorf("bad body read")

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errBodyRead
}

func (errReader) Close() error {
	return nil
}

// cretas a URL that is broken
var brokenURL string = string([]byte{0x7f})

func TestGetAPIPriceDataSun(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	// Should not be an array, it should match the exact struct
	sunDataExample = fmt.Sprintf(`{
		"results": {
			"date": "%d-%02d-%02d",
			"sunrise": "08:00:00",
			"sunset": "20:00:00",
			"first_light": "07:00:00",
			"last_light": "21:00:00",
			"dawn": "07:30:00",
			"dusk": "20:30:00",
			"solar_noon": "16:00:00",
			"golden_hour": "19:00:00",
			"day_length": "12:00:00",
			"timezone": "CET",
			"utc_offset": 1
		},
		"status": "OK"
		}`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(),
	)
	// creates a fake response
	fakeBody := fmt.Sprintf(sunDataExample)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	// Testing good cases
	// Test case: goal is no errors
	apiURL := fmt.Sprintf(`http://api.sunrisesunset.io/json?lat=%06f&lng=%06f&timezone=CET&date=%d-%02d-%02d&time_format=24`, ua.Latitude, ua.Longitude, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day())
	fmt.Println("API URL:", apiURL)

	// creates a mock HTTP transport to simulate api response for the test
	newMockTransport(resp)
	err := ua.getAPIData(apiURL)
	if err != nil {
		t.Errorf("expected no errors but got %s :", err)
	}
	// Testing bad cases
	// Test case: using wrong url leads to an error
	newMockTransport(resp)
	// Call the function (which now hits the mock server)
	err = ua.getAPIData(brokenURL)
	if err == nil {
		t.Errorf("Expected an error but got none!")
	}
	// Test case: if reading the body causes an error
	resp.Body = errReader(0)
	newMockTransport(resp)
	err = ua.getAPIData(apiURL)
	if err != errBodyRead {
		t.Errorf("expected an error %v, got %v", errBodyRead, err)
	}
	//Test case: if status code > 299
	resp.Body = io.NopCloser(strings.NewReader(fakeBody))
	resp.StatusCode = 300
	newMockTransport(resp)
	err = ua.getAPIData(apiURL)
	// check the statuscode is bad, witch is expected for the test to be successful
	if err != errStatuscode {
		t.Errorf("expected an bad status code but got %v", err)
	}
}
