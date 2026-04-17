package repository

import (
	"database/sql"
	"fmt"
	"time"

	"absensi-app/internal/model"
)

type AbsensiRepository struct {
	db *sql.DB
}

func NewAbsensiRepository(db *sql.DB) *AbsensiRepository {
	return &AbsensiRepository{db: db}
}

// FindByUserAndDate finds absensi record by user ID and date
func (r *AbsensiRepository) FindByUserAndDate(userID int64, tanggal string) (*model.Absensi, error) {
	absensi := &model.Absensi{}
	err := r.db.QueryRow(`
		SELECT id, user_id, tanggal, jam_masuk, jam_pulang, status, keterangan, created_at, updated_at
		FROM absensi
		WHERE user_id = ? AND tanggal = ?
	`, userID, tanggal).Scan(
		&absensi.ID,
		&absensi.UserID,
		&absensi.Tanggal,
		&absensi.JamMasuk,
		&absensi.JamPulang,
		&absensi.Status,
		&absensi.Keterangan,
		&absensi.CreatedAt,
		&absensi.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find absensi: %w", err)
	}

	return absensi, nil
}

// Create creates a new absensi record
func (r *AbsensiRepository) Create(absensi *model.Absensi) error {
	result, err := r.db.Exec(`
		INSERT INTO absensi (user_id, tanggal, jam_masuk, status, keterangan)
		VALUES (?, ?, ?, ?, ?)
	`, absensi.UserID, absensi.Tanggal, absensi.JamMasuk, absensi.Status, absensi.Keterangan)

	if err != nil {
		return fmt.Errorf("failed to create absensi: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	absensi.ID = id
	return nil
}

// Update updates an existing absensi record
func (r *AbsensiRepository) Update(absensi *model.Absensi) error {
	_, err := r.db.Exec(`
		UPDATE absensi
		SET jam_pulang = ?, keterangan = ?, updated_at = ?
		WHERE id = ?
	`, absensi.JamPulang, absensi.Keterangan, time.Now(), absensi.ID)

	if err != nil {
		return fmt.Errorf("failed to update absensi: %w", err)
	}

	return nil
}

// FindByUserID finds all absensi records for a user
func (r *AbsensiRepository) FindByUserID(userID int64, limit, offset int) ([]model.Absensi, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, tanggal, jam_masuk, jam_pulang, status, keterangan, created_at, updated_at
		FROM absensi
		WHERE user_id = ?
		ORDER BY tanggal DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query absensi: %w", err)
	}
	defer rows.Close()

	var results []model.Absensi
	for rows.Next() {
		var absensi model.Absensi
		err := rows.Scan(
			&absensi.ID,
			&absensi.UserID,
			&absensi.Tanggal,
			&absensi.JamMasuk,
			&absensi.JamPulang,
			&absensi.Status,
			&absensi.Keterangan,
			&absensi.CreatedAt,
			&absensi.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan absensi: %w", err)
		}
		results = append(results, absensi)
	}

	return results, nil
}

// FindByUserIDAndDateRange finds absensi records for a user within date range
func (r *AbsensiRepository) FindByUserIDAndDateRange(userID int64, startDate, endDate string) ([]model.Absensi, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, tanggal, jam_masuk, jam_pulang, status, keterangan, created_at, updated_at
		FROM absensi
		WHERE user_id = ? AND tanggal BETWEEN ? AND ?
		ORDER BY tanggal DESC
	`, userID, startDate, endDate)

	if err != nil {
		return nil, fmt.Errorf("failed to query absensi: %w", err)
	}
	defer rows.Close()

	var results []model.Absensi
	for rows.Next() {
		var absensi model.Absensi
		err := rows.Scan(
			&absensi.ID,
			&absensi.UserID,
			&absensi.Tanggal,
			&absensi.JamMasuk,
			&absensi.JamPulang,
			&absensi.Status,
			&absensi.Keterangan,
			&absensi.CreatedAt,
			&absensi.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan absensi: %w", err)
		}
		results = append(results, absensi)
	}

	return results, nil
}
