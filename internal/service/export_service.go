package service

import (
	"fmt"
	"time"

	"absensi-app/internal/repository"

	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	adminRepo *repository.AdminRepository
}

func NewExportService(adminRepo *repository.AdminRepository) *ExportService {
	return &ExportService{adminRepo: adminRepo}
}

// ExportToExcel exports attendance data to Excel format with professional formatting
func (s *ExportService) ExportToExcel(startDate, endDate string) (*excelize.File, error) {
	// Get data from repository FIRST
	records, err := s.adminRepo.GetAllAbsensi(10000, 0, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance data: %w", err)
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create sheet
	sheetName := "Laporan Absensi"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1") // Remove default sheet

	// Set column widths for better readability
	f.SetColWidth(sheetName, "A", "A", 6)   // No
	f.SetColWidth(sheetName, "B", "B", 25)  // Nama Karyawan
	f.SetColWidth(sheetName, "C", "C", 13)  // Tanggal
	f.SetColWidth(sheetName, "D", "D", 12)  // Jam Masuk
	f.SetColWidth(sheetName, "E", "E", 12)  // Jam Pulang
	f.SetColWidth(sheetName, "F", "F", 15)  // Durasi
	f.SetColWidth(sheetName, "G", "G", 20)  // Keterangan

	// Create title style (bold, centered, larger font)
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create title style: %w", err)
	}

	// Create header style (bold, white text, blue background, bordered, centered)
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  11,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// Create data style with borders and center alignment for specific columns
	dataCenterStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create data center style: %w", err)
	}

	// Create data style with borders and left alignment for text columns
	dataLeftStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create data left style: %w", err)
	}

	// Create footer style (italic, smaller font)
	footerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Italic: true,
			Size:   10,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create footer style: %w", err)
	}

	// Write title
	f.SetCellValue(sheetName, "A1", "LAPORAN ABSENSI KARYAWAN")
	f.MergeCell(sheetName, "A1", "G1")
	f.SetCellStyle(sheetName, "A1", "G1", titleStyle)
	f.SetRowHeight(sheetName, 1, 25)

	// Write period
	period := fmt.Sprintf("Periode: %s s/d %s", startDate, endDate)
	f.SetCellValue(sheetName, "A2", period)
	f.MergeCell(sheetName, "A2", "G2")
	f.SetCellStyle(sheetName, "A2", "G2", titleStyle)
	f.SetRowHeight(sheetName, 2, 20)

	// Add empty row for spacing
	f.SetRowHeight(sheetName, 3, 5)

	// Write headers
	headers := []string{"No", "Nama Karyawan", "Tanggal", "Jam Masuk", "Jam Pulang", "Durasi", "Keterangan"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 4, 22)

	// Write data rows
	row := 5
	for i, record := range records {
		// No - centered
		cell, _ := excelize.CoordinatesToCellName(1, row)
		f.SetCellValue(sheetName, cell, i+1)
		f.SetCellStyle(sheetName, cell, cell, dataCenterStyle)

		// Nama Karyawan - left aligned
		cell, _ = excelize.CoordinatesToCellName(2, row)
		f.SetCellValue(sheetName, cell, record["full_name"])
		f.SetCellStyle(sheetName, cell, cell, dataLeftStyle)

		// Tanggal - centered
		cell, _ = excelize.CoordinatesToCellName(3, row)
		f.SetCellValue(sheetName, cell, record["tanggal"])
		f.SetCellStyle(sheetName, cell, cell, dataCenterStyle)

		// Jam Masuk - centered
		cell, _ = excelize.CoordinatesToCellName(4, row)
		if record["jam_masuk"] != nil {
			f.SetCellValue(sheetName, cell, record["jam_masuk"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataCenterStyle)

		// Jam Pulang - centered
		cell, _ = excelize.CoordinatesToCellName(5, row)
		if record["jam_pulang"] != nil {
			f.SetCellValue(sheetName, cell, record["jam_pulang"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataCenterStyle)

		// Durasi - centered
		cell, _ = excelize.CoordinatesToCellName(6, row)
		durasi := s.calculateDuration(record["jam_masuk"], record["jam_pulang"])
		f.SetCellValue(sheetName, cell, durasi)
		f.SetCellStyle(sheetName, cell, cell, dataCenterStyle)

		// Keterangan - left aligned
		cell, _ = excelize.CoordinatesToCellName(7, row)
		if record["keterangan"] != nil {
			f.SetCellValue(sheetName, cell, record["keterangan"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataLeftStyle)

		row++
	}

	// Add empty row for spacing before footer
	row++

	// Write footer with better formatting
	footerRow := row
	totalText := fmt.Sprintf("Total Data: %d record", len(records))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow), totalText)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("A%d", footerRow), footerStyle)

	// Write generation timestamp on next line
	footerRow++
	generatedText := fmt.Sprintf("Digenerate pada: %s", time.Now().Format("02 January 2006, 15:04:05 WIB"))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow), generatedText)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("A%d", footerRow), footerStyle)

	return f, nil
}

// calculateDuration calculates duration between clock in and clock out
func (s *ExportService) calculateDuration(jamMasuk, jamPulang interface{}) string {
	if jamMasuk == nil || jamPulang == nil {
		return "-"
	}

	masukStr, ok1 := jamMasuk.(string)
	pulangStr, ok2 := jamPulang.(string)
	if !ok1 || !ok2 {
		return "-"
	}

	// Parse times (format: HH:MM:SS)
	masuk, err1 := time.Parse("15:04:05", masukStr)
	pulang, err2 := time.Parse("15:04:05", pulangStr)
	if err1 != nil || err2 != nil {
		return "-"
	}

	// Calculate duration
	duration := pulang.Sub(masuk)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	return fmt.Sprintf("%d jam %d menit", hours, minutes)
}
