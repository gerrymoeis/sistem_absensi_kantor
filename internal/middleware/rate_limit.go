package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

// Visitor represents a client with rate limit info
type Visitor struct {
	tokens     int
	lastSeen   time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per window
// window: time window duration (e.g., 1 minute)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup old visitors every 5 minutes
	go rl.cleanupVisitors()

	return rl
}

// RateLimit returns a middleware that limits requests per IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if the request is allowed
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			tokens:   rl.rate,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = visitor
	}
	rl.mu.Unlock()

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(visitor.lastSeen)
	
	if elapsed >= rl.window {
		// Full refill if window has passed
		visitor.tokens = rl.rate
		visitor.lastSeen = now
	} else {
		// Partial refill based on elapsed time
		tokensToAdd := int(float64(rl.rate) * (elapsed.Seconds() / rl.window.Seconds()))
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.rate {
			visitor.tokens = rl.rate
		}
		if tokensToAdd > 0 {
			visitor.lastSeen = now
		}
	}

	// Check if request is allowed
	if visitor.tokens > 0 {
		visitor.tokens--
		return true
	}

	return false
}

// cleanupVisitors removes old visitors to prevent memory leak
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			if now.Sub(visitor.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			visitor.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// LoginRateLimiter creates a stricter rate limiter for login endpoints
// Allows 5 requests per minute per IP
func LoginRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(5, 1*time.Minute)
	return limiter.RateLimit()
}

// APIRateLimiter creates a general rate limiter for API endpoints
// Allows 60 requests per minute per IP
func APIRateLimiter() gin.HandlerFunc {
	limiter := NewRateLimiter(60, 1*time.Minute)
	return limiter.RateLimit()
}
