package handler

import (
	"net/http"
	"strconv"

	"absensi-app/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
	logService   *service.ActivityLogService
}

func NewAdminHandler(adminService *service.AdminService, userService *service.UserService, logService *service.ActivityLogService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		userService:  userService,
		logService:   logService,
	}
}

// DashboardPage renders admin dashboard page
func (h *AdminHandler) DashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
		"title": "Admin Dashboard - Sistem Absensi",
	})
}

// GetStatistics returns dashboard statistics
func (h *AdminHandler) GetStatistics(c *gin.Context) {
	stats, err := h.adminService.GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetAllAbsensi returns all attendance records
func (h *AdminHandler) GetAllAbsensi(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	records, err := h.adminService.GetAllAbsensi(limit, offset, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
	})
}

// GetTodayAbsensi returns today's attendance
func (h *AdminHandler) GetTodayAbsensi(c *gin.Context) {
	records, err := h.adminService.GetTodayAbsensi()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
	})
}

// GetUserAbsensi returns attendance for specific user
func (h *AdminHandler) GetUserAbsensi(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	records, err := h.adminService.GetUserAbsensi(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": records,
	})
}

// GetAllUsers returns all users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.adminService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Remove password hash from response
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// GetActivityLogs returns activity logs
func (h *AdminHandler) GetActivityLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.logService.GetAllLogs(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}

// GetUserActivityLogs returns activity logs for specific user
func (h *AdminHandler) GetUserActivityLogs(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.logService.GetUserLogs(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": logs,
	})
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		FullName string `json:"full_name" binding:"required"`
		Role     string `json:"role" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := c.Get("user_id")
	adminUsername, _ := c.Get("username")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Create user
	user, err := h.userService.CreateUser(req.Username, req.Password, req.FullName, req.Role, req.IsActive)
	if err != nil {
		// Log failed attempt
		if id, ok := adminID.(int64); ok {
			h.logService.LogFailed(&id, "admin_create", 
				"Failed to create user "+req.Username+": "+err.Error(), 
				ipAddress, userAgent)
		}
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful creation
	if id, ok := adminID.(int64); ok {
		if username, ok := adminUsername.(string); ok {
			h.logService.LogSuccess(id, "admin_create", 
				"Admin "+username+" created user: "+user.Username, 
				ipAddress, userAgent)
		}
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    user,
	})
}

// GetUser returns user detail
func (h *AdminHandler) GetUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// UpdateUser updates user information
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req struct {
		FullName string `json:"full_name" binding:"required"`
		Role     string `json:"role" binding:"required"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := c.Get("user_id")
	adminUsername, _ := c.Get("username")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Update user
	user, err := h.userService.UpdateUser(userID, req.FullName, req.Role, req.IsActive)
	if err != nil {
		// Log failed attempt
		if id, ok := adminID.(int64); ok {
			h.logService.LogFailed(&id, "admin_update", 
				"Failed to update user ID "+strconv.FormatInt(userID, 10)+": "+err.Error(), 
				ipAddress, userAgent)
		}
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful update
	if id, ok := adminID.(int64); ok {
		if username, ok := adminUsername.(string); ok {
			h.logService.LogSuccess(id, "admin_update", 
				"Admin "+username+" updated user: "+user.Username, 
				ipAddress, userAgent)
		}
	}

	// Remove password hash from response
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user,
	})
}

// DeleteUser soft deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := c.Get("user_id")
	adminUsername, _ := c.Get("username")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Get user info before deletion
	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Prevent self-deletion
	if id, ok := adminID.(int64); ok && id == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(userID); err != nil {
		// Log failed attempt
		if id, ok := adminID.(int64); ok {
			h.logService.LogFailed(&id, "admin_delete", 
				"Failed to delete user "+user.Username+": "+err.Error(), 
				ipAddress, userAgent)
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful deletion
	if id, ok := adminID.(int64); ok {
		if username, ok := adminUsername.(string); ok {
			h.logService.LogSuccess(id, "admin_delete", 
				"Admin "+username+" deleted user: "+user.Username, 
				ipAddress, userAgent)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ResetPassword resets user password
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := c.Get("user_id")
	adminUsername, _ := c.Get("username")
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Get user info
	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Reset password
	if err := h.userService.ResetPassword(userID, req.NewPassword); err != nil {
		// Log failed attempt
		if id, ok := adminID.(int64); ok {
			h.logService.LogFailed(&id, "admin_update", 
				"Failed to reset password for user "+user.Username+": "+err.Error(), 
				ipAddress, userAgent)
		}
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful password reset
	if id, ok := adminID.(int64); ok {
		if username, ok := adminUsername.(string); ok {
			h.logService.LogSuccess(id, "admin_update", 
				"Admin "+username+" reset password for user: "+user.Username, 
				ipAddress, userAgent)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
