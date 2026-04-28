package repository

import (
	"database/sql"
	"fmt"
	"time"

	"absensi-app/internal/model"
)

type LeaveRequestRepository struct {
	db *sql.DB
}

func NewLeaveRequestRepository(db *sql.DB) *LeaveRequestRepository {
	return &LeaveRequestRepository{db: db}
}

// Create creates a new leave request
func (r *LeaveRequestRepository) Create(req *model.LeaveRequest) error {
	result, err := r.db.Exec(`
		INSERT INTO leave_requests (user_id, leave_type, leave_date, reason, proof_file, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.UserID, req.LeaveType, req.LeaveDate, req.Reason, req.ProofFile, req.Status)

	if err != nil {
		return fmt.Errorf("failed to create leave request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	req.ID = id
	return nil
}

// FindByID finds a leave request by ID with user information
func (r *LeaveRequestRepository) FindByID(id int64) (*model.LeaveRequestWithUser, error) {
	req := &model.LeaveRequestWithUser{}
	
	err := r.db.QueryRow(`
		SELECT 
			lr.id, lr.user_id, lr.leave_type, lr.leave_date, lr.reason, lr.proof_file,
			lr.status, lr.reviewed_by, lr.reviewed_at, lr.review_notes,
			lr.created_at, lr.updated_at,
			u.full_name, u.username,
			CASE WHEN lr.reviewed_by IS NOT NULL THEN r.full_name ELSE NULL END as reviewer_name
		FROM leave_requests lr
		INNER JOIN users u ON lr.user_id = u.id
		LEFT JOIN users r ON lr.reviewed_by = r.id
		WHERE lr.id = ?
	`, id).Scan(
		&req.ID, &req.UserID, &req.LeaveType, &req.LeaveDate, &req.Reason, &req.ProofFile,
		&req.Status, &req.ReviewedBy, &req.ReviewedAt, &req.ReviewNotes,
		&req.CreatedAt, &req.UpdatedAt,
		&req.FullName, &req.Username, &req.ReviewerName,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find leave request: %w", err)
	}

	return req, nil
}

// FindByUserID finds all leave requests for a user
func (r *LeaveRequestRepository) FindByUserID(userID int64, limit, offset int) ([]model.LeaveRequestWithUser, error) {
	rows, err := r.db.Query(`
		SELECT 
			lr.id, lr.user_id, lr.leave_type, lr.leave_date, lr.reason, lr.proof_file,
			lr.status, lr.reviewed_by, lr.reviewed_at, lr.review_notes,
			lr.created_at, lr.updated_at,
			u.full_name, u.username,
			CASE WHEN lr.reviewed_by IS NOT NULL THEN r.full_name ELSE NULL END as reviewer_name
		FROM leave_requests lr
		INNER JOIN users u ON lr.user_id = u.id
		LEFT JOIN users r ON lr.reviewed_by = r.id
		WHERE lr.user_id = ?
		ORDER BY lr.created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query leave requests: %w", err)
	}
	defer rows.Close()

	var results []model.LeaveRequestWithUser
	for rows.Next() {
		var req model.LeaveRequestWithUser
		err := rows.Scan(
			&req.ID, &req.UserID, &req.LeaveType, &req.LeaveDate, &req.Reason, &req.ProofFile,
			&req.Status, &req.ReviewedBy, &req.ReviewedAt, &req.ReviewNotes,
			&req.CreatedAt, &req.UpdatedAt,
			&req.FullName, &req.Username, &req.ReviewerName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave request: %w", err)
		}
		results = append(results, req)
	}

	return results, nil
}

// FindAll finds all leave requests (admin)
func (r *LeaveRequestRepository) FindAll(status string, limit, offset int) ([]model.LeaveRequestWithUser, error) {
	query := `
		SELECT 
			lr.id, lr.user_id, lr.leave_type, lr.leave_date, lr.reason, lr.proof_file,
			lr.status, lr.reviewed_by, lr.reviewed_at, lr.review_notes,
			lr.created_at, lr.updated_at,
			u.full_name, u.username,
			CASE WHEN lr.reviewed_by IS NOT NULL THEN r.full_name ELSE NULL END as reviewer_name
		FROM leave_requests lr
		INNER JOIN users u ON lr.user_id = u.id
		LEFT JOIN users r ON lr.reviewed_by = r.id
	`

	var rows *sql.Rows
	var err error

	if status != "" {
		query += " WHERE lr.status = ? ORDER BY lr.created_at DESC LIMIT ? OFFSET ?"
		rows, err = r.db.Query(query, status, limit, offset)
	} else {
		query += " ORDER BY lr.created_at DESC LIMIT ? OFFSET ?"
		rows, err = r.db.Query(query, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query leave requests: %w", err)
	}
	defer rows.Close()

	var results []model.LeaveRequestWithUser
	for rows.Next() {
		var req model.LeaveRequestWithUser
		err := rows.Scan(
			&req.ID, &req.UserID, &req.LeaveType, &req.LeaveDate, &req.Reason, &req.ProofFile,
			&req.Status, &req.ReviewedBy, &req.ReviewedAt, &req.ReviewNotes,
			&req.CreatedAt, &req.UpdatedAt,
			&req.FullName, &req.Username, &req.ReviewerName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave request: %w", err)
		}
		results = append(results, req)
	}

	return results, nil
}

// UpdateStatus updates the status of a leave request
func (r *LeaveRequestRepository) UpdateStatus(id int64, status string, reviewedBy int64, reviewNotes string) error {
	now := time.Now()
	
	_, err := r.db.Exec(`
		UPDATE leave_requests
		SET status = ?, reviewed_by = ?, reviewed_at = ?, review_notes = ?, updated_at = ?
		WHERE id = ?
	`, status, reviewedBy, now, reviewNotes, now, id)

	if err != nil {
		return fmt.Errorf("failed to update leave request status: %w", err)
	}

	return nil
}

// CheckDuplicate checks if user already has a leave request for the same date
func (r *LeaveRequestRepository) CheckDuplicate(userID int64, leaveDate string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM leave_requests
		WHERE user_id = ? AND leave_date = ? AND status IN ('pending', 'approved')
	`, userID, leaveDate).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return count > 0, nil
}

// Delete deletes a leave request (soft delete by updating status)
func (r *LeaveRequestRepository) Delete(id int64) error {
	_, err := r.db.Exec(`
		DELETE FROM leave_requests WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete leave request: %w", err)
	}

	return nil
}
