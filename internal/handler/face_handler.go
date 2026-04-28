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

type FaceHandler struct {
	faceService *service.FaceService
	logService  *service.ActivityLogService
}

func NewFaceHandler(faceService *service.FaceService, logService *service.ActivityLogService) *FaceHandler {
	return &FaceHandler{
		faceService: faceService,
		logService:  logService,
	}
}

// SelfEnroll handles self face enrollment (employee)
func (h *FaceHandler) SelfEnroll(c *gin.Context) {
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

	var req model.SelfEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Enroll face for current user
	err := h.faceService.EnrollFace(userID, req.ImageData)
	if err != nil {
		// Log failed attempt
		h.logService.LogFailed(&userID, "face_enrollment",
			fmt.Sprintf("User %s failed to enroll face: %s", username, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful enrollment
	h.logService.LogSuccess(userID, "face_enrollment",
		fmt.Sprintf("User %s successfully enrolled face", username),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Wajah berhasil didaftarkan",
	})
}

// SelfEnrollComprehensive handles comprehensive self face enrollment with 5 photos (employee)
func (h *FaceHandler) SelfEnrollComprehensive(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Pengguna tidak terautentikasi",
		})
		return
	}

	username, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	var req model.ComprehensiveEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format permintaan tidak valid",
		})
		return
	}

	// Validate exactly 5 photos
	if len(req.Photos) != 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Diperlukan tepat 5 foto",
		})
		return
	}

	// Enroll comprehensive face for current user
	err := h.faceService.EnrollComprehensive(userID, req.Photos)
	if err != nil {
		// Log failed attempt
		h.logService.LogFailed(&userID, "face_enrollment",
			fmt.Sprintf("User %s failed comprehensive face enrollment: %s", username, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful enrollment
	h.logService.LogSuccess(userID, "face_enrollment",
		fmt.Sprintf("User %s successfully enrolled face (comprehensive: 5 photos)", username),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Wajah berhasil didaftarkan dengan lengkap",
		"photos_count":   len(req.Photos),
		"enrollment_type": "comprehensive",
	})
}

// EnrollFace handles face enrollment (admin only)
func (h *FaceHandler) EnrollFace(c *gin.Context) {
	var req model.FaceEnrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := middleware.GetUserID(c)
	adminUsername, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Enroll face
	err := h.faceService.EnrollFace(req.UserID, req.ImageData)
	if err != nil {
		// Log failed attempt
		h.logService.LogFailed(&adminID, "admin_create",
			fmt.Sprintf("Admin %s failed to enroll face for user ID %d: %s", 
				adminUsername, req.UserID, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful enrollment
	h.logService.LogSuccess(adminID, "admin_create",
		fmt.Sprintf("Admin %s enrolled face for user ID %d", adminUsername, req.UserID),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Face enrolled successfully",
	})
}

// RecognizeFace handles face recognition (employee)
func (h *FaceHandler) RecognizeFace(c *gin.Context) {
	var req model.FaceRecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()

	// Recognize face
	result, err := h.faceService.RecognizeFace(req.ImageData, ipAddress, req.LivenessPassed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// If matched, log the recognition
	if result.Matched && result.UserID != nil {
		userAgent := c.GetHeader("User-Agent")
		h.logService.LogSuccess(*result.UserID, "face_recognition",
			fmt.Sprintf("Face recognized for user: %s (confidence: %.2f%%)", 
				result.FullName, result.Confidence*100),
			ipAddress, userAgent)
	}

	c.JSON(http.StatusOK, result)
}

// DeleteUserFaceData deletes all face data for a user (admin only)
func (h *FaceHandler) DeleteUserFaceData(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := middleware.GetUserID(c)
	adminUsername, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Delete face data
	err = h.faceService.DeleteUserFaceData(userID)
	if err != nil {
		// Log failed attempt
		h.logService.LogFailed(&adminID, "admin_delete",
			fmt.Sprintf("Admin %s failed to delete face data for user ID %d: %s", 
				adminUsername, userID, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful deletion
	h.logService.LogSuccess(adminID, "admin_delete",
		fmt.Sprintf("Admin %s deleted face data for user ID %d", adminUsername, userID),
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Face data deleted successfully",
	})
}

// GetEncodingStats returns face encoding statistics (admin only)
func (h *FaceHandler) GetEncodingStats(c *gin.Context) {
	stats, err := h.faceService.GetEncodingStats()
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

// CheckEnrollmentStatus checks if current user has face enrolled
func (h *FaceHandler) CheckEnrollmentStatus(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	encodings, err := h.faceService.GetUserEncodings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enrolled":       len(encodings) > 0,
		"encoding_count": len(encodings),
	})
}
