package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ysnarafat/tenantly/internal/database"
	"github.com/ysnarafat/tenantly/internal/models"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Clean old requests
	if requests, exists := rl.requests[key]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		rl.requests[key] = validRequests
	}

	// Check if limit exceeded
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// Add current request
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter, auditService *database.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.Allow(key) {
			// Log rate limit violation
			if auditService != nil {
				auditService.LogSystemAction(
					"RATE_LIMIT_EXCEEDED",
					"security",
					nil,
					nil,
					map[string]interface{}{
						"ip_address": c.ClientIP(),
						"user_agent": c.GetHeader("User-Agent"),
						"endpoint":   c.Request.URL.Path,
						"method":     c.Request.Method,
						"timestamp":  time.Now(),
					},
				)
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content security policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// SessionTimeoutMiddleware checks for session timeout
func SessionTimeoutMiddleware(auditService *database.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware runs after AuthRequired, so we have user context
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		// Check if this is a login endpoint (skip timeout check)
		if c.Request.URL.Path == "/api/auth/login" {
			c.Next()
			return
		}

		// Log user activity for session management
		if auditService != nil && userID != nil {
			if userIDInt, ok := userID.(int); ok {
				auditService.LogUserAction(
					userIDInt,
					"SESSION_ACTIVITY",
					models.TableUsers,
					nil,
					nil,
					map[string]interface{}{
						"endpoint":   c.Request.URL.Path,
						"method":     c.Request.Method,
						"ip_address": c.ClientIP(),
						"timestamp":  time.Now(),
					},
				)
			}
		}

		c.Next()
	}
}

// AuditMiddleware logs all API requests
func AuditMiddleware(auditService *database.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request after processing
		duration := time.Since(start)

		userID, _ := c.Get("user_id")
		var userIDInt *int
		if uid, ok := userID.(int); ok {
			userIDInt = &uid
		}

		if auditService != nil {
			if userIDInt != nil {
				auditService.LogUserAction(
					*userIDInt,
					"API_REQUEST",
					"api_requests",
					nil,
					nil,
					map[string]interface{}{
						"method":      c.Request.Method,
						"endpoint":    c.Request.URL.Path,
						"status_code": c.Writer.Status(),
						"duration_ms": duration.Milliseconds(),
						"ip_address":  c.ClientIP(),
						"user_agent":  c.GetHeader("User-Agent"),
						"timestamp":   start,
					},
				)
			} else {
				auditService.LogSystemAction(
					"API_REQUEST",
					"api_requests",
					nil,
					nil,
					map[string]interface{}{
						"method":      c.Request.Method,
						"endpoint":    c.Request.URL.Path,
						"status_code": c.Writer.Status(),
						"duration_ms": duration.Milliseconds(),
						"ip_address":  c.ClientIP(),
						"user_agent":  c.GetHeader("User-Agent"),
						"timestamp":   start,
					},
				)
			}
		}
	}
}
