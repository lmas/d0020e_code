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

func TestSetmethods(t *testing.T) {

	asset := initTemplate().(*UnitAsset)

	// Simulate the input signals
	MinTemp_inputSignal := forms.SignalA_v1a{
		Value: 1.0,
	}
	MaxTemp_inputSignal := forms.SignalA_v1a{
		Value: 29.0,
	}
	MinPrice_inputSignal := forms.SignalA_v1a{
		Value: 2.0,
	}
	MaxPrice_inputSignal := forms.SignalA_v1a{
		Value: 12.0,
	}
	DesTemp_inputSignal := forms.SignalA_v1a{
		Value: 23.7,
	}

	// Call the setMin_temp function
	asset.setMin_temp(MinTemp_inputSignal)
	asset.setMax_temp(MaxTemp_inputSignal)
	asset.setMin_price(MinPrice_inputSignal)
	asset.setMax_price(MaxPrice_inputSignal)
	asset.setDesired_temp(DesTemp_inputSignal)

	// check if the temprature has changed correctly
	if asset.Min_temp != 1.0 {
		t.Errorf("expected Min_temp to be 1.0, got %f", asset.Min_temp)
	}
	if asset.Max_temp != 29.0 {
		t.Errorf("expected Max_temp to be 25.0, got %f", asset.Max_temp)
	}
	if asset.Min_price != 2.0 {
		t.Errorf("expected Min_Price to be 2.0, got %f", asset.Min_price)
	}
	if asset.Max_price != 12.0 {
		t.Errorf("expected Max_Price to be 12.0, got %f", asset.Max_price)
	}
	if asset.Desired_temp != 23.7 {
		t.Errorf("expected Desierd temprature is to be 23.7, got %f", asset.Desired_temp)
	}

}

func Test_GetMethods(t *testing.T) {

	uasset := initTemplate().(*UnitAsset)
	//call the fuctions
	result := uasset.getMin_temp()
	result2 := uasset.getMax_temp()
	result3 := uasset.getMin_price()
	result4 := uasset.getMax_price()
	result5 := uasset.getDesired_temp()
	result6 := uasset.getSEK_price()

	////MinTemp////
	// check if the value from the struct is the acctual value that the func is getting
	if result.Value != uasset.Min_temp {
		t.Errorf("expected Value of the min_temp is to be %v, got %v", uasset.Min_temp, result.Value)
	}
	//check that the Unit is correct
	if result.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result.Unit)
		////MaxTemp////
	}
	if result2.Value != uasset.Max_temp {
		t.Errorf("expected Value of the Max_temp is to be %v, got %v", uasset.Max_temp, result2.Value)
	}
	//check that the Unit is correct
	if result2.Unit != "Celsius" {
		t.Errorf("expected Unit of the Max_temp is to be 'Celsius', got %v", result2.Unit)
	}
	////MinPrice////
	// check if the value from the struct is the acctual value that the func is getting
	if result3.Value != uasset.Min_price {
		t.Errorf("expected Value of the minPrice is to be %v, got %v", uasset.Min_price, result3.Value)
	}
	//check that the Unit is correct
	if result3.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result3.Unit)
	}

	////MaxPrice////
	// check if the value from the struct is the acctual value that the func is getting
	if result4.Value != uasset.Max_price {
		t.Errorf("expected Value of the maxPrice is  to be %v, got %v", uasset.Max_price, result4.Value)
	}
	//check that the Unit is correct
	if result4.Unit != "SEK" {
		t.Errorf("expected Unit to be 'SEK', got %v", result4.Unit)
	}
	////DesierdTemp////
	// check if the value from the struct is the acctual value that the func is getting
	if result5.Value != uasset.Desired_temp {
		t.Errorf("expected desired temprature is to be %v, got %v", uasset.Desired_temp, result5.Value)
	}
	//check that the Unit is correct
	if result5.Unit != "Celsius" {
		t.Errorf("expected Unit to be 'Celsius', got %v", result5.Unit)
	}
	////SEK_Price////
	if result6.Value != uasset.SEK_price {
		t.Errorf("expected electric price is to be %v, got %v", uasset.SEK_price, result6.Value)
	}
	//check that the Unit is correct
	//if result5.Unit != "SEK" {
	//	t.Errorf("expected Unit to be 'SEK', got %v", result6.Unit)
	//}

}

func Test_initTemplet(t *testing.T) {
	uasset := initTemplate().(*UnitAsset)

	name := uasset.GetName()
	Services := uasset.GetServices()
	//Cervices := uasset.GetCervices()
	Details := uasset.GetDetails()

	//// unnecessary test, but good for practicing
	if name != "Set Values" {
		t.Errorf("expected name of the resource is %v, got %v", uasset.Name, name)
	}
	if Services == nil {
		t.Fatalf("If Services is nil, not worth to continue testing")
	}
	////Services////
	if Services["SEK_price"].Definition != "SEK_price" {
		t.Errorf("expected service defenition to be SEKprice")
	}
	if Services["max_temperature"].Definition != "max_temperature" {
		t.Errorf("expected service defenition to be max_temperature")
	}
	if Services["min_temperature"].Definition != "min_temperature" {
		t.Errorf("expected service defenition to be min_temperature")
	}
	if Services["max_price"].Definition != "max_price" {
		t.Errorf("expected service defenition to be max_price")
	}
	if Services["min_price"].Definition != "min_price" {
		t.Errorf("expected service defenition to be min_price")
	}
	if Services["desired_temp"].Definition != "desired_temp" {
		t.Errorf("expected service defenition to be desired_temp")
	}
	//// Testing GetCervice
	//if Cervices == nil {
	//	t.Fatalf("If cervises is nil, not worth to continue testing")
	//}
	//// Testing Details
	if Details == nil {
		t.Errorf("expected a map, but Details was nil, ")
	}

}
