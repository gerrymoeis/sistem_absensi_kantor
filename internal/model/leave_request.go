package model

import "time"

// LeaveRequest represents a leave request from user
type LeaveRequest struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	LeaveType   string    `json:"leave_type"` // izin, sakit, cuti
	LeaveDate   string    `json:"leave_date"` // Format: YYYY-MM-DD
	Reason      string    `json:"reason"`
	ProofFile   *string   `json:"proof_file,omitempty"` // Optional file path/URL
	Status      string    `json:"status"`               // pending, approved, rejected
	ReviewedBy  *int64    `json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes *string   `json:"review_notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LeaveRequestWithUser includes user information
type LeaveRequestWithUser struct {
	LeaveRequest
	FullName     string  `json:"full_name"`
	Username     string  `json:"username"`
	ReviewerName *string `json:"reviewer_name,omitempty"`
}

// CreateLeaveRequest is the request body for creating leave request
type CreateLeaveRequest struct {
	LeaveType string `json:"leave_type" binding:"required,oneof=izin sakit cuti"`
	LeaveDate string `json:"leave_date" binding:"required"` // Format: YYYY-MM-DD
	Reason    string `json:"reason" binding:"required,min=10"`
	ProofFile string `json:"proof_file" binding:"required"` // Base64 encoded file (REQUIRED)
}

// ReviewLeaveRequest is the request body for reviewing leave request
type ReviewLeaveRequest struct {
	Status      string `json:"status" binding:"required,oneof=approved rejected"`
	ReviewNotes string `json:"review_notes,omitempty"`
}
