package model

import "time"

type ActivityLog struct {
	ID          int64     `json:"id"`
	UserID      *int64    `json:"user_id,omitempty"` // Nullable untuk failed login
	ActionType  string    `json:"action_type"`
	Description string    `json:"description"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Status      string    `json:"status"` // success, failed
	CreatedAt   time.Time `json:"created_at"`
}

// Action types constants
const (
	ActionLogin       = "login"
	ActionLogout      = "logout"
	ActionClockIn     = "clock_in"
	ActionClockOut    = "clock_out"
	ActionViewData    = "view_data"
	ActionExport      = "export"
	ActionUpdate      = "update"
	ActionAdminCreate = "admin_create"
	ActionAdminUpdate = "admin_update"
	ActionAdminDelete = "admin_delete"
)

// Status constants
const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)
