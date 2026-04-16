package model

import "time"

type Absensi struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Tanggal    string    `json:"tanggal"` // Format: YYYY-MM-DD
	JamMasuk   *string   `json:"jam_masuk,omitempty"`
	JamPulang  *string   `json:"jam_pulang,omitempty"`
	Status     string    `json:"status"`
	Keterangan *string   `json:"keterangan,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AbsensiWithUser struct {
	Absensi
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

type ClockInRequest struct {
	Keterangan string `json:"keterangan,omitempty"`
}

type ClockOutRequest struct {
	Keterangan string `json:"keterangan,omitempty"`
}
