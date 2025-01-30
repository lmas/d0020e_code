package main

import (
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
