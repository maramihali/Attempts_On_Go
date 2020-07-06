package requesttesting

import (
	"log"
	"strconv"
	"testing"
)

func TestSimpleFormParameters(t *testing.T) {
	reqBody := "Veg=Potato&Fruit=Apple"
	reqBodyLen := len(reqBody)
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length:" + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n"
	req, _ := makeRequest(postReq)
	log.Print(req.Header)
	req.ParseForm()
	if req.Form["Veg"][0] != "Potato" && req.Form["Fruit"][0] != "Apple" {
		t.Errorf("Correct form parameters: expected Potato and Apple but got %s and %s", req.Form["Veg"][0], req.Form["Fruit"][0])
	}
	
}

// Test whether passing a POST request with body and without Content-Length
// yields a 400 Bad Request
func TestFormParametersMissingContentLength(t *testing.T) {
	reqBody := "veggie=potato"
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + reqBody + "\r\n" + "\r\n"
	_, resp := makeRequest(postReq)
	if !isBadRequestResponse(resp) {
		t.Errorf("Server response: expected it to be 400, got %s", string(resp))
	}
}

// Tests behaviour when multiple Content-Length are passed. If they are of equal
// length, the request will succeed, otherwise it fails.
func TestFormParametersDuplicateContentLength(t *testing.T) {
	reqBody := "veggie=potato"
	postReq1 := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=utf-8\r\n" + "Content-Length: 13\r\n" + "Content-Length: 13\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n"
	req, _ := makeRequest(postReq1)
	log.Print(req)
	if contentLen := req.Header["Content-Length"][0]; contentLen != "13" {
		t.Errorf("Expected Content-Length to be 13 but got %s", contentLen)
	}

	postReq2 := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length: 13\r\n" + "Content-Length: 12\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n"
	req, resp := makeRequest(postReq2)
	if !isBadRequestResponse(resp) {
		t.Errorf("Expected 400 but got %s", resp)
	}

}

// Test whether form parameters with Content-Type:
// application/x-www-form-urlencoded that break percent-encoding will be
// ignored
func TestFormParametersBreakUrlEncoding(t *testing.T) {

	reqBody := "veggie=%sc"
	reqBodyLen := len(reqBody)
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "Content-Type: application/x-www-form-urlencoded; charset=ASCII\r\n" + "Content-Length:" + strconv.Itoa(reqBodyLen) + "\r\n" + "\r\n" + reqBody + "\r\n" + "\r\n"
	req, _ := makeRequest(postReq)
	if len(req.Form["veggie"]) != 0 {
		t.Error("Expected form value to break percent encoding.")
	}
	
}

// It might be the case that charset should not be a thing

// Test the correctness of a simple multipart/form
func TestBasicMultipartForm(t *testing.T) {
	postReq := "POST / HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + 
				"Content-Type: multipart/form-data; boundary=\"123\"\r\n" + "Content-Length: 66\r\n"+
				"\r\n" + "--123\r\n" + "Content-Disposition: form-data; name=\"foo\"\r\n"  + "\r\n" + "bar\r\n" + "--123--\r\n"
	log.Print(postReq)
	req, resp := makeRequest(postReq)
	log.Print(string(resp))
	req.ParseMultipartForm(0)
	log.Print("Form value in test:"+ req.FormValue("foo"))
}





// "boundary delimiter constructed with CRLF, --, boundary value"
//  "see what happens if you dont pass boundary type"
//  "boundary appeared in encapsulated part as well"
//  "check Content-Disposition is enforced"
// multipart/form-data boundary wierdness
// Content-Disposition wierdness

// NON-ASCII FIELD NAMES
