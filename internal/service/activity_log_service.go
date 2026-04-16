package service

import (
	"fmt"

	"absensi-app/internal/model"
	"absensi-app/internal/repository"
)

type ActivityLogService struct {
	logRepo *repository.ActivityLogRepository
}

func NewActivityLogService(logRepo *repository.ActivityLogRepository) *ActivityLogService {
	return &ActivityLogService{logRepo: logRepo}
}

// Log creates a new activity log entry
func (s *ActivityLogService) Log(userID *int64, actionType, description, ipAddress, userAgent, status string) error {
	log := &model.ActivityLog{
		UserID:      userID,
		ActionType:  actionType,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Status:      status,
	}

	if err := s.logRepo.Create(log); err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

// LogSuccess logs a successful action
func (s *ActivityLogService) LogSuccess(userID int64, actionType, description, ipAddress, userAgent string) error {
	return s.Log(&userID, actionType, description, ipAddress, userAgent, model.StatusSuccess)
}

// LogFailed logs a failed action
func (s *ActivityLogService) LogFailed(userID *int64, actionType, description, ipAddress, userAgent string) error {
	return s.Log(userID, actionType, description, ipAddress, userAgent, model.StatusFailed)
}

// GetUserLogs gets activity logs for a specific user
func (s *ActivityLogService) GetUserLogs(userID int64, limit, offset int) ([]model.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.logRepo.FindByUserID(userID, limit, offset)
}

// GetAllLogs gets all activity logs (for admin)
func (s *ActivityLogService) GetAllLogs(limit, offset int) ([]model.ActivityLog, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.logRepo.FindAll(limit, offset)
}
