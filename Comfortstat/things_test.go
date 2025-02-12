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
	resp *http.Response
	hits map[string]int
}

func newMockTransport(resp *http.Response) mockTransport {
	t := mockTransport{
		resp: resp,
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

// price example string in a JSON-like format
var priceExample string = fmt.Sprintf(`[{
		"SEK_per_kWh": 0.26673,
		"EUR_per_kWh": 0.02328,
		"EXR": 11.457574,
		"time_start": "%d-%02d-%02dT%02d:00:00+01:00",
		"time_end": "2025-01-06T04:00:00+01:00"
		}]`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(), time.Now().Local().Hour(),
)

// RoundTrip method is required to fulfil the RoundTripper interface (as required by the DefaultClient).
// It prevents the request from being sent over the network and count how many times
// a domain was requested.
func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.hits[req.URL.Hostname()] += 1
	t.resp.Request = req
	return t.resp, nil
}

// //////////////////////////////////////////////////////////////////////////////
const apiDomain string = "www.elprisetjustnu.se"

func TestAPIDataFetchPeriod(t *testing.T) {
	want := 3600
	if apiFetchPeriod < want {
		t.Errorf("expected API fetch period >= %d, got %d", want, apiFetchPeriod)
	}
}

func TestSingleUnitAssetOneAPICall(t *testing.T) {
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		//Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	trans := newMockTransport(resp)
	// Creates a single UnitAsset and assert it only sends a single API request
	ua := initTemplate().(*UnitAsset)
	//retrieveAPIPrice(ua)
	ua.getSEKPrice()

	// TEST CASE: cause a single API request
	hits := trans.domainHits(apiDomain)
	if hits > 1 {
		t.Errorf("expected number of api requests = 1, got %d requests", hits)
	}
}

func TestMultipleUnitAssetOneAPICall(t *testing.T) {
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
	}
	trans := newMockTransport(resp)
	// Creates multiple UnitAssets and monitor their API requests
	units := 10
	for i := 0; i < units; i++ {
		ua := initTemplate().(*UnitAsset)
		//retrieveAPIPrice(ua)
		ua.getSEKPrice()
	}
	// TEST CASE: causing only one API hit while using multiple UnitAssets
	hits := trans.domainHits(apiDomain)
	if hits > 1 {
		t.Errorf("expected number of api requests = 1, got %d requests (from %d units)", hits, units)
	}
}

func TestSetmethods(t *testing.T) {
	asset := initTemplate().(*UnitAsset)

	// Simulate the input signals
	MinTempInputSignal := forms.SignalA_v1a{
		Value: 1.0,
	}
	MaxTempInputSignal := forms.SignalA_v1a{
		Value: 29.0,
	}
	MinPriceInputSignal := forms.SignalA_v1a{
		Value: 2.0,
	}
	MaxPriceInputSignal := forms.SignalA_v1a{
		Value: 12.0,
	}
	DesTempInputSignal := forms.SignalA_v1a{
		Value: 23.7,
	}
	//call and test MinTemp
	asset.setMinTemp(MinTempInputSignal)
	if asset.MinTemp != 1.0 {
		t.Errorf("expected MinTemp to be 1.0, got %f", asset.MinTemp)
	}
	// call and test MaxTemp
	asset.setMaxTemp(MaxTempInputSignal)
	if asset.MaxTemp != 29.0 {
		t.Errorf("expected MaxTemp to be 25.0, got %f", asset.MaxTemp)
	}
	//call and test MinPrice
	asset.setMinPrice(MinPriceInputSignal)
	if asset.MinPrice != 2.0 {
		t.Errorf("expected MinPrice to be 2.0, got %f", asset.MinPrice)
	}
	//call and test MaxPrice
	asset.setMaxPrice(MaxPriceInputSignal)
	if asset.MaxPrice != 12.0 {
		t.Errorf("expected MaxPrice to be 12.0, got %f", asset.MaxPrice)
	}
	// call and test DesiredTemp
	asset.setDesiredTemp(DesTempInputSignal)
	if asset.DesiredTemp != 23.7 {
		t.Errorf("expected Desierd temprature is to be 23.7, got %f", asset.DesiredTemp)
	}
}

func TestGetMethods(t *testing.T) {
	uasset := initTemplate().(*UnitAsset)

	////MinTemp////
	// check if the value from the struct is the acctual value that the func is getting
	result := uasset.getMinTemp()
	if result.Value != uasset.MinTemp {
		t.Errorf("expected Value of the MinTemp is to be %v, got %v", uasset.MinTemp, result.Value)
	}
	//check that the Unit is correct
	if result.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result.Unit)
	}
	////MaxTemp////
	result2 := uasset.getMaxTemp()
	if result2.Value != uasset.MaxTemp {
		t.Errorf("expected Value of the MaxTemp is to be %v, got %v", uasset.MaxTemp, result2.Value)
	}
	//check that the Unit is correct
	if result2.Unit != "Celsius" {
		t.Errorf("expected Unit of the MaxTemp is to be 'Celsius', got %v", result2.Unit)
	}
	////MinPrice////
	// check if the value from the struct is the acctual value that the func is getting
	result3 := uasset.getMinPrice()
	if result3.Value != uasset.MinPrice {
		t.Errorf("expected Value of the minPrice is to be %v, got %v", uasset.MinPrice, result3.Value)
	}
	//check that the Unit is correct
	if result3.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result3.Unit)
	}
	////MaxPrice////
	// check if the value from the struct is the acctual value that the func is getting
	result4 := uasset.getMaxPrice()
	if result4.Value != uasset.MaxPrice {
		t.Errorf("expected Value of the maxPrice is  to be %v, got %v", uasset.MaxPrice, result4.Value)
	}
	//check that the Unit is correct
	if result4.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result4.Unit)
	}
	////DesierdTemp////
	// check if the value from the struct is the acctual value that the func is getting
	result5 := uasset.getDesiredTemp()
	if result5.Value != uasset.DesiredTemp {
		t.Errorf("expected desired temprature is to be %v, got %v", uasset.DesiredTemp, result5.Value)
	}
	//check that the Unit is correct
	if result5.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result5.Unit)
	}
	////SEKPrice////
	result6 := uasset.getSEKPrice()
	if result6.Value != uasset.SEKPrice {
		t.Errorf("expected electric price is to be %v, got %v", uasset.SEKPrice, result6.Value)
	}
}

func TestInitTemplate(t *testing.T) {
	uasset := initTemplate().(*UnitAsset)

	//// unnecessary test, but good for practicing
	name := uasset.GetName()
	if name != "Set Values" {
		t.Errorf("expected name of the resource is %v, got %v", uasset.Name, name)
	}
	Services := uasset.GetServices()
	if Services == nil {
		t.Fatalf("If Services is nil, not worth to continue testing")
	}
	//Services//
	if Services["SEKPrice"].Definition != "SEKPrice" {
		t.Errorf("expected service defenition to be SEKprice")
	}
	if Services["MaxTemperature"].Definition != "MaxTemperature" {
		t.Errorf("expected service defenition to be MaxTemperature")
	}
	if Services["MinTemperature"].Definition != "MinTemperature" {
		t.Errorf("expected service defenition to be MinTemperature")
	}
	if Services["MaxPrice"].Definition != "MaxPrice" {
		t.Errorf("expected service defenition to be MaxPrice")
	}
	if Services["MinPrice"].Definition != "MinPrice" {
		t.Errorf("expected service defenition to be MinPrice")
	}
	if Services["DesiredTemp"].Definition != "DesiredTemp" {
		t.Errorf("expected service defenition to be DesiredTemp")
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
	sys := components.NewSystem("Comfortstat", ctx)

	// Instatiate the Capusle
	sys.Husk = &components.Husk{
		Description: " is a controller for a consumed servo motor position based on a consumed temperature",
		Certificate: "ABCD",
		Details:     map[string][]string{"Developer": {"Arrowhead"}},
		ProtoPort:   map[string]int{"https": 0, "http": 8670, "coap": 0},
		InfoLink:    "https://github.com/lmas/d0020e_code/tree/master/Comfortstat",
	}
	setSEKPrice := components.Service{
		Definition:  "SEKPrice",
		SubPath:     "SEKPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the current electric hourly price (using a GET request)",
	}
	setMaxTemp := components.Service{
		Definition:  "MaxTemperature",
		SubPath:     "MaxTemperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the maximum temp the user wants (using a GET request)",
	}
	setMinTemp := components.Service{
		Definition:  "MinTemperature",
		SubPath:     "MinTemperature",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum temp the user could tolerate (using a GET request)",
	}
	setMaxPrice := components.Service{
		Definition:  "MaxPrice",
		SubPath:     "MaxPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the maximum price the user wants to pay (using a GET request)",
	}
	setMinPrice := components.Service{
		Definition:  "MinPrice",
		SubPath:     "MinPrice",
		Details:     map[string][]string{"Unit": {"SEK"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the minimum price the user wants to pay (using a GET request)",
	}
	setDesiredTemp := components.Service{
		Definition:  "DesiredTemp",
		SubPath:     "DesiredTemp",
		Details:     map[string][]string{"Unit": {"Celsius"}, "Forms": {"SignalA_v1a"}},
		Description: "provides the desired temperature the system calculates based on user inputs (using a GET request)",
	}
	// new Unitasset struct init.
	uac := UnitAsset{
		//These fields should reflect a unique asset (ie, a single sensor with unique ID and location)
		Name:        "Set Values",
		Details:     map[string][]string{"Location": {"Kitchen"}},
		SEKPrice:    1.5,  // Example electricity price in SEK per kWh
		MinPrice:    1.0,  // Minimum price allowed
		MaxPrice:    2.0,  // Maximum price allowed
		MinTemp:     20.0, // Minimum temperature
		MaxTemp:     25.0, // Maximum temprature allowed
		DesiredTemp: 0,    // Desired temp calculated by system
		Period:      15,

		// maps the provided services from above
		ServicesMap: components.Services{
			setMaxTemp.SubPath:     &setMaxTemp,
			setMinTemp.SubPath:     &setMinTemp,
			setMaxPrice.SubPath:    &setMaxPrice,
			setMinPrice.SubPath:    &setMinPrice,
			setSEKPrice.SubPath:    &setSEKPrice,
			setDesiredTemp.SubPath: &setDesiredTemp,
		},
	}

	ua, _ := newUnitAsset(uac, &sys, nil)
	// Calls the method that gets the name of the new unitasset.
	name := ua.GetName()
	if name != "Set Values" {
		t.Errorf("expected name to be Set values, but got: %v", name)
	}
}

// Test if the method calculateDesierdTemp() calculates a correct value
func TestCalculateDesiredTemp(t *testing.T) {
	var True_result float64 = 22.5
	asset := initTemplate().(*UnitAsset)
	// calls and saves the value
	result := asset.calculateDesiredTemp()
	// checks if actual calculated value matches the expexted value
	if result != True_result {
		t.Errorf("Expected calculated temp is %v, got %v", True_result, result)
	}
}

// This test catches the special cases, when the temprature is to be set to the minimum temprature right away
func TestSpecialCalculate(t *testing.T) {
	asset := UnitAsset{
		SEKPrice: 3.0,
		MaxPrice: 2.0,
		MinTemp:  17.0,
	}
	//call the method and save the result in a varable for testing
	result := asset.calculateDesiredTemp()
	//check the result from the call above
	if result != asset.MinTemp {
		t.Errorf("Expected temperature to be %v, got %v", asset.MinTemp, result)
	}
}

// Fuctions that help creating bad body
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

func TestGetAPIPriceData(t *testing.T) {
	// creating a price example, nessasry fore the test
	priceExample = fmt.Sprintf(`[{
		"SEK_per_kWh": 0.26673,
		"EUR_per_kWh": 0.02328,
		"EXR": 11.457574,
		"time_start": "%d-%02d-%02dT%02d:00:00+01:00",
		"time_end": "2025-01-06T04:00:00+01:00"
		}]`, time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(), time.Now().Local().Hour(),
	)
	// creates a fake response
	fakeBody := fmt.Sprintf(priceExample)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(fakeBody)),
	}
	// Testing good cases
	// Test case: goal is no errors
	url := fmt.Sprintf(
		`https://www.elprisetjustnu.se/api/v1/prices/%d/%02d-%02d_SE1.json`,
		time.Now().Local().Year(), int(time.Now().Local().Month()), time.Now().Local().Day(),
	)
	// creates a mock HTTP transport to simulate api respone for the test
	newMockTransport(resp)
	err := getAPIPriceData(url)
	if err != nil {
		t.Errorf("expected no errors but got %s :", err)
	}
	// Check if the correct price is stored
	//	expectedPrice := 0.26673
	//	if globalPrice.SEKPrice != expectedPrice {
	//		t.Errorf("Expected SEKPrice %f, but got %f", expectedPrice, globalPrice.SEKPrice)
	//	}
	// Testing bad cases
	// Test case: using wrong url leads to an error
	newMockTransport(resp)
	// Call the function (which now hits the mock server)
	err = getAPIPriceData(brokenURL)
	if err == nil {
		t.Errorf("Expected an error but got none!")
	}
	// Test case: if reading the body causes an error
	resp.Body = errReader(0)
	newMockTransport(resp)
	err = getAPIPriceData(url)
	if err != errBodyRead {
		t.Errorf("expected an error %v, got %v", errBodyRead, err)
	}
	//Test case: if status code > 299
	resp.Body = io.NopCloser(strings.NewReader(fakeBody))
	resp.StatusCode = 300
	newMockTransport(resp)
	err = getAPIPriceData(url)
	// check the statuscode is bad, witch is expected for the test to be successful
	if err != errStatuscode {
		t.Errorf("expected an bad status code but got %v", err)
	}
	// test case: if unmarshal a bad body creates a error
	resp.StatusCode = 200
	resp.Body = io.NopCloser(strings.NewReader(fakeBody + "123"))
	newMockTransport(resp)
	err = getAPIPriceData(url)
	// make the check if the unmarshal creats a error
	if err == nil {
		t.Errorf("expected an error, got %v :", err)
	}
}
