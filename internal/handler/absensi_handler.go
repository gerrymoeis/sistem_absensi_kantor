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

type AbsensiHandler struct {
	absensiService *service.AbsensiService
	logService     *service.ActivityLogService
}

func NewAbsensiHandler(absensiService *service.AbsensiService, logService *service.ActivityLogService) *AbsensiHandler {
	return &AbsensiHandler{
		absensiService: absensiService,
		logService:     logService,
	}
}

// DashboardPage renders dashboard page
func (h *AbsensiHandler) DashboardPage(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Dashboard - Sistem Absensi",
	})
}

// HistoryPage renders history page
func (h *AbsensiHandler) HistoryPage(c *gin.Context) {
	c.HTML(http.StatusOK, "history.html", gin.H{
		"title": "Riwayat Absensi",
	})
}

// ClockIn handles clock in request
func (h *AbsensiHandler) ClockIn(c *gin.Context) {
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

	var req model.ClockInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Keterangan is optional, so empty body is OK
		req.Keterangan = ""
	}

	absensi, err := h.absensiService.ClockIn(userID, req.Keterangan)
	if err != nil {
		// Log failed clock in
		h.logService.LogFailed(&userID, model.ActionClockIn, 
			fmt.Sprintf("Failed clock in for %s: %s", username, err.Error()), 
			ipAddress, userAgent)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful clock in
	h.logService.LogSuccess(userID, model.ActionClockIn, 
		fmt.Sprintf("User %s clocked in at %s", username, *absensi.JamMasuk), 
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil absen masuk",
		"data":    absensi,
	})
}

// ClockOut handles clock out request
func (h *AbsensiHandler) ClockOut(c *gin.Context) {
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

	var req model.ClockOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Keterangan is optional, so empty body is OK
		req.Keterangan = ""
	}

	absensi, err := h.absensiService.ClockOut(userID, req.Keterangan)
	if err != nil {
		// Log failed clock out
		h.logService.LogFailed(&userID, model.ActionClockOut, 
			fmt.Sprintf("Failed clock out for %s: %s", username, err.Error()), 
			ipAddress, userAgent)
		
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful clock out
	h.logService.LogSuccess(userID, model.ActionClockOut, 
		fmt.Sprintf("User %s clocked out at %s", username, *absensi.JamPulang), 
		ipAddress, userAgent)

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil absen pulang",
		"data":    absensi,
	})
}

// GetToday gets today's absensi
func (h *AbsensiHandler) GetToday(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	absensi, err := h.absensiService.GetToday(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": absensi,
	})
}

// GetHistory gets absensi history
func (h *AbsensiHandler) GetHistory(c *gin.Context) {
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

	history, err := h.absensiService.GetHistory(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": history,
	})
}
