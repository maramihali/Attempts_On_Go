package requesttesting
 
import (
   "io"
   "io/ioutil"
   "log"
   "net"
   "sync"
   "net/http"
   "regexp"
   "context"
   "strings"
)
 
type mockListener struct {
   closeOnce      sync.Once
   connChannel    chan net.Conn
   serverEndpoint net.Conn
}
 
 
type recorderHandler struct {
   req     *http.Request
   // reqBody []byte
   doneReq chan bool
}


func isBadRequestResponse(resp []byte) bool {
	
   badReq := "400 Bad Request"
   if respString := string(resp); !strings.Contains(respString, badReq) {
		return false
   }

   return true
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
   b, err := ioutil.ReadAll(r.Body)
   log.Printf("body: %s", b)
   if err != nil {
      log.Printf("%s",err.Error())
   }

   err = r.ParseMultipartForm(1024)
   if err != nil {
      log.Printf("%v", err.Error())
   }
   log.Print("Form value in handler:" + r.FormValue("foo"))
   hr.doneReq <- true
}
 
func makeRequest(req string) (reqReceived *http.Request, resp []byte) {
 
   clientEndpoint, serverEndpoint := net.Pipe()
   listener := newMockListener(serverEndpoint)
   // defer listener.Close()
   handler := &recorderHandler{req: nil,  doneReq: make(chan bool)}
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