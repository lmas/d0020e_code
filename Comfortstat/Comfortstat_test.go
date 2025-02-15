package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpSetSEKPrice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	//Good case test: GET
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/SEKPrice", nil)
	goodCode := 200
	ua.httpSetSEKPrice(w, r)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	resp := w.Result()
	if resp.StatusCode != goodCode {
		t.Errorf("expected good status code: %v, got %v", goodCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 1.5`)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
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
	// Bad test case: default part of code
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8670/Comfortstat/Set%20Values/SEKPrice", nil)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	ua.httpSetSEKPrice(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetMinTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	//Godd test case: PUT
	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 20, "unit": "Celsius", "version": "SignalA_v1.0"}`))         // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MinTemperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                           // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetMinTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))                // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MinTemperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                          // basic setup to prevent the request to be rejected.
	ua.httpSetMinTemp(w, r)
	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/MinTemperature", nil)
	goodStatusCode = 200
	ua.httpSetMinTemp(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 20`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
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
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/MinTemperature", nil)
	ua.httpSetMinTemp(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetMaxTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 25, "unit": "Celsius", "version": "SignalA_v1.0"}`))         // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MaxTemperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                           // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetMaxTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))                // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MaxTemperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                          // basic setup to prevent the request to be rejected.
	ua.httpSetMaxTemp(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/MaxTemperature", nil)
	goodStatusCode = 200
	ua.httpSetMaxTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 25`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
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
	r = httptest.NewRequest("666", "localhost:8670/Comfortstat/Set%20Values/MaxTemperature", nil)

	ua.httpSetMaxTemp(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetMinPrice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 1, "unit": "SEK", "version": "SignalA_v1.0"}`)) // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/MinPrice", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                              // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetMinPrice(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "SEK", "version": "SignalA_v1.0"}`))              // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MinPrice", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                    // basic setup to prevent the request to be rejected.
	ua.httpSetMinPrice(w, r)
	// save the rsponse
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/MinPrice", nil)
	goodStatusCode = 200
	ua.httpSetMinPrice(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 1`)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
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
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/MinPrice", nil)
	ua.httpSetMinPrice(w, r)
	//save the response
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetMaxPrice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 2, "unit": "SEK", "version": "SignalA_v1.0"}`))        // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MaxPrice", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                     // basic setup to prevent the request to be rejected.
	goodStatusCode := 200
	ua.httpSetMaxPrice(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "SEK", "version": "SignalA_v1.0"}`))              // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/MaxPrice", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                    // basic setup to prevent the request to be rejected.
	ua.httpSetMaxPrice(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/MaxPrice", nil)
	goodStatusCode = 200
	ua.httpSetMaxPrice(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 2`)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
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
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/MaxPrice", nil)

	ua.httpSetMaxPrice(w, r)
	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetDesiredTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 0, "unit": "Celsius", "version": "SignalA_v1.0"}`))       // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/DesiredTemp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                        // basic setup to prevent the request to be rejected.
	goodStatusCode := 200

	ua.httpSetDesiredTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))             // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/DesiredTemp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                       // basic setup to prevent the request to be rejected.

	ua.httpSetDesiredTemp(w, r)
	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/DesiredTemp", nil)
	goodStatusCode = 200
	ua.httpSetDesiredTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 0`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
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
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/DesiredTemp", nil)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	ua.httpSetDesiredTemp(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetUserTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 0, "unit": "Celsius", "version": "SignalA_v1.0"}`))    // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/userTemp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                     // basic setup to prevent the request to be rejected.
	goodStatusCode := 200

	ua.httpSetUserTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))          // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/userTemp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                    // basic setup to prevent the request to be rejected.

	ua.httpSetUserTemp(w, r)
	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/userTemp", nil)
	goodStatusCode = 200
	ua.httpSetUserTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 0`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
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
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/userTemp", nil)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	ua.httpSetUserTemp(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}

func TestHttpSetRegion(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 1, "version": "SignalA_v1.0"}`))                     // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/Region", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                   // basic setup to prevent the request to be rejected.
	goodStatusCode := 200

	ua.httpSetRegion(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "version": "SignalA_v1.0"}`))                           // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/Region", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                  // basic setup to prevent the request to be rejected.

	ua.httpSetRegion(w, r)
	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == goodStatusCode {
		t.Errorf("expected bad status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	//Good test case: GET
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/Region", nil)
	goodStatusCode = 200
	ua.httpSetRegion(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	if resp.StatusCode != goodStatusCode {
		t.Errorf("expected good status code: %v, got %v", goodStatusCode, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// this is a simple check if the JSON response contains the specific value/unit/version
	value := strings.Contains(string(body), `"value": 1`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if value != true {
		t.Errorf("expected the statement to be true!")
	}
	if version != true {
		t.Errorf("expected the version statement to be true!")
	}
	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/Region", nil)
	// calls the method and extracts the response and save is in resp for the upcoming tests
	ua.httpSetRegion(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)
	}
}
