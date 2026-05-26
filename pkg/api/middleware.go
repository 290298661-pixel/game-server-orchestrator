package api

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type Middleware struct {
	rateLimiter *RateLimiter
}

func NewMiddleware(requestsPerSecond int) *Middleware {
	return &Middleware{
		rateLimiter: NewRateLimiter(requestsPerSecond),
	}
}

func (mw *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Rate limiting.
		if mw.rateLimiter != nil && !mw.rateLimiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Logging.
		log.Printf("[INFO] [api] %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		log.Printf("[INFO] [api] %s %s → %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// RateLimiter is a simple token-bucket rate limiter.
type RateLimiter struct {
	rate     int
	interval time.Duration
	mu       sync.Mutex
	tokens   int
	lastFill time.Time
}

func NewRateLimiter(ratePerSecond int) *RateLimiter {
	if ratePerSecond <= 0 {
		ratePerSecond = 100
	}
	return &RateLimiter{
		rate:     ratePerSecond,
		interval: time.Second,
		tokens:   ratePerSecond,
		lastFill: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastFill).Seconds()
	newTokens := int(elapsed * float64(rl.rate))
	if newTokens > 0 {
		rl.tokens += newTokens
		if rl.tokens > rl.rate {
			rl.tokens = rl.rate
		}
		rl.lastFill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}
