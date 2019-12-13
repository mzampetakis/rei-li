// Package reili provides a simple but general purpose
// http rate limiter http middleware.
// It uses the rate package and support limit and burst at second window.
// Visitor identifier is defined by the end user.
package reili

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is the basic struct of reili that NewRateLimiter returns.
// All field are not exported so it should be used only for variable initialization.
type RateLimiter struct {
	requestsPerSecond float64
	burstRequests     int
	visitors          map[interface{}]*visitor
	mx                sync.RWMutex
	visitorIdentifier VisitorIdentifier
}

type visitor struct {
	id       interface{}
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter instantiates a new RateLimiter object.
// rps: requests per second
// burstReq: burst requests
// identifier: a Visitor Identifier interface
func NewRateLimiter(rps float64, burstReq int, identifier VisitorIdentifier) *RateLimiter {
	newRateLimiter := RateLimiter{
		requestsPerSecond: rps,
		burstRequests:     burstReq,
		visitors:          make(map[interface{}]*visitor),
		mx:                sync.RWMutex{},
		visitorIdentifier: identifier,
	}
	go newRateLimiter.cleanupVisitors()

	return &newRateLimiter
}

// VisitorIdentifier is an interface with a single function requirement.
// The function that requires is a IdentifyVisitor(r *http.Request) (string, error)
// which is used to identify a user by a string using the *http.Request
// It is used in every Limit call to identify the user.
type VisitorIdentifier interface {
	IdentifyVisitor(r *http.Request) (string, error)
}

// Limit is an http limiter middleware that uses a specific RateLimiter
// to limit or allow http requests.
func (reili *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := reili.visitorIdentifier.IdentifyVisitor(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		reqVisitor := reili.getVisitor(id)
		if reqVisitor.limiter.Allow() == false {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (reili *RateLimiter) getVisitor(identifier interface{}) *visitor {
	reili.mx.RLock()
	defer reili.mx.RUnlock()

	_, exists := reili.visitors[identifier]
	if !exists {
		return reili.addVisitor(identifier)
	}
	reili.visitors[identifier].lastSeen = time.Now()
	return reili.visitors[identifier]
}

func (reili *RateLimiter) addVisitor(identifier interface{}) *visitor {
	limiter := rate.NewLimiter(rate.Limit(reili.requestsPerSecond), reili.burstRequests)
	reili.visitors[identifier] = &visitor{
		id:       identifier,
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	return reili.visitors[identifier]
}

// Every minute check the map for visitors that haven't been seen for
// more than 5 minutes and delete the entries.
func (reili *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		reili.mx.Lock()
		defer reili.mx.Unlock()
		for id, v := range reili.visitors {
			if time.Now().Sub(v.lastSeen) > 5*time.Minute {
				delete(reili.visitors, id)
			}
		}
	}
}
