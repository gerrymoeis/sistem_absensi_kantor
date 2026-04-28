package repository

import (
	"database/sql"
	"fmt"
	"time"

	"absensi-app/internal/model"
)

type FaceRepository struct {
	db *sql.DB
}

func NewFaceRepository(db *sql.DB) *FaceRepository {
	return &FaceRepository{db: db}
}

// SaveEncoding saves a face encoding to the database
func (r *FaceRepository) SaveEncoding(encoding *model.FaceEncoding) error {
	query := `
		INSERT INTO face_encodings (user_id, encoding, quality_score, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(query, encoding.UserID, encoding.Encoding, encoding.QualityScore)
	if err != nil {
		return fmt.Errorf("failed to save encoding: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	encoding.ID = id
	return nil
}

// GetUserEncodings retrieves all face encodings for a user
func (r *FaceRepository) GetUserEncodings(userID int64) ([]model.FaceEncoding, error) {
	query := `
		SELECT id, user_id, encoding, quality_score, created_at
		FROM face_encodings
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query encodings: %w", err)
	}
	defer rows.Close()

	var encodings []model.FaceEncoding
	for rows.Next() {
		var enc model.FaceEncoding
		err := rows.Scan(&enc.ID, &enc.UserID, &enc.Encoding, &enc.QualityScore, &enc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan encoding: %w", err)
		}
		encodings = append(encodings, enc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return encodings, nil
}

// GetAllEncodings retrieves all face encodings from database
func (r *FaceRepository) GetAllEncodings() ([]model.FaceEncoding, error) {
	query := `
		SELECT id, user_id, encoding, quality_score, created_at
		FROM face_encodings
		ORDER BY user_id, created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all encodings: %w", err)
	}
	defer rows.Close()

	var encodings []model.FaceEncoding
	for rows.Next() {
		var enc model.FaceEncoding
		err := rows.Scan(&enc.ID, &enc.UserID, &enc.Encoding, &enc.QualityScore, &enc.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan encoding: %w", err)
		}
		encodings = append(encodings, enc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return encodings, nil
}

// DeleteUserEncodings deletes all face encodings for a user
func (r *FaceRepository) DeleteUserEncodings(userID int64) error {
	query := `DELETE FROM face_encodings WHERE user_id = ?`

	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete encodings: %w", err)
	}

	// Note: It's OK if no rows are affected (user has no encodings yet)
	// This is not an error condition
	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	return nil
}

// CountUserEncodings counts the number of encodings for a user
func (r *FaceRepository) CountUserEncodings(userID int64) (int, error) {
	query := `SELECT COUNT(*) FROM face_encodings WHERE user_id = ?`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count encodings: %w", err)
	}

	return count, nil
}

// LogAttempt logs a face recognition attempt
func (r *FaceRepository) LogAttempt(attempt *model.FaceAttempt) error {
	query := `
		INSERT INTO face_attempts (user_id, matched, confidence, liveness_passed, image_hash, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := r.db.Exec(query,
		attempt.UserID,
		attempt.Matched,
		attempt.Confidence,
		attempt.LivenessPassed,
		attempt.ImageHash,
		attempt.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to log attempt: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	attempt.ID = id
	return nil
}

// GetRecentAttempts retrieves recent face recognition attempts for a user
func (r *FaceRepository) GetRecentAttempts(userID int64, limit int) ([]model.FaceAttempt, error) {
	query := `
		SELECT id, user_id, matched, confidence, liveness_passed, image_hash, ip_address, created_at
		FROM face_attempts
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query attempts: %w", err)
	}
	defer rows.Close()

	var attempts []model.FaceAttempt
	for rows.Next() {
		var att model.FaceAttempt
		err := rows.Scan(
			&att.ID,
			&att.UserID,
			&att.Matched,
			&att.Confidence,
			&att.LivenessPassed,
			&att.ImageHash,
			&att.IPAddress,
			&att.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}
		attempts = append(attempts, att)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return attempts, nil
}

// GetAllAttempts retrieves all face recognition attempts (admin)
func (r *FaceRepository) GetAllAttempts(limit int) ([]model.FaceAttempt, error) {
	query := `
		SELECT id, user_id, matched, confidence, liveness_passed, image_hash, ip_address, created_at
		FROM face_attempts
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query all attempts: %w", err)
	}
	defer rows.Close()

	var attempts []model.FaceAttempt
	for rows.Next() {
		var att model.FaceAttempt
		err := rows.Scan(
			&att.ID,
			&att.UserID,
			&att.Matched,
			&att.Confidence,
			&att.LivenessPassed,
			&att.ImageHash,
			&att.IPAddress,
			&att.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}
		attempts = append(attempts, att)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return attempts, nil
}

// CheckImageHashExists checks if an image hash already exists (prevent replay attacks)
func (r *FaceRepository) CheckImageHashExists(imageHash string, withinMinutes int) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM face_attempts 
		WHERE image_hash = ? 
		AND created_at > datetime('now', '-' || ? || ' minutes')
	`

	var count int
	err := r.db.QueryRow(query, imageHash, withinMinutes).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check image hash: %w", err)
	}

	return count > 0, nil
}

// GetEncodingStats retrieves face encoding statistics for all users (admin)
func (r *FaceRepository) GetEncodingStats() ([]model.FaceEncodingInfo, error) {
	query := `
		SELECT 
			u.id,
			u.username,
			u.full_name,
			COUNT(fe.id) as encoding_count,
			MAX(fe.created_at) as last_enrolled
		FROM users u
		LEFT JOIN face_encodings fe ON u.id = fe.user_id
		WHERE u.is_active = 1
		GROUP BY u.id, u.username, u.full_name
		HAVING encoding_count > 0
		ORDER BY u.full_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query encoding stats: %w", err)
	}
	defer rows.Close()

	var stats []model.FaceEncodingInfo
	for rows.Next() {
		var info model.FaceEncodingInfo
		var lastEnrolledStr string
		err := rows.Scan(
			&info.UserID,
			&info.Username,
			&info.FullName,
			&info.EncodingCount,
			&lastEnrolledStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan encoding info: %w", err)
		}
		
		// Parse time string to time.Time
		if lastEnrolledStr != "" {
			info.LastEnrolled, _ = time.Parse("2006-01-02 15:04:05", lastEnrolledStr)
		}
		
		stats = append(stats, info)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return stats, nil
}
