package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"time"
)

func GetTicketsByUser(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		// Mendapatkan ID pengguna dari token
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengambil tiket yang telah dibeli oleh pengguna berdasarkan UserID
		var tickets []model.Ticket
		result = db.Where("user_id = ?", user.ID).Find(&tickets)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user's tickets"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Membuat respons dengan data tiket yang telah dibeli
		var ticketDetails []map[string]interface{}
		for _, ticket := range tickets {
			// Menangani perubahan paid_status yang dilakukan oleh admin
			if ticket.PaidStatus {
				// Jika paid_status sudah true, perbarui tenggat_pembayaran dan status_order
				if ticket.TenggatPembayaran != nil && time.Now().After(*ticket.TenggatPembayaran) {
					// Jika tenggat_pembayaran sudah lewat, ubah status_order menjadi "dibatalkan" dan tenggat_pembayaran menjadi nil
					ticket.StatusOrder = "dibatalkan"
					ticket.TenggatPembayaran = nil
				} else {
					// Jika tenggat_pembayaran masih berlaku, ubah status_order menjadi "success" dan tenggat_pembayaran menjadi nil
					ticket.StatusOrder = "success"
					ticket.TenggatPembayaran = nil
				}

				// Simpan perubahan ke dalam database
				if err := db.Save(&ticket).Error; err != nil {
					errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update ticket status"}
					return c.JSON(http.StatusInternalServerError, errorResponse)
				}
			}

			// Mengambil detail event berdasarkan EventID yang ada pada tiket
			var wisata model.Wisata
			eventResult := db.First(&wisata, ticket.WisataID)
			if eventResult.Error != nil {
				// Handle jika event tidak ditemukan
				continue
			}

			// Menambahkan informasi kode voucher yang digunakan
			var kodeVoucher string
			if ticket.KodeVoucher != "" {
				kodeVoucher = ticket.KodeVoucher
			}

			// Menambahkan tenggat pembayaran dan status order ke detail tiket
			ticketDetail := map[string]interface{}{
				"ticket_id":                   ticket.ID,
				"user_id":                     ticket.UserID,
				"username":                    user.Username,
				"email":                       user.Email,
				"wisata_id":                   ticket.WisataID,
				"wisata_name":                 wisata.Title,
				"lokasi_wisata":               wisata.Location,
				"kota_wisata":                 wisata.Kota,
				"maps_link":                   wisata.MapsLink,
				"total_cost":                  ticket.TotalCost,
				"invoice_number":              ticket.InvoiceNumber,
				"kode_voucher":                kodeVoucher,
				"quantity":                    ticket.Quantity,
				"paid":                        ticket.PaidStatus,
				"carboon_footprint":           ticket.CarbonFootprint,
				"points_earned":               ticket.PointsEarned,
				"checkin_booking":             ticket.CheckinBooking,
				"photo_wisata1":               wisata.PhotoWisata1,
				"photo_wisata2":               wisata.PhotoWisata2,
				"photo_wisata3":               wisata.PhotoWisata3,
				"tenggat_pembayaran":          ticket.TenggatPembayaran,
				"status_order":                ticket.StatusOrder,
				"harga_sebelum_diskon":        ticket.HargaSebelumDiskon,
				"total_potongan_kode_voucher": ticket.TotalPotonganKodeVoucher,
				"total_potongan_points":       ticket.TotalPotonganPoints,
				"lat":                         wisata.Lat,
				"long":                        wisata.Long,
			}

			// Menambahkan objek tiket ke daftar ticketDetails
			ticketDetails = append(ticketDetails, ticketDetail)
		}

		// Mengembalikan respons dengan detail tiket yang telah dibeli
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "User's tickets retrieved successfully",
			"ticket_data": ticketDetails,
		})
	}
}

func GetTransactionHistoryByInvoiceNumber(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		// Mendapatkan ID pengguna dari token
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mendapatkan invoice_number dari parameter URL
		invoiceNumber := c.Param("invoice_number")

		// Mengambil tiket yang memiliki invoice_number yang sesuai
		var tickets []model.Ticket
		result = db.Where("user_id = ? AND invoice_number = ?", user.ID, invoiceNumber).Find(&tickets)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch transaction history"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Membuat respons dengan data transaksi
		var transactionDetails []map[string]interface{}
		for _, ticket := range tickets {
			// Mendapatkan detail event berdasarkan EventID yang ada pada tiket
			var wisata model.Wisata
			eventResult := db.First(&wisata, ticket.WisataID)
			if eventResult.Error != nil {
				// Handle jika event tidak ditemukan
				continue
			}

			// Menambahkan informasi kode voucher yang digunakan
			var kodeVoucher string
			if ticket.KodeVoucher != "" {
				kodeVoucher = ticket.KodeVoucher
			}

			// Menambahkan objek tiket ke daftar transactionDetails
			transactionDetail := map[string]interface{}{
				"ticket_id":                   ticket.ID,
				"user_id":                     ticket.UserID,
				"username":                    user.Username,
				"email":                       user.Email,
				"wisata_id":                   ticket.WisataID,
				"wisata_name":                 wisata.Title,
				"lokasi_wisata":               wisata.Location,
				"kota_wisata":                 wisata.Kota,
				"maps_link":                   wisata.MapsLink,
				"total_cost":                  ticket.TotalCost,
				"invoice_number":              ticket.InvoiceNumber,
				"kode_voucher":                kodeVoucher,
				"quantity":                    ticket.Quantity,
				"paid":                        ticket.PaidStatus,
				"carboon_footprint":           ticket.CarbonFootprint,
				"points_earned":               ticket.PointsEarned,
				"checkin_booking":             ticket.CheckinBooking,
				"photo_wisata1":               wisata.PhotoWisata1,
				"photo_wisata2":               wisata.PhotoWisata2,
				"photo_wisata3":               wisata.PhotoWisata3,
				"tenggat_pembayaran":          ticket.TenggatPembayaran,
				"status_order":                ticket.StatusOrder,
				"harga_sebelum_diskon":        ticket.HargaSebelumDiskon,
				"total_potongan_kode_voucher": ticket.TotalPotonganKodeVoucher,
				"total_potongan_points":       ticket.TotalPotonganPoints,
				"lat":                         wisata.Lat,
				"long":                        wisata.Long,
			}

			// Menambahkan objek transaksi ke daftar transactionDetails
			transactionDetails = append(transactionDetails, transactionDetail)
		}

		// Mengembalikan respons dengan detail transaksi berdasarkan invoice_number
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Transaction history retrieved successfully",
			"ticket_data": transactionDetails,
		})
	}
}

func GetPointsHistory(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var pointsHistory []map[string]interface{}
		var tickets []model.Ticket
		result = db.Where("user_id = ? AND (points_earned > 0 OR used_points > 0) AND paid_status = ?", user.ID, true).Find(&tickets)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user's points history"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		for _, ticket := range tickets {
			if ticket.PaidStatus {
				if ticket.TenggatPembayaran != nil && time.Now().After(*ticket.TenggatPembayaran) {
					ticket.StatusOrder = "dibatalkan"
					ticket.TenggatPembayaran = nil
				} else {
					ticket.StatusOrder = "success"
					ticket.TenggatPembayaran = nil
				}

				if err := db.Save(&ticket).Error; err != nil {
					errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update ticket status"}
					return c.JSON(http.StatusInternalServerError, errorResponse)
				}
			}

			var wisata model.Wisata
			eventResult := db.First(&wisata, ticket.WisataID)
			if eventResult.Error != nil {
				continue
			}

			if ticket.PointsEarned > 0 {
				pointsDetailEarned := map[string]interface{}{
					"wisata_name":   wisata.Title,
					"points_earned": ticket.PointsEarned,
					"message":       "Poin bertambah",
				}
				pointsHistory = append(pointsHistory, pointsDetailEarned)
			}

			if ticket.UsedPoints > 0 {
				pointsDetailUsed := map[string]interface{}{
					"wisata_name": wisata.Title,
					"points_used": ticket.UsedPoints,
					"message":     "Poin berkurang",
				}
				pointsHistory = append(pointsHistory, pointsDetailUsed)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":           http.StatusOK,
			"error":          false,
			"message":        "User's points history retrieved successfully",
			"points_history": pointsHistory,
		})
	}
}
