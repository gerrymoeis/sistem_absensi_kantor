package service

import (
	"fmt"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"
)

type AdminService struct {
	adminRepo *repository.AdminRepository
	userRepo  *repository.UserRepository
}

func NewAdminService(adminRepo *repository.AdminRepository, userRepo *repository.UserRepository) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		userRepo:  userRepo,
	}
}

// GetStatistics returns dashboard statistics
func (s *AdminService) GetStatistics() (map[string]interface{}, error) {
	stats, err := s.adminRepo.GetStatistics()
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}
	return stats, nil
}

// GetAllAbsensi returns all attendance records
func (s *AdminService) GetAllAbsensi(limit, offset int, startDate, endDate string) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	records, err := s.adminRepo.GetAllAbsensi(limit, offset, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get absensi records: %w", err)
	}
	return records, nil
}

// GetTodayAbsensi returns today's attendance
func (s *AdminService) GetTodayAbsensi() ([]map[string]interface{}, error) {
	records, err := s.adminRepo.GetTodayAbsensi()
	if err != nil {
		return nil, fmt.Errorf("failed to get today absensi: %w", err)
	}
	return records, nil
}

// GetUserAbsensi returns attendance for specific user
func (s *AdminService) GetUserAbsensi(userID int64, limit, offset int) ([]model.Absensi, error) {
	if limit <= 0 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	records, err := s.adminRepo.GetUserAbsensi(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user absensi: %w", err)
	}
	return records, nil
}

// GetAllUsers returns all users (for admin)
func (s *AdminService) GetAllUsers() ([]model.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}
