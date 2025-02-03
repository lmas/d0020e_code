package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sdoque/mbaigo/forms"
	"github.com/sdoque/mbaigo/usecases"
)

func TestSetpt(t *testing.T) {
	ua := initTemplate().(*UnitAsset)

	// Good case test: GET
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost:8670/ZigBee/Template/setpoint", nil)
	good_code := 200
	ua.setpt(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	stringBody := string(body)
	// fmt.Println(stringBody)

	value := strings.Contains(string(stringBody), `"value": 20`)
	unit := strings.Contains(string(stringBody), `"unit": "Celcius"`)
	version := strings.Contains(string(stringBody), `"version": "SignalA_v1.0"`)

	if resp.StatusCode != good_code {
		t.Errorf("Good GET: Expected good status code: %v, got %v", good_code, resp.StatusCode)
	}
	if value != true {
		t.Errorf("Good GET: The statment to be true!")
	}
	if unit != true {
		t.Errorf("Good GET: Expected the unit statement to be true!")
	}
	if version != true {
		t.Errorf("Good GET: Expected the version statment to be true!")
	}
	// Bad test case: default part of code
	w = httptest.NewRecorder()
	r = httptest.NewRequest("123", "http://localhost:8670/ZigBee/Template/setpoint", nil)

	ua.setpt(w, r)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected the status to be bad but got: %v", resp.StatusCode)
	}
	// ALL THE ABOVE PASSES TESTS

	// Good test case: PUT
	// Send PUT request to change
	w = httptest.NewRecorder()
	// Make the body
	var of forms.SignalA_v1a
	of.NewForm()
	of.Value = 25.0
	of.Unit = "Celcius"
	of.Timestamp = time.Now()
	op, _ := usecases.Pack(&of, "application/json")
	log.Println(string(op))
	sentBody := io.NopCloser(strings.NewReader(string(op)))
	// Send the request
	r = httptest.NewRequest("PUT", "http://localhost:8870/ZigBee/Template/setpoint", sentBody)
	ua.setpt(w, r)
	resp = w.Result()
	good_code = 200
	// Check for errors
	if resp.StatusCode != good_code {
		t.Errorf("Good PUT: Expected good status code: %v, got %v", good_code, resp.StatusCode)
	}

}
