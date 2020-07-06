package requesttesting

import (
	"log"
	"testing"
)


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

