package repository

import (
	"database/sql"
	"fmt"
	"time"

	"absensi-app/internal/model"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// GetStatistics returns dashboard statistics
func (r *AdminRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users (excluding deleted)
	var totalUsers int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'employee'").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["total_users"] = totalUsers

	// Today's attendance count
	today := time.Now().Format("2006-01-02")
	var hadirHariIni int
	err = r.db.QueryRow("SELECT COUNT(*) FROM absensi WHERE tanggal = ? AND jam_masuk IS NOT NULL", today).Scan(&hadirHariIni)
	if err != nil {
		return nil, fmt.Errorf("failed to count today attendance: %w", err)
	}
	stats["hadir_hari_ini"] = hadirHariIni

	// Not yet clocked in today
	belumAbsen := totalUsers - hadirHariIni
	if belumAbsen < 0 {
		belumAbsen = 0
	}
	stats["belum_absen"] = belumAbsen

	// Completed attendance today (both clock in and out)
	var selesai int
	err = r.db.QueryRow("SELECT COUNT(*) FROM absensi WHERE tanggal = ? AND jam_masuk IS NOT NULL AND jam_pulang IS NOT NULL", today).Scan(&selesai)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed attendance: %w", err)
	}
	stats["selesai"] = selesai

	return stats, nil
}

// GetAllAbsensi returns all attendance records with user info
func (r *AdminRepository) GetAllAbsensi(limit, offset int, startDate, endDate string) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			a.id, a.user_id, a.tanggal, a.jam_masuk, a.jam_pulang, 
			a.keterangan, a.status, a.created_at,
			u.username, u.full_name
		FROM absensi a
		JOIN users u ON a.user_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}

	if startDate != "" {
		query += " AND a.tanggal >= ?"
		args = append(args, startDate)
	}
	if endDate != "" {
		query += " AND a.tanggal <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY a.tanggal DESC, a.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query absensi: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			id         int64
			userID     int64
			tanggal    string
			jamMasuk   sql.NullString
			jamPulang  sql.NullString
			keterangan sql.NullString
			status     string
			createdAt  time.Time
			username   string
			fullName   string
		)

		err := rows.Scan(&id, &userID, &tanggal, &jamMasuk, &jamPulang, &keterangan, &status, &createdAt, &username, &fullName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result := map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"username":   username,
			"full_name":  fullName,
			"tanggal":    tanggal,
			"jam_masuk":  nil,
			"jam_pulang": nil,
			"keterangan": nil,
			"status":     status,
			"created_at": createdAt,
		}

		if jamMasuk.Valid {
			result["jam_masuk"] = jamMasuk.String
		}
		if jamPulang.Valid {
			result["jam_pulang"] = jamPulang.String
		}
		if keterangan.Valid {
			result["keterangan"] = keterangan.String
		}

		results = append(results, result)
	}

	return results, nil
}

// GetTodayAbsensi returns today's attendance for all users
func (r *AdminRepository) GetTodayAbsensi() ([]map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	return r.GetAllAbsensi(100, 0, today, today)
}

// GetUserAbsensi returns attendance records for specific user
func (r *AdminRepository) GetUserAbsensi(userID int64, limit, offset int) ([]model.Absensi, error) {
	query := `
		SELECT id, user_id, tanggal, jam_masuk, jam_pulang, keterangan, status, created_at, updated_at
		FROM absensi
		WHERE user_id = ?
		ORDER BY tanggal DESC, created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user absensi: %w", err)
	}
	defer rows.Close()

	var results []model.Absensi
	for rows.Next() {
		var a model.Absensi
		err := rows.Scan(&a.ID, &a.UserID, &a.Tanggal, &a.JamMasuk, &a.JamPulang, &a.Keterangan, &a.Status, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, a)
	}

	return results, nil
}
