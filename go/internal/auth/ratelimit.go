package auth

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides per-key token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*entry
	rate     rate.Limit
	burst    int
	ttl      time.Duration
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter that allows r requests per second
// with a burst capacity of b, per unique key.
func NewRateLimiter(r float64, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*entry),
		rate:     rate.Limit(r),
		burst:    b,
		ttl:      10 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

// Allow checks whether the given key is within its rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	e, ok := rl.limiters[key]
	if !ok {
		e = &entry{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[key] = e
	}
	e.lastSeen = time.Now()
	rl.mu.Unlock()

	return e.limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for key, e := range rl.limiters {
			if time.Since(e.lastSeen) > rl.ttl {
				delete(rl.limiters, key)
			}
		}
		rl.mu.Unlock()
	}
}

// LimitByIP returns HTTP middleware that rate limits requests by client IP.
func LimitByIP(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
				ip = fwd
			}
			if !limiter.Allow(ip) {
				http.Error(w, `{"code":"RESOURCE_EXHAUSTED","error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
