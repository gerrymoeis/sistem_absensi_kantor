package service

import (
	"fmt"
	"time"

	"absensi-app/internal/repository"

	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
	"github.com/xuri/excelize/v2"
)

type ExportService struct {
	adminRepo *repository.AdminRepository
}

func NewExportService(adminRepo *repository.AdminRepository) *ExportService {
	return &ExportService{adminRepo: adminRepo}
}

// ExportToExcel exports attendance data to Excel format
func (s *ExportService) ExportToExcel(startDate, endDate string) (*excelize.File, error) {
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

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 5)   // No
	f.SetColWidth(sheetName, "B", "B", 20)  // Nama
	f.SetColWidth(sheetName, "C", "C", 12)  // Tanggal
	f.SetColWidth(sheetName, "D", "D", 10)  // Jam Masuk
	f.SetColWidth(sheetName, "E", "E", 10)  // Jam Pulang
	f.SetColWidth(sheetName, "F", "F", 12)  // Durasi
	f.SetColWidth(sheetName, "G", "G", 15)  // Keterangan

	// Create header style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
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
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// Create title style
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

	// Create data style
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create data style: %w", err)
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

	// Write headers
	headers := []string{"No", "Nama Karyawan", "Tanggal", "Jam Masuk", "Jam Pulang", "Durasi", "Keterangan"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 4)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	f.SetRowHeight(sheetName, 4, 20)

	// Get data from repository
	records, err := s.adminRepo.GetAllAbsensi(1000, 0, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance data: %w", err)
	}

	// Write data
	row := 5
	for i, record := range records {
		// No
		cell, _ := excelize.CoordinatesToCellName(1, row)
		f.SetCellValue(sheetName, cell, i+1)
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Nama Karyawan
		cell, _ = excelize.CoordinatesToCellName(2, row)
		f.SetCellValue(sheetName, cell, record["full_name"])
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Tanggal
		cell, _ = excelize.CoordinatesToCellName(3, row)
		f.SetCellValue(sheetName, cell, record["tanggal"])
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Jam Masuk
		cell, _ = excelize.CoordinatesToCellName(4, row)
		if record["jam_masuk"] != nil {
			f.SetCellValue(sheetName, cell, record["jam_masuk"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Jam Pulang
		cell, _ = excelize.CoordinatesToCellName(5, row)
		if record["jam_pulang"] != nil {
			f.SetCellValue(sheetName, cell, record["jam_pulang"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Durasi
		cell, _ = excelize.CoordinatesToCellName(6, row)
		durasi := s.calculateDuration(record["jam_masuk"], record["jam_pulang"])
		f.SetCellValue(sheetName, cell, durasi)
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		// Keterangan
		cell, _ = excelize.CoordinatesToCellName(7, row)
		if record["keterangan"] != nil {
			f.SetCellValue(sheetName, cell, record["keterangan"])
		} else {
			f.SetCellValue(sheetName, cell, "-")
		}
		f.SetCellStyle(sheetName, cell, cell, dataStyle)

		row++
	}

	// Write footer
	footerRow := row + 1
	footerText := fmt.Sprintf("Total: %d record | Digenerate: %s", len(records), time.Now().Format("02 Jan 2006 15:04"))
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", footerRow), footerText)
	f.MergeCell(sheetName, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("G%d", footerRow))

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


// ExportToWord exports attendance data to Word format
func (s *ExportService) ExportToWord(startDate, endDate string) (*document.Document, error) {
	// Create new Word document
	doc := document.New()

	// Add title
	para := doc.AddParagraph()
	run := para.AddRun()
	run.AddText("LAPORAN ABSENSI KARYAWAN")
	run.Properties().SetBold(true)
	run.Properties().SetSize(16)
	para.Properties().SetAlignment(wml.ST_JcCenter)

	// Add empty line
	doc.AddParagraph()

	// Add period
	para = doc.AddParagraph()
	run = para.AddRun()
	period := fmt.Sprintf("Periode: %s s/d %s", startDate, endDate)
	run.AddText(period)
	run.Properties().SetBold(true)
	run.Properties().SetSize(12)
	para.Properties().SetAlignment(wml.ST_JcCenter)

	// Add empty line
	doc.AddParagraph()

	// Get data from repository
	records, err := s.adminRepo.GetAllAbsensi(1000, 0, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance data: %w", err)
	}

	// Create table
	table := doc.AddTable()
	table.Properties().SetWidthPercent(100)

	// Set table borders
	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, measurement.Point)

	// Add header row
	headers := []string{"No", "Nama Karyawan", "Tanggal", "Jam Masuk", "Jam Pulang", "Durasi", "Keterangan"}
	headerRow := table.AddRow()
	for _, header := range headers {
		cell := headerRow.AddCell()
		para := cell.AddParagraph()
		run := para.AddRun()
		run.AddText(header)
		run.Properties().SetBold(true)
		para.Properties().SetAlignment(wml.ST_JcCenter)
	}

	// Add data rows
	for i, record := range records {
		row := table.AddRow()

		// No
		cell := row.AddCell()
		para := cell.AddParagraph()
		para.AddRun().AddText(fmt.Sprintf("%d", i+1))
		para.Properties().SetAlignment(wml.ST_JcCenter)

		// Nama Karyawan
		cell = row.AddCell()
		para = cell.AddParagraph()
		fullName := ""
		if name, ok := record["full_name"].(string); ok {
			fullName = name
		}
		para.AddRun().AddText(fullName)

		// Tanggal
		cell = row.AddCell()
		para = cell.AddParagraph()
		tanggal := ""
		if date, ok := record["tanggal"].(string); ok {
			tanggal = date
		}
		para.AddRun().AddText(tanggal)
		para.Properties().SetAlignment(wml.ST_JcCenter)

		// Jam Masuk
		cell = row.AddCell()
		para = cell.AddParagraph()
		jamMasuk := "-"
		if record["jam_masuk"] != nil {
			if jm, ok := record["jam_masuk"].(string); ok {
				jamMasuk = jm
			}
		}
		para.AddRun().AddText(jamMasuk)
		para.Properties().SetAlignment(wml.ST_JcCenter)

		// Jam Pulang
		cell = row.AddCell()
		para = cell.AddParagraph()
		jamPulang := "-"
		if record["jam_pulang"] != nil {
			if jp, ok := record["jam_pulang"].(string); ok {
				jamPulang = jp
			}
		}
		para.AddRun().AddText(jamPulang)
		para.Properties().SetAlignment(wml.ST_JcCenter)

		// Durasi
		cell = row.AddCell()
		para = cell.AddParagraph()
		durasi := s.calculateDuration(record["jam_masuk"], record["jam_pulang"])
		para.AddRun().AddText(durasi)
		para.Properties().SetAlignment(wml.ST_JcCenter)

		// Keterangan
		cell = row.AddCell()
		para = cell.AddParagraph()
		keterangan := "-"
		if record["keterangan"] != nil {
			if ket, ok := record["keterangan"].(string); ok {
				keterangan = ket
			}
		}
		para.AddRun().AddText(keterangan)
	}

	// Add footer
	doc.AddParagraph()
	para = doc.AddParagraph()
	run = para.AddRun()
	footerText := fmt.Sprintf("Total: %d record", len(records))
	run.AddText(footerText)
	run.Properties().SetItalic(true)

	para = doc.AddParagraph()
	run = para.AddRun()
	generatedText := fmt.Sprintf("Digenerate: %s", time.Now().Format("02 January 2006 15:04"))
	run.AddText(generatedText)
	run.Properties().SetItalic(true)

	// Add signature section
	doc.AddParagraph()
	doc.AddParagraph()

	// Create signature table
	sigTable := doc.AddTable()
	sigTable.Properties().SetWidthPercent(100)
	sigTable.Properties().Borders().SetAll(wml.ST_BorderNone, color.Auto, 0)

	// Row 1: Titles
	row := sigTable.AddRow()
	for _, title := range []string{"Mengetahui,", "Menyetujui,", "Dibuat oleh,"} {
		cell := row.AddCell()
		para := cell.AddParagraph()
		para.AddRun().AddText(title)
		para.Properties().SetAlignment(wml.ST_JcCenter)
	}

	// Row 2: Positions
	row = sigTable.AddRow()
	for _, position := range []string{"Manager", "HRD", "Admin"} {
		cell := row.AddCell()
		para := cell.AddParagraph()
		para.AddRun().AddText(position)
		para.Properties().SetAlignment(wml.ST_JcCenter)
	}

	// Row 3: Empty space for signature
	row = sigTable.AddRow()
	for i := 0; i < 3; i++ {
		cell := row.AddCell()
		para := cell.AddParagraph()
		para.AddRun().AddText("")
		para.AddRun().AddBreak()
		para.AddRun().AddBreak()
	}

	// Row 4: Name lines
	row = sigTable.AddRow()
	for i := 0; i < 3; i++ {
		cell := row.AddCell()
		para := cell.AddParagraph()
		para.AddRun().AddText("(_______________)")
		para.Properties().SetAlignment(wml.ST_JcCenter)
	}

	return doc, nil
}
