package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"absensi-app/internal/middleware"
	"absensi-app/internal/model"
	"absensi-app/internal/service"

	"github.com/gin-gonic/gin"
)

type LeaveRequestHandler struct {
	leaveService *service.LeaveRequestService
	logService   *service.ActivityLogService
}

func NewLeaveRequestHandler(leaveService *service.LeaveRequestService, logService *service.ActivityLogService) *LeaveRequestHandler {
	return &LeaveRequestHandler{
		leaveService: leaveService,
		logService:   logService,
	}
}

// Create handles leave request creation
func (h *LeaveRequestHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	username, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	var req model.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Data tidak valid: " + err.Error(),
		})
		return
	}

	if err := h.leaveService.CreateLeaveRequest(userID, &req); err != nil {
		// Log failed request
		h.logService.LogFailed(&userID, "leave_request_create",
			fmt.Sprintf("Failed to create leave request for %s: %s", username, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful request
	h.logService.LogSuccess(userID, "leave_request_create",
		fmt.Sprintf("User %s created leave request for %s (%s)", username, req.LeaveDate, req.LeaveType),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Permohonan izin berhasil diajukan. Menunggu persetujuan admin.",
	})
}

// GetUserRequests gets all leave requests for the authenticated user
func (h *LeaveRequestHandler) GetUserRequests(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse pagination params
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	requests, err := h.leaveService.GetUserLeaveRequests(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": requests,
	})
}

// GetByID gets a specific leave request
func (h *LeaveRequestHandler) GetByID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	role, _ := middleware.GetRole(c)
	isAdmin := role == "admin"

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tidak valid",
		})
		return
	}

	request, err := h.leaveService.GetLeaveRequestByID(id, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": request,
	})
}

// GetAllRequests gets all leave requests (admin only)
func (h *LeaveRequestHandler) GetAllRequests(c *gin.Context) {
	// Parse query params
	status := c.DefaultQuery("status", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	requests, err := h.leaveService.GetAllLeaveRequests(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": requests,
	})
}

// Review handles leave request review (approve/reject) - admin only
func (h *LeaveRequestHandler) Review(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	username, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tidak valid",
		})
		return
	}

	var req model.ReviewLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Data tidak valid: " + err.Error(),
		})
		return
	}

	if err := h.leaveService.ReviewLeaveRequest(id, userID, &req); err != nil {
		// Log failed review
		h.logService.LogFailed(&userID, "leave_request_review",
			fmt.Sprintf("Failed to review leave request #%d by %s: %s", id, username, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful review
	action := "disetujui"
	if req.Status == "rejected" {
		action = "ditolak"
	}
	h.logService.LogSuccess(userID, "leave_request_review",
		fmt.Sprintf("Admin %s %s leave request #%d", username, action, id),
		ipAddress, userAgent)

	message := "Permohonan berhasil disetujui"
	if req.Status == "rejected" {
		message = "Permohonan berhasil ditolak"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// Delete handles leave request deletion
func (h *LeaveRequestHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	role, _ := middleware.GetRole(c)
	isAdmin := role == "admin"
	username, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tidak valid",
		})
		return
	}

	if err := h.leaveService.DeleteLeaveRequest(id, userID, isAdmin); err != nil {
		// Log failed deletion
		h.logService.LogFailed(&userID, "leave_request_delete",
			fmt.Sprintf("Failed to delete leave request #%d by %s: %s", id, username, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful deletion
	h.logService.LogSuccess(userID, "leave_request_delete",
		fmt.Sprintf("User %s deleted leave request #%d", username, id),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Permohonan berhasil dihapus",
	})
}
