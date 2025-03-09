package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var good_code = 200

func TestSetpt(t *testing.T) {
	// --- ZHAThermostat ---
	ua := initTemplate().(*UnitAsset)

	// --- Good case test: GET ---
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.setpt(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// Check if correct values are present in the body, each line returns true/false
	value := strings.Contains(string(stringBody), `"value": 20`)
	unit := strings.Contains(string(stringBody), `"unit": "Celsius"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)
	// Check that above statements are true
	if value != true {
		t.Errorf("Good GET: The value statement should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statement to be true!")
	}
	// --- Good test case: not correct device type
	ua.Model = "Wrong Device"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", nil)
	ua.setpt(w, r)
	// Read response and check statuscode
	resp = w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected the status to be 500 but got: %v", resp.StatusCode)
	}

	// --- Default part of code (faulty http method) ---
	ua = initTemplate().(*UnitAsset)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.setpt(w, r)
	// Read response and check statuscode, expecting 404 (StatusNotFound)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}

	// --- Good test case: PUT ---
	w = httptest.NewRecorder()
	// Make the body and request
	fakebody := string(`{"value": 24, "version": "SignalA_v1.0"}`)
	sentBody := io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", sentBody)
	r.Header.Set("Content-Type", "application/json")
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

	// --- Bad test case: PUT Failing @ HTTPProcessSetRequest ---
	w = httptest.NewRecorder()
	// Make the body
	fakebody = string(`{"value": "24"`) // MISSING VERSION IN SENTBODY
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	// Send the request
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.setpt(w, r)
	resp = w.Result()
	// Check for errors
	if resp.StatusCode == good_code {
		t.Errorf("Bad PUT: Expected an error during HTTPProcessSetRequest")
	}

	// --- Bad PUT (Cant reach ZigBee) ---
	w = httptest.NewRecorder()
	ua.Model = "Wrong device"
	// Make the body
	fakebody = string(`{"value": 24, "version": "SignalA_v1.0"}`)
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	// Send the request
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartThermostat1/setpoint", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.setpt(w, r)
	resp = w.Result()
	// Check for errors, should not be 200
	if resp.StatusCode == good_code {
		t.Errorf("Bad PUT: Expected bad status code: got %v.", resp.StatusCode)
	}
}

func TestConsumption(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Slaves["ZHAConsumption"] = "14:ef:14:10:00:b3:b3:89-01"
	// --- Good case test: GET ---
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)

	zBeeResponse := `{
			"state": {"consumption": 1},
			"name": "SmartPlug1",
			"uniqueid": "14:ef:14:10:00:b3:b3:89-XX-XXXX",
			"type": "ZHAConsumption"
			}`

	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	newMockTransport(zResp, false, nil)
	ua.consumption(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// Check if correct values are present in the body, each line returns true/false
	value := strings.Contains(string(stringBody), `"value": 1`)
	unit := strings.Contains(string(stringBody), `"unit": "Wh"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)

	// Check that above statements are true
	if value != true {
		t.Errorf("Good GET: The value statement should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statement to be true!")
	}
	// --- Wrong model ---
	ua.Model = "Wrong model"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)
	newMockTransport(zResp, false, nil)
	ua.consumption(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected statuscode 500, got: %d", resp.StatusCode)
	}
	// --- Bad test case: error from getConsumption() because of broken body ---
	ua.Model = "Smart plug"
	zBeeResponse = `{
			"state": {"consumption": 1},
			"name": "SnartPlug1",
			"uniqueid": "ConsumptionTest",
			"type": "ZHAConsumption"
			} + 123`

	zResp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)
	newMockTransport(zResp, false, nil)
	ua.consumption(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}
	// --- Default part of code (Method not supported)
	ua.Model = "Smart plug"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)
	ua.consumption(w, r)
	resp = w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected statuscode to be 404, got %d", resp.StatusCode)
	}
}

func TestPower(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Slaves["ZHAPower"] = "14:ef:14:10:00:b3:b3:89-01"
	// --- Good case test: GET ---
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/power", nil)

	zBeeResponse := `{
			"state": {"power": 2},
			"name": "SmartPlug1",
			"uniqueid": "14:ef:14:10:00:b3:b3:89-XX-XXXX",
			"type": "ZHAPower"
			}`

	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	newMockTransport(zResp, false, nil)
	ua.power(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// Check if correct values are present in the body, each line returns true/false
	value := strings.Contains(string(stringBody), `"value": 2`)
	unit := strings.Contains(string(stringBody), `"unit": "W"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)

	// Check that above statements are true
	if value != true {
		t.Errorf("Good GET: The value statement should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statement to be true!")
	}

	// --- Wrong model ---
	ua.Model = "Wrong model"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)
	newMockTransport(zResp, false, nil)
	ua.power(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected statuscode 500, got: %d", resp.StatusCode)
	}

	// --- Bad test case: error from getPower() because of broken body ---
	ua.Model = "Smart plug"
	zBeeResponse = `{
			"state": {"consumption": 1},
			"name": "SnartPlug1",
			"uniqueid": "ConsumptionTest",
			"type": "ZHAConsumption"
			} + 123`

	zResp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/power", nil)
	newMockTransport(zResp, false, nil)
	ua.power(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	// --- Default part of code (Method not supported)
	ua.Model = "Smart plug"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartPlug1/power", nil)
	ua.power(w, r)
	resp = w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected statuscode to be 404, got %d", resp.StatusCode)
	}
}

func TestCurrent(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Slaves["ZHAPower"] = "14:ef:14:10:00:b3:b3:89-01"
	// --- Good case test: GET ---
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/current", nil)

	zBeeResponse := `{
	"state": {"current": 3},
	"name": "SmartPlug1",
	"uniqueid": "14:ef:14:10:00:b3:b3:89-XX-XXXX",
	"type": "ZHAPower"
	}`

	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	newMockTransport(zResp, false, nil)
	ua.current(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// Check if correct values are present in the body, each line returns true/false
	value := strings.Contains(string(stringBody), `"value": 3`)
	unit := strings.Contains(string(stringBody), `"unit": "mA"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)

	// Check that above statements are true
	if value != true {
		t.Errorf("Good GET: The value statement should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statement to be true!")
	}

	// --- Wrong model ---
	ua.Model = "Wrong model"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/consumption", nil)
	newMockTransport(zResp, false, nil)
	ua.current(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected statuscode 500, got: %d", resp.StatusCode)
	}

	// --- Bad test case: error from getPower() because of broken body ---
	ua.Model = "Smart plug"
	zBeeResponse = `{
			"state": {"consumption": 1},
			"name": "SnartPlug1",
			"uniqueid": "ConsumptionTest",
			"type": "ZHAConsumption"
			} + 123`

	zResp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/power", nil)
	newMockTransport(zResp, false, nil)
	ua.current(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	// --- Default part of code (Method not supported)
	ua.Model = "Smart plug"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartPlug1/power", nil)
	ua.current(w, r)
	resp = w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected statuscode to be 404, got %d", resp.StatusCode)
	}
}

func TestVoltage(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"
	ua.Slaves["ZHAPower"] = "14:ef:14:10:00:b3:b3:89-01"
	// --- Good case test: GET ---
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/voltage", nil)
	zBeeResponse := `{
	"state": {"voltage": 4},
	"name": "SmartPlug1",
	"uniqueid": "14:ef:14:10:00:b3:b3:89-XX-XXXX",
	"type": "ZHAPower"
	}`
	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	newMockTransport(zResp, false, nil)
	ua.voltage(w, r)
	// Read response to a string, and save it in stringBody
	resp := w.Result()
	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// Check if correct values are present in the body, each line returns true/false
	value := strings.Contains(string(stringBody), `"value": 4`)
	unit := strings.Contains(string(stringBody), `"unit": "V"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)
	// Check that above statements are true
	if value != true {
		t.Errorf("Good GET: The value statement should be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statement to be true!")
	}

	// --- Wrong model ---
	ua.Model = "Wrong model"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/voltage", nil)
	newMockTransport(zResp, false, nil)
	ua.voltage(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected statuscode 500, got: %d", resp.StatusCode)
	}

	// --- Bad test case: error from getPower() because of broken body ---
	ua.Model = "Smart plug"
	zBeeResponse = `{
			"state": {"consumption": 1},
			"name": "SmartPlug1",
			"uniqueid": "ConsumptionTest",
			"type": "ZHAConsumption"
			} + 123`
	zResp = &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/voltage", nil)
	newMockTransport(zResp, false, nil)
	ua.voltage(w, r)
	resp = w.Result()
	if resp.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", resp.StatusCode)
	}

	// --- Default part of code (Method not supported)
	ua.Model = "Smart plug"
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartPlug1/voltage", nil)
	ua.voltage(w, r)
	resp = w.Result()
	if resp.StatusCode != 404 {
		t.Errorf("Expected statuscode to be 404, got %d", resp.StatusCode)
	}

}

func TestState(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	ua.Name = "SmartPlug1"
	ua.Model = "Smart plug"

	zBeeResponse := `{
		"state": {"on": true},
		"name": "SmartPlug1",
		"uniqueid": "14:ef:14:10:00:b3:b3:89-XX-XXXX",
		"type": "ZHAPower"
		}`
	zResp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(zBeeResponse)),
	}
	// --- Default part of code ---
	newMockTransport(zResp, false, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("123", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.state(w, r)
	res := w.Result()
	_, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error("Expected no errors")
	}
	if res.StatusCode != 404 {
		t.Errorf("Expected no errors in default part of code, got: %d", res.StatusCode)
	}

	// --- Good test case: GET ---
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.state(w, r)
	res = w.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error("Expected no errors reading body")
	}
	stringBody := string(body)
	value := strings.Contains(string(stringBody), `"value": 1`)
	unit := strings.Contains(string(stringBody), `"unit": "Binary"`)
	if value == false {
		t.Error("Expected value to be 1, but wasn't")
	}
	if unit == false {
		t.Error("Expected unit to be Binary, was something else")
	}

	// --- Bad test case: GET Wrong model ---
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Wrong model"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 500 {
		t.Errorf("Expected status code 500 w/ wrong model, was: %d", res.StatusCode)
	}

	// --- Bad test case: GET Error from getState() ---
	zResp.Body = errReader(0)
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", nil)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Smart plug"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 500 {
		t.Errorf("Expected status code 500 w/ error from getState(), was: %d", res.StatusCode)
	}

	// --- Good test case: PUT ---
	zResp.Body = io.NopCloser(strings.NewReader(zBeeResponse))
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	fakebody := `{"value": 0, "version": "SignalA_v1.0"}`
	sentBody := io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Smart plug"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 200 {
		t.Errorf("Expected status code 200, was: %d", res.StatusCode)
	}

	// --- Bad test case: PUT Wrong model ---
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	fakebody = `{"value": 0, "version": "SignalA_v1.0"}`
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Wrong model"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 500 {
		t.Errorf("Expected status code 500, was: %d", res.StatusCode)
	}

	// --- Bad test case: PUT Incorrectly formatted form ---
	zResp.Body = io.NopCloser(strings.NewReader(zBeeResponse))
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	fakebody = `{"value": a}`
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Smart plug"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 400 {
		t.Errorf("Expected status code to be 400, was %d", res.StatusCode)
	}

	// --- Bad test case: PUT breaking setState() ---
	newMockTransport(zResp, false, nil)
	w = httptest.NewRecorder()
	fakebody = `{"value": 3, "version": "SignalA_v1.0"}` // Value 3 not supported
	sentBody = io.NopCloser(strings.NewReader(fakebody))
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBeeHandler/SmartPlug1/state", sentBody)
	r.Header.Set("Content-Type", "application/json")
	ua.Model = "Smart plug"
	ua.state(w, r)
	res = w.Result()
	if res.StatusCode != 400 {
		t.Errorf("Expected status code to be 400, was %d", res.StatusCode)
	}
}
