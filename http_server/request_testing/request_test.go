package requesttesting

import (
	"context"
	"log"
	"net"
	"net/http"
	"testing"
)

type handlerRecorder struct {
	req         *http.Request
	finishedReq chan bool
	//TODO(mara): add server response
}


func (hr *handlerRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.req = r
	log.Print("Handler was called.")
	// log.Print(hr.req.URL.Query()["vegetable"])
	hr.finishedReq <- true
}

func makeRequest(req string) *http.Request {
	e1, e2 := net.Pipe()
	listener := newMockListener(e2)
	done := make(chan bool)
	defer listener.Close()
	handler := &handlerRecorder{req: nil, finishedReq: done}
	server := &http.Server{Handler: handler}
	go server.Serve(&listener)
	defer server.Shutdown(context.Background())
	e1.Write([]byte(req))
	<-handler.finishedReq
	return handler.req
}

func TestSetup(t *testing.T) {
	req := makeRequest("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")

	if req == nil {
		t.Error("Expected GET Request, received nil")
	}
	if req.Method != http.MethodGet {
		t.Errorf("Expected %s, received %s", http.MethodGet, req.Method)
	}
}

// Ensures Query() only returns a map of size one and verifies whether sending a
// request with two values for the same key eturns a []string of length 2
// containing the correct values 
func TestQMultipleValsSameKey(t *testing.T) {
	valueOne := "potatO"
	valueTwo := "Tomato"
	reqString := "GET /?vegetable=" + valueOne + "&vegetable=" + valueTwo + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
	log.Print(reqString)
	req := makeRequest(reqString)

	queryParams := req.URL.Query()
	if len(queryParams) != 1 {
		t.Errorf("Expected one query parameter to have been received, got %d", len(queryParams))
	}

	vegetableParamValues := queryParams["vegetable"]
	if len(vegetableParamValues) != 2 {
		t.Errorf("Expected 2 query parameter values got %d", len(vegetableParamValues))
	}
	if vegetableParamValues[0] != valueOne || vegetableParamValues[1] != valueTwo {
		t.Errorf("Expected "+valueOne+" and "+valueTwo+", got %s and %s", vegetableParamValues[0], vegetableParamValues[1])
	}

}

// Ensure different casing in  keys results in different query parameters
func TestQKeysDifferentCasing(t *testing.T) {
		req := makeRequest("GET /?vegetable=potato&Vegetable=tomato HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n")
		queryParams := req.URL.Query()
		if len(queryParams) != 2 {
			t.Errorf("Expected 2 query parameters, got %d", len(queryParams))
		}

	if len(queryParams["vegetable"]) != 1 || len(queryParams["Vegetable"]) != 1 {
		t.Errorf("Expected one value for query parameter vegetables and one for Vegetables, got %d and %d", len(queryParams["vegetable"]), len(queryParams["Vegetable"]))
	}
}

// Ensure keys and values that contain non-ASCII characters are parsed correctly
func TestQRomanianCharacters(t *testing.T) {
  value := "ăȚâȘî"
  reqString := "GET /?vegetable=" + value + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
  req := makeRequest(reqString)
 
  if   valueReceived := req.URL.Query()["vegetable"][0]; valueReceived != value {
	  t.Errorf("Expected %s as value but got %s", value, valueReceived)
  }

  key := "ăȚâȘî"
  reqString = "GET /?" + key + "=vegetable HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
  req = makeRequest(reqString)
  
  if  listLen := len(req.URL.Query()[key]); listLen != 1 {
	  t.Errorf("Expected 1 value for key %s but got %d", key, listLen)
  }


}


// Check whether ParseQuery() gives us all the request
