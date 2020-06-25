package main

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func TestWelcome(t *testing.T) {

	req := httptest.NewRequest("GET", "localhost:8080/welcome?name=mara", nil)
	w := httptest.NewRecorder()
	welcomeHandler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))
}
