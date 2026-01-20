package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"golang.org/x/time/rate"
)

// client struct which represents a single instance communicating with the system,
type client struct {
	// defines the rate limiter structure, uses gos x/time/rate library which uses
	// the token bucket algorithm
	limiter *rate.Limiter

	// keeps track of the last time the client communicated with the server so we
	// we know which clients to invalidate after they haven't been used for a while
	lastSeen time.Time
}

// RateLimiter struct manages clients which recently communicated with the service,
type RateLimiter struct {
	// clients are stored in a map K: ip_address, V: client instance
	clients map[string]*client

	// since go http servers handle every request in a separate go routine if multiple
	// requests are being handled at once, they will all try to write/ read from the
	// clients map and that's why we use the Mutex mechanism
	mu *sync.RWMutex

	// defines how much tokens are added to the token bucket
	rate rate.Limit

	// defines the size of the bucker, or to be precise how much tokens can be stored
	// in the bucket at once
	burst int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	limiter := &RateLimiter{
		clients: make(map[string]*client),
		mu:      &sync.RWMutex{},
		rate:    r,
		burst:   b,
	}

	// run the cleanupClients method in a separate go routine to make sure that unused
	// connections don't use up space
	go limiter.cleanupClients()

	return limiter
}

// getLimiter represents a method which is used to get the limiter for a specific client
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// if a client doesn't exist for the given ip address instantiate a new one
	if _, found := rl.clients[ip]; !found {
		rl.clients[ip] = &client{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
	}

	// update the lastSeen property
	rl.clients[ip].lastSeen = time.Now()
	return rl.clients[ip].limiter
}

// cleanupClients is always running in a separate thread and is checking every 5 minutes
// for clients which weren't used recently so it can free up resources
func (rl *RateLimiter) cleanupClients() {
	for {
		time.Sleep(time.Minute * 5)

		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit represents the middleware method which is wraping the server executing
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// since there is a load balancer/gateway between the service and the client
		// we have to check for the value of the "X-Forwarded-For" header
		//
		// if we don't we are going to keep track only of the load balancer ip address
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		limiter := rl.getLimiter(clientIP)

		// if there aren't any tokens left in the bucket return error 429
		if !limiter.Allow() {
			_ = respond.TooManyRequests(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AttachRateLimitMiddleware(r *mux.Router, rl *RateLimiter) {
	r.Use(rl.Limit)
}
