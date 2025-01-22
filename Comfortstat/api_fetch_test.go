package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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
