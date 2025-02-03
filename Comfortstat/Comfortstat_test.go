package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_set_SEKprice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	//Good case test: GET
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/SEK_price", nil)
	good_code := 200

	ua.set_SEKprice(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 1.5`)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_code {
		t.Errorf("expected good status code: %v, got %v", good_code, resp.StatusCode)
	}

	if value != true {
		t.Errorf("expected the statment to be true!")

	}
	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("expected the version statment to be true!")
	}
	// Bad test case: default part of code
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8670/Comfortstat/Set%20Values/SEK_price", nil)

	ua.set_SEKprice(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}

}

func Test_set_minTemp(t *testing.T) {

	ua := initTemplate().(*UnitAsset)

	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 20, "unit": "Celsius", "version": "SignalA_v1.0"}`))          // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/min_temperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                            // basic setup to prevent the request to be rejected.
	good_statuscode := 200

	ua.set_minTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))                 // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/min_temperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                           // basic setup to prevent the request to be rejected.

	ua.set_minTemp(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == good_statuscode {
		t.Errorf("expected bad status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//Good test case: GET

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/min_temperature", nil)
	good_statuscode = 200
	ua.set_minTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 20`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	if value != true {
		t.Errorf("expected the statment to be true!")

	}

	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}

	if version != true {
		t.Errorf("expected the version statment to be true!")
	}

	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/min_temperature", nil)

	ua.set_minTemp(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)

	}
}

func Test_set_maxTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 25, "unit": "Celsius", "version": "SignalA_v1.0"}`))          // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/max_temperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                            // basic setup to prevent the request to be rejected.
	good_statuscode := 200

	ua.set_maxTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))                 // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "http://localhost:8670/Comfortstat/Set%20Values/max_temperature", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                           // basic setup to prevent the request to be rejected.

	ua.set_maxTemp(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == good_statuscode {
		t.Errorf("expected bad status code: %v, got %v", good_statuscode, resp.StatusCode)
	}
	//Good test case: GET

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/max_temperature", nil)
	good_statuscode = 200
	ua.set_maxTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 25`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	if value != true {
		t.Errorf("expected the statment to be true!")

	}

	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}

	if version != true {
		t.Errorf("expected the version statment to be true!")
	}

	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "localhost:8670/Comfortstat/Set%20Values/max_temperature", nil)

	ua.set_maxTemp(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)

	}
}

func Test_set_minPrice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 1, "unit": "SEK", "version": "SignalA_v1.0"}`))  // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/min_price", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                               // basic setup to prevent the request to be rejected.
	good_statuscode := 200

	ua.set_minPrice(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "SEK", "version": "SignalA_v1.0"}`))        // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/min_price", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                              // basic setup to prevent the request to be rejected.

	ua.set_minPrice(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == good_statuscode {
		t.Errorf("expected bad status code: %v, got %v", good_statuscode, resp.StatusCode)
	}
	//Good test case: GET

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "localhost:8670/Comfortstat/Set%20Values/min_price", nil)
	good_statuscode = 200
	ua.set_minPrice(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 1`) //EVENTUELL BUGG, enligt webb-app minPrice = 0 ( kanske dock är för att jag inte startat sregistrar och orchastrator)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	if value != true {
		t.Errorf("expected the statment to be true!")

	}

	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}

	if version != true {
		t.Errorf("expected the version statment to be true!")
	}

	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "localhost:8670/Comfortstat/Set%20Values/min_price", nil)

	ua.set_minPrice(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)

	}
}

func Test_set_maxPrice(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 2, "unit": "SEK", "version": "SignalA_v1.0"}`))  // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/max_price", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                               // basic setup to prevent the request to be rejected.
	good_statuscode := 200

	ua.set_maxPrice(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "SEK", "version": "SignalA_v1.0"}`))        // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/max_price", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                              // basic setup to prevent the request to be rejected.

	ua.set_maxPrice(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == good_statuscode {
		t.Errorf("expected bad status code: %v, got %v", good_statuscode, resp.StatusCode)
	}
	//Good test case: GET

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/max_price", nil)
	good_statuscode = 200
	ua.set_maxPrice(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 2`)
	unit := strings.Contains(string(body), `"unit": "SEK"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	if value != true {
		t.Errorf("expected the statment to be true!")

	}

	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}

	if version != true {
		t.Errorf("expected the version statment to be true!")
	}

	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/max_price", nil)

	ua.set_maxPrice(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)

	}
}

func Test_set_desiredTemp(t *testing.T) {
	ua := initTemplate().(*UnitAsset)
	//Godd test case: PUT

	// creates a fake request body with JSON data
	w := httptest.NewRecorder()
	fakebody := bytes.NewReader([]byte(`{"value": 0, "unit": "Celsius", "version": "SignalA_v1.0"}`)) // converts the Jason data so it can be read
	r := httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/desired_temp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                  // basic setup to prevent the request to be rejected.
	good_statuscode := 200

	ua.set_desiredTemp(w, r)

	// save the rsponse and read the body
	resp := w.Result()
	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}

	//BAD case: PUT, if the fake body is formatted incorrectly

	// creates a fake request body with JSON data
	w = httptest.NewRecorder()
	fakebody = bytes.NewReader([]byte(`{"123, "unit": "Celsius", "version": "SignalA_v1.0"}`))       // converts the Jason data so it can be read
	r = httptest.NewRequest("PUT", "localhost:8670/Comfortstat/Set%20Values/desired_temp", fakebody) // simulating a put request from a user to update the min temp
	r.Header.Set("Content-Type", "application/json")                                                 // basic setup to prevent the request to be rejected.

	ua.set_desiredTemp(w, r)

	// save the rsponse and read the body
	resp = w.Result()
	if resp.StatusCode == good_statuscode {
		t.Errorf("expected bad status code: %v, got %v", good_statuscode, resp.StatusCode)
	}
	//Good test case: GET

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://localhost:8670/Comfortstat/Set%20Values/desired_temp", nil)
	good_statuscode = 200
	ua.set_desiredTemp(w, r)

	// save the rsponse and read the body

	resp = w.Result()
	body, _ := io.ReadAll(resp.Body)

	value := strings.Contains(string(body), `"value": 0`)
	unit := strings.Contains(string(body), `"unit": "Celsius"`)
	version := strings.Contains(string(body), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_statuscode {
		t.Errorf("expected good status code: %v, got %v", good_statuscode, resp.StatusCode)
	}
	if value != true {
		t.Errorf("expected the statment to be true!")

	}

	if unit != true {
		t.Errorf("expected the unit statement to be true!")
	}

	if version != true {
		t.Errorf("expected the version statment to be true!")
	}

	// bad test case: default part of code

	// force the case to hit default statement but alter the method
	w = httptest.NewRecorder()
	r = httptest.NewRequest("666", "http://localhost:8670/Comfortstat/Set%20Values/desired_temp", nil)

	ua.set_desiredTemp(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected the status to be bad but got: %v", resp.StatusCode)

	}
}
