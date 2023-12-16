package helper

import (
	"myproject/model"
	"time"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//Nambahin foto profil

type UserResponse struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	PhoneNumber      string     `json:"phone_number"`
	PhotoProfil      string     `json:"photo_profil"`
	Points           int        `json:"points"`
	IsVerified       bool       `json:"is_verified"`
	CategoryKesukaan string     `json:"category_kesukaan"`
	CategoryID       uint       `json:"category_id"`
	CreatedAt        *time.Time `json:"created_at"`
}

//Nambahin foto profil

type EditUserResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Username         string `json:"username"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phone_number"`
	PhotoProfil      string `json:"photo_profil"`
	CategoryKesukaan string `json:"category_kesukaan"`
	CategoryID       uint   `json:"category_id"`
}

type EditUserLocation struct {
	ID   uint    `json:"id"`
	Lat  float64 `json:"lat,omitempty"`  // Menambahkan Lat (Latitude)
	Long float64 `json:"long,omitempty"` // Menambahkan Long (Longitude)
}

type CarbonFootprintResponse struct {
	CarbonFootprint int `json:"carbon_footprint"`
	EquivalentHours int `json:"equivalent_hours"`
}

// Response struct untuk menyimpan informasi respons
type Response struct {
	Code     int             `json:"code"`
	Error    bool            `json:"error"`
	Message  string          `json:"message,omitempty"`
	Category *model.Category `json:"category,omitempty"`
}
