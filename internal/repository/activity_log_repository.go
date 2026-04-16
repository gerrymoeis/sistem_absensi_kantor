package repository

import (
	"database/sql"
	"fmt"

	"absensi-app/internal/model"
)

type ActivityLogRepository struct {
	db *sql.DB
}

func NewActivityLogRepository(db *sql.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

// Create creates a new activity log
func (r *ActivityLogRepository) Create(log *model.ActivityLog) error {
	result, err := r.db.Exec(`
		INSERT INTO activity_logs (user_id, action_type, description, ip_address, user_agent, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, log.UserID, log.ActionType, log.Description, log.IPAddress, log.UserAgent, log.Status)

	if err != nil {
		return fmt.Errorf("failed to create activity log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	log.ID = id
	return nil
}

// FindByUserID finds activity logs by user ID
func (r *ActivityLogRepository) FindByUserID(userID int64, limit, offset int) ([]model.ActivityLog, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, action_type, description, ip_address, user_agent, status, created_at
		FROM activity_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	var logs []model.ActivityLog
	for rows.Next() {
		var log model.ActivityLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.ActionType,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.Status,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// FindAll finds all activity logs (for admin)
func (r *ActivityLogRepository) FindAll(limit, offset int) ([]model.ActivityLog, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, action_type, description, ip_address, user_agent, status, created_at
		FROM activity_logs
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query activity logs: %w", err)
	}
	defer rows.Close()

	var logs []model.ActivityLog
	for rows.Next() {
		var log model.ActivityLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.ActionType,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.Status,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}
