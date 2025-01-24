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

const thermostatDomain string = "http://localhost:port/api/apikey/sensors/thermostat_index/config"
const plugDomain string = "http://localhost:port/api/apikey/lights/plug_index/config"

func TestUnitAssetChanged(t *testing.T) {

	// Don't understand how to check my own deConz API calls, will extend the test with this once i understand
	trans := newMockTransport()

	// Create a form
	f := forms.SignalA_v1a{
		Value: 27.0,
	}

	// Creates a single UnitAsset and assert it changes
	ua := UnitAsset{
		Setpt: 20.0,
	}

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
