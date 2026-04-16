package service

import (
	"fmt"
	"time"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"
)

type AbsensiService struct {
	absensiRepo *repository.AbsensiRepository
	userRepo    *repository.UserRepository
}

func NewAbsensiService(absensiRepo *repository.AbsensiRepository, userRepo *repository.UserRepository) *AbsensiService {
	return &AbsensiService{
		absensiRepo: absensiRepo,
		userRepo:    userRepo,
	}
}

// ClockIn records clock in time
func (s *AbsensiService) ClockIn(userID int64, keterangan string) (*model.Absensi, error) {
	today := time.Now().Format("2006-01-02")
	now := time.Now().Format("15:04:05")

	// Check if already clocked in today
	existing, err := s.absensiRepo.FindByUserAndDate(userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing absensi: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("sudah absen masuk hari ini")
	}

	// Create new absensi record
	absensi := &model.Absensi{
		UserID:     userID,
		Tanggal:    today,
		JamMasuk:   &now,
		Status:     "hadir",
		Keterangan: nil,
	}

	if keterangan != "" {
		absensi.Keterangan = &keterangan
	}

	if err := s.absensiRepo.Create(absensi); err != nil {
		return nil, fmt.Errorf("failed to create absensi: %w", err)
	}

	return absensi, nil
}

// ClockOut records clock out time
func (s *AbsensiService) ClockOut(userID int64, keterangan string) (*model.Absensi, error) {
	today := time.Now().Format("2006-01-02")
	now := time.Now().Format("15:04:05")

	// Find today's absensi
	absensi, err := s.absensiRepo.FindByUserAndDate(userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing absensi: %w", err)
	}

	if absensi == nil {
		return nil, fmt.Errorf("belum absen masuk hari ini")
	}

	if absensi.JamPulang != nil {
		return nil, fmt.Errorf("sudah absen pulang hari ini")
	}

	// Update with clock out time
	absensi.JamPulang = &now
	if keterangan != "" {
		absensi.Keterangan = &keterangan
	}

	if err := s.absensiRepo.Update(absensi); err != nil {
		return nil, fmt.Errorf("failed to update absensi: %w", err)
	}

	return absensi, nil
}

// GetToday gets today's absensi for user
func (s *AbsensiService) GetToday(userID int64) (*model.Absensi, error) {
	today := time.Now().Format("2006-01-02")
	return s.absensiRepo.FindByUserAndDate(userID, today)
}

// GetHistory gets absensi history for user
func (s *AbsensiService) GetHistory(userID int64, limit, offset int) ([]model.Absensi, error) {
	if limit <= 0 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	return s.absensiRepo.FindByUserID(userID, limit, offset)
}
