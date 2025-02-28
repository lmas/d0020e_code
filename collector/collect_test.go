package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/components"
	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

type mockTransport struct {
	oldTrans http.RoundTripper
	respCode int
	respBody io.ReadCloser
}

func newMockTransport() (trans mockTransport, restore func()) {
	trans = mockTransport{
		oldTrans: http.DefaultClient.Transport,
		respCode: 200,
		respBody: io.NopCloser(strings.NewReader("")),
	}
	restore = func() {
		// Use this func to restore the default value
		http.DefaultClient.Transport = trans.oldTrans
	}
	// Hijack the default http client so no actual http requests are sent over the network
	http.DefaultClient.Transport = trans
	return
}

func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	return &http.Response{
		Request:    req,
		StatusCode: t.respCode,
		Body:       t.respBody,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

var mockStates = map[string]string{
	"temperature": `{ "value": 0, "unit": "Celcius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"SEKPrice":    `{ "value": 0.10403, "unit": "SEK", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"DesiredTemp": `{ "value": 25, "unit": "Celsius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"setpoint":    `{ "value": 20, "unit": "Celsius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"consumption": `{ "value": 32, "unit": "Wh", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"state":       `{ "value": 1, "unit": "Binary", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"power":       `{ "value": 330, "unit": "Wh", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"current":     `{ "value": 9, "unit": "mA", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"voltage":     `{ "value": 229, "unit": "V", "timestamp": "%s", "version": "SignalA_v1.0" }`,
}

const (
	mockBodyType string = "application/json"

	mockStateIncomplete string = `{ "value": 20, "timestamp": "%s" }`
	mockStateBadVersion string = `{ "value": false, "timestamp": "%s", "version": "SignalB_v1.0" }`
)

func mockGetState(c *components.Cervice, s *components.System) (f forms.Form, err error) {
	if c == nil {
		err = fmt.Errorf("got empty *Cervice instance")
		return
	}
	b := mockStates[c.Name]
	if len(b) < 1 {
		err = fmt.Errorf("found no mock body for service: %s", c.Name)
		return
	}
	body := fmt.Sprintf(b, time.Now().Format(time.RFC3339))
	f, err = usecases.Unpack([]byte(body), mockBodyType)
	if err != nil {
		err = fmt.Errorf("failed to unpack mock body: %s", err)
	}
	return
}

func TestCollectService(t *testing.T) {
	_, restore := newMockTransport()
	ua := newUnitAsset(*initTemplate(), newSystem())
	defer func() {
		// Make sure to run cleanups! Otherwise you'll get leftover errors from influx
		ua.cleanup()
		restore()
	}()
	ua.apiGetState = mockGetState
	sample := Sample{"setpoint", map[string][]string{"Location": {"Kitchen"}}}

	// Good case
	err := ua.collectService(sample)
	if err != nil {
		t.Fatalf("Expected nil error, got: %s", err)
	}
	good := mockStates["setpoint"]

	// Bad case: a service returns incomplete data
	mockStates["setpoint"] = mockStateIncomplete
	err = ua.collectService(sample)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Bad case: a service returns bad form version
	mockStates["setpoint"] = mockStateBadVersion
	err = ua.collectService(sample)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// WARN: Don't forget to restore the mocks!
	mockStates["setpoint"] = good
}

func TestCollectAllServices(t *testing.T) {
	_, restore := newMockTransport()
	ua := newUnitAsset(*initTemplate(), newSystem())
	defer func() {
		ua.cleanup()
		restore()
	}()
	ua.apiGetState = mockGetState

	// Good case
	err := ua.collectAllServices()
	if err != nil {
		t.Fatalf("Expected nil error, got: %s", err)
	}
	good := mockStates["setpoint"]

	// Bad case: a service returns incomplete data
	mockStates["setpoint"] = mockStateIncomplete
	err = ua.collectAllServices()
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// WARN: Don't forget to restore the mocks!
	mockStates["setpoint"] = good
}
