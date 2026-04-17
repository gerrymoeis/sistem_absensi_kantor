package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ProductionLogger is a custom logger for production
// Logs requests without verbose debug info
func ProductionLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// Build query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Color codes for status
		statusColor := getStatusColor(statusCode)
		methodColor := getMethodColor(method)
		resetColor := "\033[0m"

		// Format: [GIN] timestamp | status | latency | clientIP | method | path
		fmt.Printf("[GIN] %s | %s%3d%s | %8s | %15s | %s%-7s%s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, resetColor,
			latency,
			clientIP,
			methodColor, method, resetColor,
			path,
		)
	}
}

// getStatusColor returns color code based on HTTP status
func getStatusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[32m" // Green
	case code >= 300 && code < 400:
		return "\033[36m" // Cyan
	case code >= 400 && code < 500:
		return "\033[33m" // Yellow
	default:
		return "\033[31m" // Red
	}
}

// getMethodColor returns color code based on HTTP method
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // Blue
	case "POST":
		return "\033[36m" // Cyan
	case "PUT":
		return "\033[33m" // Yellow
	case "DELETE":
		return "\033[31m" // Red
	default:
		return "\033[37m" // White
	}
}
