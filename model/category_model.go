package model

import "time"

type Category struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CategoryName string    `gorm:"unique;not null;size:255" json:"category_name"`
	CreatedAt    time.Time `json:"created_at"`
}
