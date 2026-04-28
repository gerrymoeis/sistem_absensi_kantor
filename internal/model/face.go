package model

import "time"

// FaceEncoding represents a stored face encoding for a user
type FaceEncoding struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Encoding     []byte    `json:"-"` // Never expose raw encoding in JSON
	QualityScore float64   `json:"quality_score"`
	Angle        string    `json:"angle"` // 'frontal', 'left_profile', 'right_profile'
	CreatedAt    time.Time `json:"created_at"`
}

// FaceAttempt represents a face recognition attempt (audit log)
type FaceAttempt struct {
	ID             int64     `json:"id"`
	UserID         *int64    `json:"user_id,omitempty"` // Nullable for failed matches
	Matched        bool      `json:"matched"`
	Confidence     float64   `json:"confidence"`
	LivenessPassed bool      `json:"liveness_passed"`
	ImageHash      string    `json:"image_hash"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

// FaceEnrollRequest represents a request to enroll a face (admin)
type FaceEnrollRequest struct {
	UserID    int64  `json:"user_id" binding:"required"`
	ImageData string `json:"image_data" binding:"required"` // Base64 encoded image
}

// SelfEnrollRequest represents a request for self face enrollment (employee)
type SelfEnrollRequest struct {
	ImageData string `json:"image_data" binding:"required"` // Base64 encoded image
}

// EnrollmentPhoto represents a single photo in comprehensive enrollment
type EnrollmentPhoto struct {
	Step      string `json:"step" binding:"required"`      // 'frontal', 'left', 'right', 'up', 'down'
	Data      string `json:"data" binding:"required"`      // Base64 encoded image
	Timestamp int64  `json:"timestamp" binding:"required"` // Capture timestamp
}

// ComprehensiveEnrollRequest represents a comprehensive face enrollment with multiple angles
type ComprehensiveEnrollRequest struct {
	Photos    []EnrollmentPhoto `json:"photos" binding:"required,min=5,max=5"` // Exactly 5 photos required
	Timestamp int64             `json:"timestamp" binding:"required"`
	Metadata  struct {
		FPS     int    `json:"fps"`
		Device  string `json:"device"`
		Version string `json:"version"`
	} `json:"metadata"`
}

// FaceRecognitionRequest represents a request to recognize a face
type FaceRecognitionRequest struct {
	ImageData      string `json:"image_data" binding:"required"` // Base64 encoded image
	LivenessData   string `json:"liveness_data,omitempty"`       // Optional liveness check data
	LivenessPassed bool   `json:"liveness_passed"`               // Client-side liveness result
}

// FaceRecognitionResponse represents the result of face recognition
type FaceRecognitionResponse struct {
	Matched    bool    `json:"matched"`
	UserID     *int64  `json:"user_id,omitempty"`
	FullName   string  `json:"full_name,omitempty"`
	Confidence float64 `json:"confidence"`
	Message    string  `json:"message"`
}

// FaceEncodingInfo represents face encoding information for admin
type FaceEncodingInfo struct {
	UserID        int64     `json:"user_id"`
	Username      string    `json:"username"`
	FullName      string    `json:"full_name"`
	EncodingCount int       `json:"encoding_count"`
	LastEnrolled  time.Time `json:"last_enrolled"`
}

// EnrollmentSession represents a face enrollment session
type EnrollmentSession struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	Status           string     `json:"status"` // 'in_progress', 'completed', 'failed'
	FrontalCount     int        `json:"frontal_count"`
	LeftProfileCount int        `json:"left_profile_count"`
	RightProfileCount int       `json:"right_profile_count"`
	LivenessPassed   bool       `json:"liveness_passed"`
	TotalEncodings   int        `json:"total_encodings"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

// LivenessAttempt represents a liveness detection attempt
type LivenessAttempt struct {
	ID           int64     `json:"id"`
	UserID       *int64    `json:"user_id,omitempty"`
	AttemptType  string    `json:"attempt_type"`  // 'enrollment' or 'attendance'
	LivenessType string    `json:"liveness_type"` // 'active' or 'passive'
	Passed       bool      `json:"passed"`
	Confidence   float64   `json:"confidence"`
	Challenge    string    `json:"challenge"` // 'blink', 'smile', 'mouth_open', 'passive'
	ImageHash    string    `json:"image_hash"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
}
