package database

import (
	"database/sql"
	"fmt"
)

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
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}
