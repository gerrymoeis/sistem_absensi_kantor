package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// columnExists checks if a column exists in a table
func columnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return false, err
		}

		if strings.EqualFold(name, columnName) {
			return true, nil
		}
	}

	return false, rows.Err()
}

// RunMigrations runs database migrations
func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			role VARCHAR(20) DEFAULT 'employee',
			is_active BOOLEAN DEFAULT 1,
			failed_login_attempts INTEGER DEFAULT 0,
			locked_until DATETIME NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS absensi (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			tanggal DATE NOT NULL,
			jam_masuk TIME,
			jam_pulang TIME,
			status VARCHAR(20) DEFAULT 'hadir',
			keterangan TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, tanggal)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_absensi_user_tanggal ON absensi(user_id, tanggal)`,
		`CREATE INDEX IF NOT EXISTS idx_absensi_tanggal ON absensi(tanggal)`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			action_type VARCHAR(50) NOT NULL,
			description TEXT,
			ip_address VARCHAR(45),
			user_agent TEXT,
			status VARCHAR(20) DEFAULT 'success',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action_type)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created ON activity_logs(created_at)`,
		// Face Recognition Tables
		`CREATE TABLE IF NOT EXISTS face_encodings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			encoding BLOB NOT NULL,
			quality_score FLOAT DEFAULT 0.0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_face_encodings_user ON face_encodings(user_id)`,
		`CREATE TABLE IF NOT EXISTS face_attempts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			matched BOOLEAN DEFAULT 0,
			confidence FLOAT DEFAULT 0.0,
			liveness_passed BOOLEAN DEFAULT 0,
			image_hash VARCHAR(64),
			ip_address VARCHAR(45),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_face_attempts_user ON face_attempts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_face_attempts_created ON face_attempts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_face_attempts_image_hash ON face_attempts(image_hash)`,
		// Create enrollment_sessions table
		`CREATE TABLE IF NOT EXISTS enrollment_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'in_progress',
			frontal_count INTEGER DEFAULT 0,
			left_profile_count INTEGER DEFAULT 0,
			right_profile_count INTEGER DEFAULT 0,
			liveness_passed BOOLEAN DEFAULT 0,
			total_encodings INTEGER DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_enrollment_sessions_user ON enrollment_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_enrollment_sessions_status ON enrollment_sessions(status)`,
		// Create liveness_attempts table
		`CREATE TABLE IF NOT EXISTS liveness_attempts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			attempt_type VARCHAR(20),
			liveness_type VARCHAR(20),
			passed BOOLEAN DEFAULT 0,
			confidence FLOAT DEFAULT 0.0,
			challenge VARCHAR(50),
			image_hash VARCHAR(64),
			ip_address VARCHAR(45),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_liveness_attempts_user ON liveness_attempts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_liveness_attempts_type ON liveness_attempts(attempt_type)`,
		`CREATE INDEX IF NOT EXISTS idx_liveness_attempts_created ON liveness_attempts(created_at)`,
		// Leave Requests Table
		`CREATE TABLE IF NOT EXISTS leave_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			leave_type VARCHAR(20) NOT NULL,
			leave_date DATE NOT NULL,
			reason TEXT NOT NULL,
			proof_file TEXT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			reviewed_by INTEGER NULL,
			reviewed_at DATETIME NULL,
			review_notes TEXT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_requests_user ON leave_requests(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_requests_status ON leave_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_leave_requests_date ON leave_requests(leave_date)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	// Conditional ALTER TABLE migrations (check if column exists first)
	conditionalMigrations := []struct {
		table  string
		column string
		sql    string
	}{
		// Users table columns
		{"users", "face_enrolled", "ALTER TABLE users ADD COLUMN face_enrolled BOOLEAN DEFAULT 0"},
		{"users", "face_enrolled_at", "ALTER TABLE users ADD COLUMN face_enrolled_at DATETIME NULL"},
		{"users", "enrollment_type", "ALTER TABLE users ADD COLUMN enrollment_type VARCHAR(20) DEFAULT 'comprehensive'"},
		// Absensi table columns
		{"absensi", "photo_masuk", "ALTER TABLE absensi ADD COLUMN photo_masuk TEXT NULL"},
		{"absensi", "photo_pulang", "ALTER TABLE absensi ADD COLUMN photo_pulang TEXT NULL"},
		{"absensi", "face_verified", "ALTER TABLE absensi ADD COLUMN face_verified BOOLEAN DEFAULT 0"},
		{"absensi", "liveness_score", "ALTER TABLE absensi ADD COLUMN liveness_score FLOAT DEFAULT 0.0"},
		{"absensi", "face_confidence", "ALTER TABLE absensi ADD COLUMN face_confidence FLOAT DEFAULT 0.0"},
		// Face encodings table column
		{"face_encodings", "angle", "ALTER TABLE face_encodings ADD COLUMN angle VARCHAR(20) DEFAULT 'frontal'"},
	}

	for _, migration := range conditionalMigrations {
		exists, err := columnExists(db, migration.table, migration.column)
		if err != nil {
			return fmt.Errorf("failed to check if column %s.%s exists: %w", migration.table, migration.column, err)
		}

		if !exists {
			if _, err := db.Exec(migration.sql); err != nil {
				return fmt.Errorf("failed to add column %s.%s: %w", migration.table, migration.column, err)
			}
		}
	}

	return nil
}
