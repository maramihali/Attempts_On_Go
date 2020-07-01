package requesttesting 

import (
	"log"
	"context"
	"net/http"
	"net"
	"testing"
)

type handlerRecorder struct {
	req *http.Request
	finishedReq chan bool
}


func (hr *handlerRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.req = r
	log.Print("Handler was called.")
	hr.finishedReq <- true
} 


func makeRequest(req string) *http.Request {
	e1, e2 := net.Pipe()
	listener := newMockListener(e2)
	done := make(chan bool)
	defer listener.Close()
	handler := &handlerRecorder{req: nil, finishedReq: done,}
	server := &http.Server{Handler: handler}
	go server.Serve(&listener)
	defer server.Shutdown(context.Background())
	e1.Write([]byte(req))
	<- handler.finishedReq

	return handler.req
}

func TestSetup(t *testing.T) {
	req := makeRequest("GET / HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" + "\r\n" )
		
	if req == nil {
		t.Error("Expected GET Request, received nil")
	}
	if req.Method != http.MethodGet {
		t.Errorf("Expected %s, received %s", http.MethodGet, req.Method)
	}
}

			
