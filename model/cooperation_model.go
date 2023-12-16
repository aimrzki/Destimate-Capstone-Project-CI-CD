package model

import "time"

type CooperationMessage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"` // Tambahkan field ini
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}
