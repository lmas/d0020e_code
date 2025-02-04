package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetpt(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	// Good case test: GET
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8670/ZigBee/Template/setpoint", nil)
	good_code := 200
	ua.setpt(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)

	value := strings.Contains(string(stringBody), `"value": 20`)
	unit := strings.Contains(string(stringBody), `"unit": "Celsius"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_code {
		t.Errorf("Good GET: Expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	if value != true {
		t.Errorf("Good GET: The value statment should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statment to be true!")
	}
	// Bad test case: Default part of code (faulty http method)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8670/ZigBee/Template/setpoint", nil)
	ua.setpt(w, r)
	// Read response and check statuscode, expecting 404 (StatusNotFound)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}

	// Bad PUT (Cant reach ZigBee)
	w = httptest.NewRecorder()
	// Make the body
	fakebody := string(`{"value": 24, "version": "SignalA_v1.0"}`)
	sentBody := io.NopCloser(strings.NewReader(fakebody))
	// Send the request
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBee/Template/setpoint", sentBody)
	ua.setpt(w, r)
	resp = w.Result()
	good_code = 200
	// Check for errors, should not be 200
	if resp.StatusCode == good_code {
		t.Errorf("Bad PUT: Expected bad status code: got %v", resp.StatusCode)
	}

	// Bad test case: PUT Failing @ HTTPProcessSetRequest
	w = httptest.NewRecorder()
	// Make the body
	fakebody = string(`{"value": "24", "version": "SignalA_v1.0"}`) // MISSING VERSION IN SENTBODY
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	// Send the request
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBee/Template/setpoint", sentBody)
	ua.setpt(w, r)
	resp = w.Result()
	// Check for errors
	if resp.StatusCode == good_code {
		t.Errorf("Bad PUT: Expected an error during HTTPProcessSetRequest")
	}

	// Good test case: PUT
	w = httptest.NewRecorder()
	// Make the body and request
	fakebody = string(`{"value": 24, "version": "SignalA_v1.0"}`)
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBee/Template/setpoint", sentBody)
	// Mock the http response/traffic to zigbee
	zBeeResponse := `[{"success":{"/sensors/7/config/heatsetpoint":2400}}]`
	resp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	newMockTransport(resp, false, nil)
	// Set the response body to same as mock response
	w.Body = bytes.NewBuffer([]byte(zBeeResponse))
	// Send the request
	ua.setpt(w, r)
	resp = w.Result()
	// Check for errors
	if resp.StatusCode != good_code {
		t.Errorf("Good PUT: Expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	// Convert body to a string and check that it's correct
	respBodyBytes, _ := io.ReadAll(resp.Body)
	respBody := string(respBodyBytes)
	if respBody != `[{"success":{"/sensors/7/config/heatsetpoint":2400}}]` {
		t.Errorf("Wrong body")
	}
}
