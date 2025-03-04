package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpSetButton(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	// Good case test: GET
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://172.30.106.39:8670/SunButton/Button/ButtonStatus", nil)
	goodStatusCode := 200
	ua.httpSetButton(w, r)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 0.5`)
	unit := strings.Contains(string(body), `"unit": "bool"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)
	// check results from above
	if value != true {
		t.Errorf("expected the statement to be true!")
	}
	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("expected the version statement to be true!")
	}

	//Godd test case: PUT
	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 0, "unit": "bool", "version": "SignalA_v1.0"}`))      // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/ButtonStatus", fakebody) // simulating a put request from a user to update the button status
	r.Header.Set("Content-Type", "application/json")                                                    // basic setup to prevent the request to be rejected.
	ua.httpSetButton(w, r)

	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "bool", "version": "SignalA_v1.0"}`))             // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/ButtonStatus", fakebody) // simulating a put request from a user to update the button status
	r.Header.Set("Content-Type", "application/json")                                                    // basic setup to prevent the request to be rejected.
	ua.httpSetButton(w, r)
	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	// Bad test case: default part of code
	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://172.30.106.39:8670/SunButton/Button/ButtonStatus", nil)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	ua.httpSetButton(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetLatitude(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	//Godd test case: PUT
	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 65.584816, "unit": "Degrees", "version": "SignalA_v1.0"}`)) // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/Latitude", fakebody)          // simulating a put request from a user to update the latitude
	r.Header.Set("Content-Type", "application/json")                                                          // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetLatitude(w, r)

	// save the response and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Degrees", "version": "SignalA_v1.0"}`))      // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/Latitude", fakebody) // simulating a put request from a user to update the latitude
	r.Header.Set("Content-Type", "application/json")                                                // basic setup to prevent the request to be rejected.
	ua.httpSetLatitude(w, r)
	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://172.30.106.39:8670/SunButton/Button/Latitude", nil)
	goodStatusCode = 200
	ua.httpSetLatitude(w, r)

	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 65.584816`)
	unit := strings.Contains(string(body), `"unit": "Degrees"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)
	// check the result from above
	if value != true {
		t.Errorf("expected the statement to be true!")
	}
	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("expected the version statement to be true!")
	}
	// bad test case: default part of code
	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://172.30.106.39:8670/SunButton/Button/Latitude", nil)
	ua.httpSetLatitude(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetLongitude(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 22.156704, "unit": "Degrees", "version": "SignalA_v1.0"}`)) // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/Longitude", fakebody)         // simulating a put request from a user to update the longitude
	r.Header.Set("Content-Type", "application/json")                                                          // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetLongitude(w, r)

	// save the response and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Degrees", "version": "SignalA_v1.0"}`))       // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://172.30.106.39:8670/SunButton/Button/Longitude", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                 // basic setup to prevent the request to be rejected.
	ua.httpSetLongitude(w, r)

	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://172.30.106.39:8670/SunButton/Button/Longitude", nil)
	goodStatusCode = 200
	ua.httpSetLongitude(w, r)

	// save the response and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 22.156704`)
	unit := strings.Contains(string(body), `"unit": "Degrees"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)
	if value != true {
		t.Errorf("expected the statement to be true!")
	}
	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("expected the version statement to be true!")
	}
	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://172.30.106.39:8670/SunButton/Button/Longitude", nil)
	ua.httpSetLongitude(w, r)
	//save the response
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}
