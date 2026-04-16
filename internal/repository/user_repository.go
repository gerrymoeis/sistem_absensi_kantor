package repository

import (
	"database/sql"
	"fmt"

	"absensi-app/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByUsername finds user by username
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(`
		SELECT id, username, password_hash, full_name, role, is_active, 
		       failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE username = ?
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// FindByID finds user by ID
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(`
		SELECT id, username, password_hash, full_name, role, is_active,
		       failed_login_attempts, locked_until, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.FailedLoginAttempts,
		&user.LockedUntil,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	result, err := r.db.Exec(`
		INSERT INTO users (username, password_hash, full_name, role, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, user.Username, user.PasswordHash, user.FullName, user.Role, user.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = id
	return nil
}

// FindAll returns all users
func (r *UserRepository) FindAll() ([]model.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, password_hash, full_name, role, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.FullName,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates user information
func (r *UserRepository) Update(user *model.User) error {
	result, err := r.db.Exec(`
		UPDATE users 
		SET full_name = ?, role = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, user.FullName, user.Role, user.IsActive, user.ID)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(userID int64, passwordHash string) error {
	result, err := r.db.Exec(`
		UPDATE users 
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, passwordHash, userID)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete soft deletes a user (sets is_active to false)
func (r *UserRepository) Delete(userID int64) error {
	result, err := r.db.Exec(`
		UPDATE users 
		SET is_active = 0, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CheckUsernameExists checks if username already exists
func (r *UserRepository) CheckUsernameExists(username string, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE username = ? AND id != ?
	`, username, excludeID).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check username: %w", err)
	}

	return count > 0, nil
}

// IncrementFailedLoginAttempts increments failed login attempts counter
func (r *UserRepository) IncrementFailedLoginAttempts(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to increment failed login attempts: %w", err)
	}

	return nil
}

// ResetFailedLoginAttempts resets failed login attempts counter
func (r *UserRepository) ResetFailedLoginAttempts(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE users 
		SET failed_login_attempts = 0,
		    locked_until = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, userID)

	if err != nil {
		return fmt.Errorf("failed to reset failed login attempts: %w", err)
	}

	return nil
}

// LockAccount locks user account until specified time
func (r *UserRepository) LockAccount(userID int64, lockUntil string) error {
	_, err := r.db.Exec(`
		UPDATE users 
		SET locked_until = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, lockUntil, userID)

	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	return nil
}
