package model

import "time"

type Promo struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	Title                string    `json:"title"`
	NamaPromo            string    `json:"nama_promo"`
	KodeVoucher          string    `gorm:"uniqueIndex;size:255" json:"kode_voucher"`
	JumlahPotonganPersen int       `json:"jumlah_potongan_persen"`
	StatusAktif          bool      `json:"status_aktif"`
	TanggalKadaluarsa    time.Time `json:"tanggal_kadaluarsa"`
	ImageVoucher         string    `json:"image_voucher"`
	Deskripsi            string    `json:"deskripsi"`
	Peraturan            string    `json:"peraturan"`
	CreatedAt            time.Time `json:"created_at"`
}
