package requesttesting
 
import (
   "log"
   "net/http"
   "testing"
)
 
// Tests the setup created to make requests and receive responses
func TestMakeRequest(t *testing.T) {
 
   req, _ := makeRequest("GET / HTTP/1.1\r\n" +
       "Host: localhost:8080\r\n" + "\r\n")
 
   if req.Method != http.MethodGet {
       t.Errorf("Expected %s, received %s", http.MethodGet, req.Method)
   }
}
 
// Ensures Query() only returns a map of size one and verifies whether sending a
// request with two values for the same key eturns a []string of length 2
// containing the correct values
func TestMultipleQueryParametersSameKey(t *testing.T) {
   var (
       valueOne  = "potatO"
       valueTwo  = "Tomato"
       reqString = "GET /?vegetable=" + valueOne + "&vegetable=" + valueTwo + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   )
   req, _ := makeRequest(reqString)
   queryParams := req.URL.Query()
   if len(queryParams) != 1 {
       t.Errorf("len(queryParams): got %d, want %d", len(queryParams), 1)
   }
 
   vegetableParamValues := queryParams["vegetable"]
   if len(vegetableParamValues) != 2 {
       t.Errorf("len(vegetableQueryParams): got %d, want %d", len(vegetableParamValues), 2)
   }
   if vegetableParamValues[0] != valueOne || vegetableParamValues[1] != valueTwo {
       t.Errorf("queryParams values: expected "+valueOne+" and "+valueTwo+", got %s and %s", vegetableParamValues[0], vegetableParamValues[1])
   }
 
}
 
// Ensure different casing in  keys results in different query parameters
func TestQueryParametersSameKeyDifferentCasing(t *testing.T) {
   req, _ := makeRequest("GET /?vegetable=potato&Vegetable=tomato HTTP/1.1\r\n" +
       "Host: localhost:8080\r\n" + "\r\n")
   queryParams := req.URL.Query()
   if len(queryParams) != 2 {
       t.Errorf("len(queryParams): got %d, want %d", len(queryParams), 2)
   }
 
   if len(queryParams["vegetable"]) != 1 || len(queryParams["Vegetable"]) != 1 {
       t.Errorf("Expected one value for query parameter vegetables and one for Vegetables, got %d and %d", len(queryParams["vegetable"]), len(queryParams["Vegetable"]))
   }
   log.Print(queryParams["vegetable"][0])
}
 
// Ensure keys and values that contain non-ASCII characters are parsed correctly
func TestQueryParametersValidUnicode(t *testing.T) {
   value := "ăȚâȘî"
   reqString := "GET /?vegetable=" + value + " HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   req, _ := makeRequest(reqString)
 
   if valueReceived := req.URL.Query()["vegetable"][0]; valueReceived != value {
       t.Errorf("queryParams values: got %s, want %s", valueReceived, value)
   }
 
   key := "ăȚâȘî"
   reqString = "GET /?" + key + "=vegetable HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   req, _ = makeRequest(reqString)
 
   if listLen := len(req.URL.Query()[key]); listLen != 1 {
       t.Errorf("len(queryParamsKey): got %d, want 1 value for %s", listLen, key)
   }
}
 
// Tests whether passing invalid Unicode will result  in a 400 Bad Request error
func TestQueryParametersInvalidUnicodes(t *testing.T) {
   keyOne := "\x0F"
   reqString := "GET /?" + keyOne + "=tomato&Vegetable=potato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   req, resp := makeRequest(reqString)
   if req != nil {
       t.Error("Expected the server not to receive a request containing invalid Unicode.")
   }
 
   if !isBadRequestResponse(resp) {
       t.Errorf("Server response: expected it to be 400, got %s", string(resp))
   }
 
}
 
// Test behaviour of query parameter parser when passing malformed key or
// values (by breaking URL encoding). If using Query(), it is supposed to
// silently discard malformed values. In URLs, percent are used used for
// percent-encoding special characters
func TestQueryParametersBreakUrlEncoding(t *testing.T) {
   brokenKeyReq := "GET /?vegetable%=tomato HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   req, _ := makeRequest(brokenKeyReq)
   if lenQueryParams := len(req.URL.Query()); lenQueryParams != 0 {
       t.Errorf("len(queryParams): got %d, want %d", lenQueryParams, 0)
   }
 
   brokenValueReq := "GET /?vegetable=tomato% HTTP/1.1\r\n" + "Host: localhost:8080\r\n" + "\r\n"
   req, _ = makeRequest(brokenValueReq)
   if lenVeg := len(req.URL.Query()["vegetable"]); lenVeg != 0 {
       t.Errorf("len(queryParams): got %d, want %d", lenVeg, 0)
   }
 
}

// func TestQueryParametersSpecial(t *testing.T) {

// 	 req, resp := makeRequest("GET /?vegetable=potato HTTP/1.1\r\n" +
//        "Host: localhost:8080\r\n" + "\r\n")
// 	log.Print(string(resp))
// 	log.Print(req.URL.Query()["vegetable"])
//       if isBadRequestResponse(resp) {
//        t.Errorf("Server response: expected it to be 400, got %s", string(resp))
//    }
   
   
// }

 
 
// Test whether both + and %20 are interpreted as space. Having a space in the
// actual request will not be escaped and will result in a 400
func TestQueryParametersSpaceBehaviour(t *testing.T) {
       _, resp := makeRequest("GET /?vegetable=potato pear&Vegetable=tomato HTTP/1.1\r\n" +
       "Host: localhost:8080\r\n" + "\r\n")
       if !isBadRequestResponse(resp) {
         t.Errorf("Server response: expected it to be 400, got %s", string(resp))
        }
 
       req1, _ := makeRequest("GET /?vegetable=potato+pear&Vegetable=tomato HTTP/1.1\r\n" +
       "Host: localhost:8080\r\n" + "\r\n")
       req2, _ := makeRequest("GET /?vegetable=potato%20pear&Vegetable=tomato HTTP/1.1\r\n" +
       "Host: localhost:8080\r\n" + "\r\n")
       s1 := req1.URL.Query()["vegetable"][0]
       s2 := req2.URL.Query()["vegetable"][0]
       if s1 != s2 {
           t.Errorf("Expected identical parsing but got %s and %s", s1, s2)
       }
}
 
