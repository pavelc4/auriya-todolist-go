package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.Mutex
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter creates a new IP rate limiter.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getVisitorLimiter returns the rate limiter for the given IP address.
func (i *IPRateLimiter) getVisitorLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.visitors[ip] = limiter
	}
	return limiter
}

// RateLimiter creates a middleware that limits requests per IP address.
func RateLimiter() gin.HandlerFunc {
	// Create a limiter that allows 2 requests per second with a burst of 4.
	// This is a much more reasonable limit for testing.
	limiter := NewIPRateLimiter(2, 4)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		visitorLimiter := limiter.getVisitorLimiter(ip)

		if !visitorLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too_many_requests"})
			return
		}
		c.Next()
	}
}
