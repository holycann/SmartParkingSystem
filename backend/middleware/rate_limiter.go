package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements a token bucket rate limiter for API requests
type RateLimiter struct {
	// Map of IP addresses to limiters
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	// Rate at which tokens are added to the bucket
	rate rate.Limit
	// Maximum number of tokens in the bucket
	burst int
	// Duration to keep limiters in memory
	expiryTime time.Duration
	// Map of IP addresses to last seen times
	lastSeen map[string]time.Time
}

// NewRateLimiter creates a new rate limiter with the specified requests per duration
func NewRateLimiter(requestsPerDuration int, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters:   make(map[string]*rate.Limiter),
		rate:       rate.Limit(float64(requestsPerDuration) / duration.Seconds()),
		burst:      requestsPerDuration,
		expiryTime: time.Hour, // Clean up limiters after 1 hour of inactivity
		lastSeen:   make(map[string]time.Time),
	}
}

// getLimiter returns the rate limiter for the provided IP address
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Create a new limiter if one doesn't exist
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	// Update last seen time
	rl.lastSeen[ip] = time.Now()
	return limiter
}

// cleanupLimiters removes limiters for IPs that haven't been seen recently
func (rl *RateLimiter) cleanupLimiters() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, lastSeen := range rl.lastSeen {
		if now.Sub(lastSeen) > rl.expiryTime {
			delete(rl.limiters, ip)
			delete(rl.lastSeen, ip)
		}
	}
}

// Middleware returns a Gin middleware function that applies rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	// Start a goroutine to periodically clean up old limiters
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			rl.cleanupLimiters()
		}
	}()

	return func(c *gin.Context) {
		// Get client IP address
		ip := c.ClientIP()

		// Get the limiter for this IP
		limiter := rl.getLimiter(ip)

		// Check if the request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		// Continue processing the request
		c.Next()
	}
}
