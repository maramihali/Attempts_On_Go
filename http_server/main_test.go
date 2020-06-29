package main

import (
	"io/ioutil"
	"net/http/httptest"
	"net/http"
	"testing"
	"strings"

)

func TestWelcome(t *testing.T) {
	
	data := &serviceData{}
	req := httptest.NewRequest("GET", "/welcome/?name=mara", nil)

	w := httptest.NewRecorder()
	mux := data.muxSetup()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if !strings.Contains(string(body), "mara") {
		t.Error("Query parameters have not been parsed properly")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200 but got %d", resp.StatusCode)
	}
}
