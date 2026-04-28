package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// PageAuthRequired validates JWT token from cookie for HTML pages
// If invalid, redirects to login page instead of returning JSON
func PageAuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie
		tokenString, err := c.Cookie("token")
		
		if err != nil || tokenString == "" {
			// No token in cookie, redirect to login
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			// Invalid token, clear cookie and redirect
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Store user info in context
			c.Set("user_id", int64(claims["user_id"].(float64)))
			c.Set("username", claims["username"].(string))
			c.Set("role", claims["role"].(string))
		} else {
			// Invalid claims, clear cookie and redirect
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

// PageAdminRequired checks if user has admin role for HTML pages
// If not admin, redirects to dashboard instead of returning JSON
func PageAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			// Redirect non-admin users to their dashboard
			c.Redirect(http.StatusFound, "/dashboard")
			c.Abort()
			return
		}
		c.Next()
	}
}
