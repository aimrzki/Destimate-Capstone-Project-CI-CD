package model

import (
	"time"
)

type Notification struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `json:"user_id"`
	Message       string     `json:"message"`
	Status        string     `json:"status"`
	Title         string     `json:"title"`
	InvoiceNumber string     `json:"invoice_number"`
	CreatedAt     *time.Time `json:"created_at"`
	IsRead        bool       `json:"is_read"`
	PromoID       uint       `json:"promo_id"` // Add this line
}
