package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/forms"
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

const apiDomain string = "www.elprisetjustnu.se"

func TestAPIDataFetchPeriod(t *testing.T) {
	want := 3600
	if apiFetchPeriod < want {
		t.Errorf("expected API fetch period >= %d, got %d", want, apiFetchPeriod)
	}
}

func TestSingleUnitAssetOneAPICall(t *testing.T) {
	trans := newMockTransport()
	// Creates a single UnitAsset and assert it only sends a single API request
	ua := initTemplate().(*UnitAsset)
	retrieveAPI_price(ua)

	// TEST CASE: cause a single API request
	hits := trans.domainHits(apiDomain)
	if hits > 1 {
		t.Errorf("expected number of api requests = 1, got %d requests", hits)
	}

	// TODO: try more test cases!
}

func TestMultipleUnitAssetOneAPICall(t *testing.T) {
	trans := newMockTransport()
	// Creates multiple UnitAssets and monitor their API requests
	units := 10
	for i := 0; i < units; i++ {
		ua := initTemplate().(*UnitAsset)
		retrieveAPI_price(ua)
	}

	// TEST CASE: causing only one API hit while using multiple UnitAssets
	hits := trans.domainHits(apiDomain)
	if hits > 1 {
		t.Errorf("expected number of api requests = 1, got %d requests (from %d units)", hits, units)
	}

	// TODO: more test cases??
}

func Test_structupdate_minTemp(t *testing.T) {

	asset := UnitAsset{
		Min_temp:  20.0,
		Max_temp:  30.0,
		Max_price: 10.0,
		Min_price: 5.0,
		SEK_price: 7.0,
	}
	// Simulate the input signal
	Min_inputSignal := forms.SignalA_v1a{
		Value: 1.0,
		
	}
	// Call the setMin_temp function
	asset.setMin_temp(Min_inputSignal)

	// check if the temprature has changed correctly
	if asset.Min_temp != 1.0 {
		t.Errorf("expected Min_temp to be 1.0, got %f", asset.Min_temp)
	}

}

func Test_GetTemprature(t *testing.T) {
	expectedminTemp := 25.0
	expectedmaxTemp := 30.0
	expectedminPrice := 1.0
	expectedmaxPrice := 5.0
	expectedDesiredTemp := 22.5

	uasset := UnitAsset{
		Min_temp:     expectedminTemp,
		Max_temp:     expectedmaxTemp,
		Min_price:    expectedminPrice,
		Max_price:    expectedmaxPrice,
		Desired_temp: expectedDesiredTemp,
	}
	//call the fuctions
	result := uasset.getMin_temp()
	result2 := uasset.getMax_temp()
	result3 := uasset.getMin_price()
	result4 := uasset.getMax_price()
	result5 := uasset.getDesired_temp()

	////MinTemp////
	// check if the value from the struct is the acctual value that the func is getting
	if result.Value != expectedminTemp {
		t.Errorf("expected Value to be %v, got %v", expectedminTemp, result.Value)
	}
	//check that the Unit is correct
	if result.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result.Unit)
		////MaxTemp////
	}
	if result2.Value != expectedmaxTemp {
		t.Errorf("expected Value of the Min_temp is to be %v, got %v", expectedmaxTemp, result2.Value)
	}
	//check that the Unit is correct
	if result2.Unit != "Celsius" {
		t.Errorf("expected Unit of the Max_temp is to be 'Celsius', got %v", result2.Unit)
	}
	////MinPrice////
	// check if the value from the struct is the acctual value that the func is getting
	if result3.Value != expectedminPrice {
		t.Errorf("expected Value of the maxPrice is to be %v, got %v", expectedminPrice, result3.Value)
	}
	//check that the Unit is correct
	if result3.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result3.Unit)
	}

	////MaxPrice////
	// check if the value from the struct is the acctual value that the func is getting
	if result4.Value != expectedmaxPrice {
		t.Errorf("expected Value of the maxPrice is  to be %v, got %v", expectedmaxPrice, result4.Value)
	}
	//check that the Unit is correct
	if result4.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result4.Unit)
	}
	////DesierdTemp////
	// check if the value from the struct is the acctual value that the func is getting
	if result5.Value != expectedDesiredTemp {
		t.Errorf("expected desired temprature is to be %v, got %v", expectedDesiredTemp, result5.Value)
	}
	//check that the Unit is correct
	if result5.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result5.Unit)
	}

}

func Test_structupdate_maxTemp(t *testing.T) {

	asset := &UnitAsset{
		Min_temp:  20.0,
		Max_temp:  30.0,
		Max_price: 10.0,
		Min_price: 5.0,
		SEK_price: 7.0,
	}
	// Simulate the input signal
	Max_inputSignal := forms.SignalA_v1a{
		Value: 21.0,
	}
	// Call the setMin_temp function
	asset.setMax_temp(Max_inputSignal)

	// check if the temprature has changed correctly
	if asset.Max_temp != 21.0 {
		t.Errorf("expected Max_temp to be 21.0, got %f", asset.Max_temp)
	}

}
