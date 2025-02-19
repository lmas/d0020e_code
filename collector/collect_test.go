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
	respCode int
	respBody io.ReadCloser

	// hits     map[string]int
	// returnError bool
	// resp        *http.Response
	// err         error
}

func newMockTransport() mockTransport {
	t := mockTransport{
		respCode: 200,
		respBody: io.NopCloser(strings.NewReader("")),

		// hits: make(map[string]int),
		// err:         err,
		// returnError: retErr,
		// resp:        resp,
	}
	// Hijack the default http client so no actual http requests are sent over the network
	http.DefaultClient.Transport = t
	return t
}

func (t mockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	// log.Println("HIJACK:", req.URL.String())
	// t.hits[req.URL.Hostname()] += 1
	// if t.err != nil {
	// 	return nil, t.err
	// }
	// if t.returnError != false {
	// 	req.GetBody = func() (io.ReadCloser, error) {
	// 		return nil, errHTTP
	// 	}
	// }
	// t.resp.Request = req
	// return t.resp, nil

	// b, err := io.ReadAll(req.Body)
	// if err != nil {
	// 	return
	// }
	// fmt.Println(string(b))

	return &http.Response{
		Request:    req,
		StatusCode: t.respCode,
		Body:       t.respBody,
	}, nil
}

const mockBodyType string = "application/json"

var mockStates = map[string]string{
	"temperature": `{ "value": 0, "unit": "Celcius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"SEKPrice":    `{ "value": 0.10403, "unit": "SEK", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"DesiredTemp": `{ "value": 25, "unit": "Celsius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
	"setpoint":    `{ "value": 20, "unit": "Celsius", "timestamp": "%s", "version": "SignalA_v1.0" }`,
}

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
	newMockTransport()
	ua := newUnitAsset(*initTemplate(), newSystem())
	ua.apiGetState = mockGetState

	// for _, service := range consumeServices {
	// 	err := ua.collectService(service)
	// 	if err != nil {
	// 		t.Fatalf("Expected nil error while pulling %s, got: %s", service, err)
	// 	}
	// }
	err := ua.collectAllServices()
	if err != nil {
		t.Fatalf("Expected nil error, got: %s", err)
	}
}
