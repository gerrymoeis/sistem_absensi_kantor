package handler

import (
	"fmt"
	"net/http"
	"time"

	"absensi-app/internal/middleware"
	"absensi-app/internal/service"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	exportService *service.ExportService
	logService    *service.ActivityLogService
}

func NewExportHandler(exportService *service.ExportService, logService *service.ActivityLogService) *ExportHandler {
	return &ExportHandler{
		exportService: exportService,
		logService:    logService,
	}
}

// ExportExcel exports attendance data to Excel
func (h *ExportHandler) ExportExcel(c *gin.Context) {
	// Get date range from query params
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	// Validate dates
	_, err1 := time.Parse("2006-01-02", startDate)
	_, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := middleware.GetUserID(c)
	adminUsername, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Generate Excel file
	file, err := h.exportService.ExportToExcel(startDate, endDate)
	if err != nil {
		// Log failed export
		h.logService.LogFailed(&adminID, "export_excel",
			fmt.Sprintf("Failed to export Excel (%s to %s): %s", startDate, endDate, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate Excel file: " + err.Error(),
		})
		return
	}

	// Log successful export
	h.logService.LogSuccess(adminID, "export_excel",
		fmt.Sprintf("Admin %s exported attendance data to Excel (%s to %s)", adminUsername, startDate, endDate),
		ipAddress, userAgent)

	// Set headers for file download
	filename := fmt.Sprintf("Laporan_Absensi_%s_to_%s.xlsx", startDate, endDate)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// Write file to response
	if err := file.Write(c.Writer); err != nil {
		h.logService.LogFailed(&adminID, "export_excel",
			fmt.Sprintf("Failed to write Excel file: %s", err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write Excel file",
		})
		return
	}
}


// ExportWord exports attendance data to Word
func (h *ExportHandler) ExportWord(c *gin.Context) {
	// Get date range from query params
	startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	// Validate dates
	_, err1 := time.Parse("2006-01-02", startDate)
	_, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	// Get admin info for logging
	adminID, _ := middleware.GetUserID(c)
	adminUsername, _ := middleware.GetUsername(c)
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Generate Word document
	doc, err := h.exportService.ExportToWord(startDate, endDate)
	if err != nil {
		// Log failed export
		h.logService.LogFailed(&adminID, "export_word",
			fmt.Sprintf("Failed to export Word (%s to %s): %s", startDate, endDate, err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate Word document: " + err.Error(),
		})
		return
	}

	// Log successful export
	h.logService.LogSuccess(adminID, "export_word",
		fmt.Sprintf("Admin %s exported attendance data to Word (%s to %s)", adminUsername, startDate, endDate),
		ipAddress, userAgent)

	// Set headers for file download
	filename := fmt.Sprintf("Laporan_Absensi_%s_to_%s.docx", startDate, endDate)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")

	// Write file to response
	if err := doc.Save(c.Writer); err != nil {
		h.logService.LogFailed(&adminID, "export_word",
			fmt.Sprintf("Failed to write Word file: %s", err.Error()),
			ipAddress, userAgent)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to write Word file",
		})
		return
	}
}
