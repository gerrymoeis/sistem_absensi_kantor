package service

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"
)

type LeaveRequestService struct {
	leaveRepo   *repository.LeaveRequestRepository
	absensiRepo *repository.AbsensiRepository
}

func NewLeaveRequestService(leaveRepo *repository.LeaveRequestRepository, absensiRepo *repository.AbsensiRepository) *LeaveRequestService {
	return &LeaveRequestService{
		leaveRepo:   leaveRepo,
		absensiRepo: absensiRepo,
	}
}

// CreateLeaveRequest creates a new leave request
func (s *LeaveRequestService) CreateLeaveRequest(userID int64, req *model.CreateLeaveRequest) error {
	// Validate leave date (not in past)
	leaveDate, err := time.Parse("2006-01-02", req.LeaveDate)
	if err != nil {
		return fmt.Errorf("format tanggal tidak valid (gunakan YYYY-MM-DD)")
	}

	today := time.Now().Truncate(24 * time.Hour)
	if leaveDate.Before(today) {
		return fmt.Errorf("tanggal izin tidak boleh di masa lalu")
	}

	// Check duplicate request
	duplicate, err := s.leaveRepo.CheckDuplicate(userID, req.LeaveDate)
	if err != nil {
		return fmt.Errorf("failed to check duplicate: %w", err)
	}

	if duplicate {
		return fmt.Errorf("Anda sudah memiliki permohonan izin untuk tanggal ini")
	}

	// Validate proof file is provided (REQUIRED)
	if req.ProofFile == "" {
		return fmt.Errorf("bukti pendukung wajib dilampirkan")
	}

	// Save proof file (REQUIRED)
	path, err := s.saveProofFile(userID, req.ProofFile)
	if err != nil {
		return fmt.Errorf("failed to save proof file: %w", err)
	}
	proofFilePath := &path

	// Create leave request
	leaveRequest := &model.LeaveRequest{
		UserID:    userID,
		LeaveType: req.LeaveType,
		LeaveDate: req.LeaveDate,
		Reason:    req.Reason,
		ProofFile: proofFilePath,
		Status:    "pending",
	}

	if err := s.leaveRepo.Create(leaveRequest); err != nil {
		// Cleanup proof file if database insert fails
		if proofFilePath != nil {
			os.Remove(*proofFilePath)
		}
		return fmt.Errorf("failed to create leave request: %w", err)
	}

	return nil
}

// GetUserLeaveRequests gets all leave requests for a user
func (s *LeaveRequestService) GetUserLeaveRequests(userID int64, limit, offset int) ([]model.LeaveRequestWithUser, error) {
	if limit <= 0 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	return s.leaveRepo.FindByUserID(userID, limit, offset)
}

// GetLeaveRequestByID gets a specific leave request
func (s *LeaveRequestService) GetLeaveRequestByID(id int64, userID int64, isAdmin bool) (*model.LeaveRequestWithUser, error) {
	req, err := s.leaveRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req == nil {
		return nil, fmt.Errorf("permohonan izin tidak ditemukan")
	}

	// Check authorization (user can only see their own, admin can see all)
	if !isAdmin && req.UserID != userID {
		return nil, fmt.Errorf("Anda tidak memiliki akses ke permohonan ini")
	}

	return req, nil
}

// GetAllLeaveRequests gets all leave requests (admin only)
func (s *LeaveRequestService) GetAllLeaveRequests(status string, limit, offset int) ([]model.LeaveRequestWithUser, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.leaveRepo.FindAll(status, limit, offset)
}

// ReviewLeaveRequest reviews a leave request (approve/reject)
func (s *LeaveRequestService) ReviewLeaveRequest(id int64, reviewedBy int64, req *model.ReviewLeaveRequest) error {
	// Get leave request
	leaveReq, err := s.leaveRepo.FindByID(id)
	if err != nil {
		return err
	}

	if leaveReq == nil {
		return fmt.Errorf("permohonan izin tidak ditemukan")
	}

	// Check if already reviewed
	if leaveReq.Status != "pending" {
		return fmt.Errorf("permohonan ini sudah ditinjau sebelumnya")
	}

	// Update status
	if err := s.leaveRepo.UpdateStatus(id, req.Status, reviewedBy, req.ReviewNotes); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// If approved, create absensi record
	if req.Status == "approved" {
		if err := s.createAbsensiFromLeaveRequest(leaveReq); err != nil {
			// Log error but don't fail the approval
			// Admin can manually create absensi if needed
			fmt.Printf("Warning: failed to create absensi record: %v\n", err)
		}
	}

	return nil
}

// createAbsensiFromLeaveRequest creates an absensi record from approved leave request
func (s *LeaveRequestService) createAbsensiFromLeaveRequest(req *model.LeaveRequestWithUser) error {
	// Check if absensi already exists for this date
	existing, err := s.absensiRepo.FindByUserAndDate(req.UserID, req.LeaveDate)
	if err != nil {
		return fmt.Errorf("failed to check existing absensi: %w", err)
	}

	if existing != nil {
		// Already has absensi record, skip
		return nil
	}

	// Create absensi record with leave type as status
	keterangan := fmt.Sprintf("Permohonan %s disetujui: %s", req.LeaveType, req.Reason)
	absensi := &model.Absensi{
		UserID:     req.UserID,
		Tanggal:    req.LeaveDate,
		Status:     req.LeaveType, // izin, sakit, or cuti
		Keterangan: &keterangan,
	}

	if err := s.absensiRepo.Create(absensi); err != nil {
		return fmt.Errorf("failed to create absensi: %w", err)
	}

	return nil
}

// saveProofFile saves proof file from base64 data
func (s *LeaveRequestService) saveProofFile(userID int64, base64Data string) (string, error) {
	// Create uploads directory if not exists
	uploadsDir := "uploads/leave_proofs"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Parse base64 data (handle data URL format)
	var fileData []byte
	var fileExt string

	if strings.HasPrefix(base64Data, "data:") {
		// Extract MIME type and base64 data
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid base64 data format")
		}

		// Determine file extension from MIME type
		mimeType := parts[0]
		if strings.Contains(mimeType, "image/jpeg") || strings.Contains(mimeType, "image/jpg") {
			fileExt = ".jpg"
		} else if strings.Contains(mimeType, "image/png") {
			fileExt = ".png"
		} else if strings.Contains(mimeType, "application/pdf") {
			fileExt = ".pdf"
		} else {
			return "", fmt.Errorf("unsupported file type (only jpg, png, pdf allowed)")
		}

		// Decode base64
		decoded, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
		fileData = decoded
	} else {
		// Assume raw base64 without data URL prefix
		decoded, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64: %w", err)
		}
		fileData = decoded
		fileExt = ".jpg" // Default to jpg
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("user_%d_%d%s", userID, timestamp, fileExt)
	filePath := filepath.Join(uploadsDir, filename)

	// Write file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// DeleteLeaveRequest deletes a leave request (user can only delete pending requests)
func (s *LeaveRequestService) DeleteLeaveRequest(id int64, userID int64, isAdmin bool) error {
	// Get leave request
	req, err := s.leaveRepo.FindByID(id)
	if err != nil {
		return err
	}

	if req == nil {
		return fmt.Errorf("permohonan izin tidak ditemukan")
	}

	// Check authorization
	if !isAdmin && req.UserID != userID {
		return fmt.Errorf("Anda tidak memiliki akses untuk menghapus permohonan ini")
	}

	// User can only delete pending requests
	if !isAdmin && req.Status != "pending" {
		return fmt.Errorf("hanya permohonan yang masih pending yang dapat dihapus")
	}

	// Delete proof file if exists
	if req.ProofFile != nil && *req.ProofFile != "" {
		os.Remove(*req.ProofFile)
	}

	// Delete from database
	if err := s.leaveRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete leave request: %w", err)
	}

	return nil
}
