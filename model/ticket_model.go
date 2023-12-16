package model

import "time"

type Ticket struct {
	ID                       uint       `gorm:"primaryKey" json:"id"`
	WisataID                 uint       `json:"wisata_id"`
	UserID                   uint       `json:"user_id"`
	KodeVoucher              string     `json:"kode_voucher"`
	UsedPoints               int        `json:"used_points"`
	UseAllPoints             bool       `json:"use_all_points"`
	PointsEarned             int        `json:"points_earned"`
	TotalCost                int        `json:"total_cost"`
	InvoiceNumber            string     `json:"invoice_number"`
	Quantity                 int        `json:"quantity"`
	CheckinBooking           *time.Time `json:"checkin_booking"` // Tanggal check-in
	CreatedAt                *time.Time `json:"created_at"`
	UpdatedAt                time.Time
	CarbonFootprint          float64    `json:"carbon_footprint"`
	PaidStatus               bool       `gorm:"default:false" json:"paid_status"` // Tambahkan kolom ini dengan default false
	StatusOrder              string     `gorm:"default:pending" json:"status_order"`
	TenggatPembayaran        *time.Time `json:"tenggat_pembayaran"`
	TotalPotonganKodeVoucher int        `json:"total_potongan_kode_voucher"`
	TotalPotonganPoints      int        `json:"total_potongan_points"`
	HargaSebelumDiskon       int        `json:"harga_sebelum_diskon"`
	UsedPointsOnPurchase     int        `json:"used_points_on_purchase"`
}
