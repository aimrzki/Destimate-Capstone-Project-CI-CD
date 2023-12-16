package model

import "time"

type TermCondition struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `json:"tnc_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
