package model

import "time"

// nambahin description is open, is open, dan maps link

type Wisata struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Kode              string     `gorm:"uniqueIndex;size:255" json:"kode"`
	Title             string     `json:"title"`
	Location          string     `json:"location"`
	Kota              string     `json:"kota"`
	Description       string     `json:"description"`
	Price             int        `json:"price"`
	Lat               float64    `json:"lat,omitempty"`  // Tambahkan Lat (Latitude)
	Long              float64    `json:"long,omitempty"` // Tambahkan Long (Longitude)
	UserID            uint       `json:"user_id"`        // ID pengguna yang membuat event
	AvailableTickets  int        `json:"available_tickets"`
	PhotoWisata1      string     `json:"photo_wisata1"`
	PhotoWisata2      string     `json:"photo_wisata2"`
	PhotoWisata3      string     `json:"photo_wisata3"`
	CategoryID        uint       `json:"category_id"`
	Category          Category   `gorm:"foreignKey:CategoryID" json:"category"`
	MapsLink          string     `json:"maps_link"`           // Menambahkan link maps
	IsOpen            bool       `json:"is_open"`             // Menambahkan isopen boolean
	DescriptionIsOpen string     `json:"description_is_open"` // menambahkan deskripsi isopen
	Fasilitas         string     `gorm:"type:json" json:"fasilitas"`
	VideoLink         string     `json:"video_link"`
	CreatedAt         *time.Time `json:"created_at"`
	UpdatedAt         time.Time
}

// Metode untuk mengurangi jumlah tiket yang tersedia
func (e *Wisata) DecrementTickets(quantity int) {
	if e.AvailableTickets >= quantity {
		e.AvailableTickets -= quantity
	}
}
