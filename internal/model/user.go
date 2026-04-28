package model

import "time"

type User struct {
	ID                  int64      `json:"id"`
	Username            string     `json:"username"`
	PasswordHash        string     `json:"-"` // Never expose password hash in JSON
	FullName            string     `json:"full_name"`
	Role                string     `json:"role"`
	IsActive            bool       `json:"is_active"`
	FailedLoginAttempts int        `json:"-"` // Don't expose in API
	LockedUntil         *time.Time `json:"-"` // Don't expose in API
	FaceEnrolled        bool       `json:"face_enrolled"`
	FaceEnrolledAt      *time.Time `json:"face_enrolled_at,omitempty"`
	EnrollmentType      string     `json:"enrollment_type,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
