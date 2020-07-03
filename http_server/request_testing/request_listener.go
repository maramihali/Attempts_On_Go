package requesttesting

import (
	"io"
	"log"
	"net"
	"sync"
)

type mockListener struct {
	closeOnce      sync.Once
	connChannel    chan net.Conn
	serverEndpoint net.Conn
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

func (l *mockListener) Close() error {
	log.Print("Close() called")
	o := &l.closeOnce
	o.Do(func() {
		l.serverEndpoint.Close()
		close(l.connChannel)
	})
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return l.serverEndpoint.LocalAddr()
}
