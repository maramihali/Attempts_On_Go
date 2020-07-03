package requesttesting

import (
	"io"
	"log"
	"net"
	"sync"
	"net/http"
	"regexp"
	"context"
)

type mockListener struct {
	closeOnce      sync.Once
	connChannel    chan net.Conn
	serverEndpoint net.Conn
}


type recorderHandler struct {
	req     *http.Request
	doneReq chan bool
}

// Creates a mock listener that passes requests to the HTTP server as part of
// the test harness
func newMockListener(endpoint net.Conn) *mockListener {

	c := make(chan net.Conn, 1)
	c <- endpoint
	listener := &mockListener{
		connChannel:    c,
		serverEndpoint: endpoint,
	}
	return listener
}

// Passes an endpoint to the server to enable communication to client
func (l *mockListener) Accept() (net.Conn, error) {
	log.Print("Accept() called")
	ch, ok := <-l.connChannel
	if !ok {
		return nil, io.EOF
	}
	return ch, nil
}

func (l *mockListener) Close() (err error) {
	log.Print("Close() called")
	o := &l.closeOnce
	o.Do(func() {
		err = l.serverEndpoint.Close()
		close(l.connChannel)
	})
	return err 
}

func (l *mockListener) Addr() net.Addr {
	return l.serverEndpoint.LocalAddr()
}

func (hr *recorderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hr.req = r
	log.Print("Handler was called.")
	r.ParseForm()
	log.Print(r.FormValue("veggie"))
	hr.doneReq <- true
} 

func makeRequest(req string) (reqReceived *http.Request, resp []byte) {

	clientEndpoint, serverEndpoint := net.Pipe()
	listener := newMockListener(serverEndpoint)
	// defer listener.Close()
	handler := &recorderHandler{req: nil, doneReq: make(chan bool)}
	server := &http.Server{Handler: handler}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())
	clientEndpoint.Write([]byte(req))
	resp = make([]byte, 4096)
	go func() {
		clientEndpoint.Read(resp)
		// TODO(MARA) replace with something bettere
		r, _ := regexp.Compile("HTTP/[0-9].[0-9] [4-5][0-9][0-9]")
		if r.MatchString(string(resp)) {
			handler.doneReq <- false
		}
	}()
	<-handler.doneReq
	reqReceived = handler.req
	return
}
