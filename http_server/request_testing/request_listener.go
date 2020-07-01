package requesttesting

import (
	"net"
	"errors"
	"log"
)

type mockListener struct {
	connChannel chan net.Conn
	serverEndpoint net.Conn
}

// For each request, we create an HTTP listener that passes the request provided
// as a function parameter to the server.
func newMockListener(endpoint  net.Conn) mockListener {

	channel := make(chan net.Conn, 1)
	channel <- endpoint
	listener := mockListener{
		connChannel: channel,
		serverEndpoint: endpoint,
	}
	
	return listener
}

// Passes an endpoint to the server. Weirdly enough, when I make one request,
// Accept() and hence Close() are being called twice, but the handler only once.
func (l *mockListener) Accept() (net.Conn, error) {
	log.Print("Accept() called")
	ch, ok := <- l.connChannel
	if !ok {
		return nil, errors.New("EOF")
	}
	return ch, nil
}

func (l *mockListener) Close() error {
	log.Print("Close() called")
	// My intuition here would have been to close both the endpoint and the
	// channel. However, if I do that, I trigger a panic of "close of closed
	// channel" when I attempt to close the channel. If I only close the // //
	// channel, the program never terminates (I
	// assume the server continues listening)
	l.serverEndpoint.Close()
	// close(l.connChannel)
	return nil
}

func (l *mockListener) Addr() net.Addr {
	return l.serverEndpoint.LocalAddr()
}

