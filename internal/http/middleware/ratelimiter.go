package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Initialize a global limiter that allows 40 requests per second, with a burst of 60.
var globalLimiter = rate.NewLimiter(40, 60)

// RateLimiter creates a middleware that limits the number of requests using a global limiter.
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Rate limiter middleware is executing.")
		if !globalLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "Too many requests"})
			return
		}
		c.Next()
	}
}
