package model

import "time"

// FaceEncoding represents a stored face encoding for a user
type FaceEncoding struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Encoding     []byte    `json:"-"` // Never expose raw encoding in JSON
	QualityScore float64   `json:"quality_score"`
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

// FaceEnrollRequest represents a request to enroll a face
type FaceEnrollRequest struct {
	UserID    int64  `json:"user_id" binding:"required"`
	ImageData string `json:"image_data" binding:"required"` // Base64 encoded image
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
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name"`
	EncodingCount int      `json:"encoding_count"`
	LastEnrolled time.Time `json:"last_enrolled"`
}
