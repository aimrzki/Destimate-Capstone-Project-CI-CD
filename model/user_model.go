package model

import "time"

// nambahin foto profil

type User struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Name              string     `json:"name"`
	Username          string     `gorm:"uniqueIndex;size:255" json:"username"`
	Email             string     `gorm:"uniqueIndex;size:255" json:"email"`
	Password          string     `json:"password"`
	PhotoProfil       string     `json:"photo_profil"`
	PhoneNumber       string     `gorm:"uniqueIndex;size:255" json:"phone_number"`
	Points            int        `json:"points"`
	IsAdmin           bool       `gorm:"default:false" json:"isAdmin"`
	IsVerified        bool       `gorm:"default:false" json:"is_verified"` // Tambahkan kolom ini
	VerificationToken string     `json:"verification_token"`               // Tambahkan kolom ini
	Lat               float64    `json:"lat,omitempty"`                    // Menambahkan Lat (Latitude)
	Long              float64    `json:"long,omitempty"`                   // Menambahkan Long (Longitude)
	CreatedAt         *time.Time `json:"created_at"`
	Wisata            []Wisata   `gorm:"foreignKey:UserID" json:"wisata"`
	CategoryID        uint       `json:"category_id"`
	CategoryKesukaan  string     `json:"category_kesukaan"`
	ConfirmPassword   string     `json:"confirm_password"`
	StatusCategory    bool       `gorm:"default:false" json:"status_category"`
}

// Buat struct untuk permintaan perubahan kata sandi
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
	ConfirmPassword string `json:"confirmPassword" binding:"required"` // Tambahkan ConfirmPassword
}
