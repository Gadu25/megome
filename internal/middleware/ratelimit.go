package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"megome/internal/pkg/httputil"

	"golang.org/x/time/rate"
)

const rateLimiterCleanupInterval = 1 * time.Minute
const rateLimiterStaleTimeout = 3 * time.Minute

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	stop    chan struct{}

	rate  rate.Limit
	burst int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		stop:    make(chan struct{}),
		rate:    r,
		burst:   burst,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rateLimiterCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.evictStaleClients()
		case <-rl.stop:
			return
		}
	}
}

func (rl *RateLimiter) evictStaleClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rateLimiterStaleTimeout)
	for ip, c := range rl.clients {
		if c.lastSeen.Before(cutoff) {
			delete(rl.clients, ip)
		}
	}
}

func (rl *RateLimiter) getClient(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.clients[ip] = &client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	c.lastSeen = time.Now()
	return c.limiter
}

func getIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getClient(ip)

		if !limiter.Allow() {
			httputil.WriteError(w, http.StatusTooManyRequests, fmt.Errorf("rate limit exceeded"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
