package handler

import (
	"fmt"
	"net/http"

	"absensi-app/internal/middleware"
	"absensi-app/internal/model"
	"absensi-app/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	logService  *service.ActivityLogService
}

func NewAuthHandler(authService *service.AuthService, logService *service.ActivityLogService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logService:  logService,
	}
}

// LoginPage renders login page
func (h *AuthHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Sistem Absensi",
	})
}

// Login handles login request
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get client info for logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Authenticate user
	response, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// Log failed login attempt
		h.logService.LogFailed(nil, model.ActionLogin, 
			fmt.Sprintf("Failed login attempt for username: %s", req.Username), 
			ipAddress, userAgent)
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful login
	h.logService.LogSuccess(response.User.ID, model.ActionLogin, 
		fmt.Sprintf("User %s logged in successfully", response.User.Username), 
		ipAddress, userAgent)

	c.JSON(http.StatusOK, response)
}

// Logout handles logout request
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get user info for logging
	userID, _ := middleware.GetUserID(c)
	username, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Log logout
	h.logService.LogSuccess(userID, model.ActionLogout, 
		fmt.Sprintf("User %s logged out", username), 
		ipAddress, userAgent)

	// In stateless JWT, logout is handled client-side by removing token
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// Me returns current user info
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
