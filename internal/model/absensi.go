package model

import "time"

type Absensi struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Tanggal        string    `json:"tanggal"` // Format: YYYY-MM-DD
	JamMasuk       *string   `json:"jam_masuk,omitempty"`
	JamPulang      *string   `json:"jam_pulang,omitempty"`
	Status         string    `json:"status"`
	Keterangan     *string   `json:"keterangan,omitempty"`
	PhotoMasuk     *string   `json:"photo_masuk,omitempty"`
	PhotoPulang    *string   `json:"photo_pulang,omitempty"`
	FaceVerified   bool      `json:"face_verified"`
	LivenessScore  float64   `json:"liveness_score,omitempty"`
	FaceConfidence float64   `json:"face_confidence,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AbsensiWithUser struct {
	Absensi
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

type ClockInRequest struct {
	Keterangan string `json:"keterangan,omitempty"`
	PhotoData  string `json:"photo_data,omitempty"` // Base64 encoded photo
}

type ClockOutRequest struct {
	Keterangan string `json:"keterangan,omitempty"`
	PhotoData  string `json:"photo_data,omitempty"` // Base64 encoded photo
}
