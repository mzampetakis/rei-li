package main

import (
	"fmt"
	"github.com/mzampetakis/reili"
	"net"
	"net/http"
)

func main() {
	var visitorIdentifier reili.VisitorIdentifier
	visitorIdentifier = &VisitorIdentifierByIP{}
	reqPerSec := 1.0
	burstReq := 3
	limiter := reili.NewRateLimiter(reqPerSec, burstReq, visitorIdentifier)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	// Wrap the servemux with the limit middleware.
	port := "4000"
	fmt.Println("Listening on :" + port)
	http.ListenAndServe(":"+port, limiter.Limit(mux))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hi from index!"))
}

// VisitorIdentifierByIP used to implement reili.VisitorIdentifier Interface
type VisitorIdentifierByIP struct {
}

// Returns the request's IP or error
func (vi *VisitorIdentifierByIP) IdentifyVisitor(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}
