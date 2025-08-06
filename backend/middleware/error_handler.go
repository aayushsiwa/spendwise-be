package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware for centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			slog.Error("Panic recovered", 
				"error", err,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"type":    "panic",
					"message": "Internal server error",
				},
			})
		} else {
			slog.Error("Panic recovered with unknown error",
				"recovered", recovered,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
				"stack", string(debug.Stack()),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"type":    "panic",
					"message": "Internal server error",
				},
			})
		}
	})
}

// RequestLogger middleware for structured request logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Create log entry
		logEntry := slog.Info
		if statusCode >= 400 {
			logEntry = slog.Warn
		}
		if statusCode >= 500 {
			logEntry = slog.Error
		}

		// Log with structured fields
		logEntry("HTTP request",
			"method", method,
			"path", path,
			"query", raw,
			"status", statusCode,
			"latency", latency,
			"ip", clientIP,
			"user_agent", c.Request.UserAgent(),
			"errors", errorMessage,
		)
	}
}

// RateLimiter middleware for basic rate limiting
func RateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, you might want to use Redis or a more sophisticated solution
	clients := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old entries (older than 1 minute)
		if times, exists := clients[clientIP]; exists {
			var validTimes []time.Time
			for _, t := range times {
				if now.Sub(t) < time.Minute {
					validTimes = append(validTimes, t)
				}
			}
			clients[clientIP] = validTimes
		}
		
		// Check rate limit (100 requests per minute)
		if times, exists := clients[clientIP]; exists && len(times) >= 100 {
			slog.Warn("Rate limit exceeded", "ip", clientIP, "requests", len(times))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"type":    "rate_limit_exceeded",
					"message": "Too many requests",
				},
			})
			c.Abort()
			return
		}
		
		// Add current request
		clients[clientIP] = append(clients[clientIP], now)
		c.Next()
	}
}

// SecurityHeaders middleware for adding security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// ValidationErrorHandler middleware for handling validation errors
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for validation errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if validationErr, ok := err.Err.(errors.ValidationErrors); ok {
					errors.HandleValidationErrors(c, validationErr)
					return
				}
			}
		}
	}
} 