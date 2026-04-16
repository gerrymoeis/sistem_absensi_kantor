package middleware

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IPRestriction middleware checks if client IP is in allowed list
func IPRestriction(allowedCIDRs []string) gin.HandlerFunc {
	// Parse CIDR blocks once during initialization
	allowedNets := make([]*net.IPNet, 0, len(allowedCIDRs))
	for _, cidr := range allowedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			// Try parsing as single IP
			ip := net.ParseIP(cidr)
			if ip != nil {
				// Convert single IP to CIDR
				if ip.To4() != nil {
					_, ipNet, _ = net.ParseCIDR(cidr + "/32")
				} else {
					_, ipNet, _ = net.ParseCIDR(cidr + "/128")
				}
				allowedNets = append(allowedNets, ipNet)
			}
			continue
		}
		allowedNets = append(allowedNets, ipNet)
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		ip := net.ParseIP(clientIP)
		
		if ip == nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid IP address",
			})
			c.Abort()
			return
		}

		// Check if IP is in allowed networks
		allowed := false
		for _, ipNet := range allowedNets {
			if ipNet.Contains(ip) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: IP not allowed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
