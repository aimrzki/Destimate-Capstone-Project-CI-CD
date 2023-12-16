package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"math"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"time"
)

func CalculateCarbonFootprint(user model.User, wisata model.Wisata) float64 {
	earthRadius := 6371.0
	userLat := user.Lat
	userLong := user.Long
	wisataLat := wisata.Lat
	wisataLong := wisata.Long

	if wisataLat < 0 {
		wisataLat = -wisataLat
	}
	if wisataLong < 0 {
		wisataLong = -wisataLong
	}
	if userLat < 0 {
		userLat = -userLat
	}
	if userLong < 0 {
		userLong = -userLong
	}

	userLatRad := userLat * (math.Pi / 180.0)
	userLongRad := userLong * (math.Pi / 180.0)
	wisataLatRad := wisataLat * (math.Pi / 180.0)
	wisataLongRad := wisataLong * (math.Pi / 180.0)

	dLat := wisataLatRad - userLatRad
	dLong := wisataLongRad - userLongRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(userLatRad)*math.Cos(wisataLatRad)*math.Sin(dLong/2)*math.Sin(dLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c

	carbonFootprint := distance * 14.8
	return carbonFootprint
}

func BuyTicket(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var ticketPurchase struct {
			WisataID       uint   `json:"wisata_id"`
			KodeVoucher    string `json:"kode_voucher"`
			UseAllPoints   bool   `json:"use_all_points"`
			UsedPoints     int    `json:"used_points"`
			Quantity       int    `json:"quantity"`
			CheckinBooking string `json:"checkin_booking"`
		}

		if err := c.Bind(&ticketPurchase); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var wisata model.Wisata
		wisataResult := db.First(&wisata, ticketPurchase.WisataID)
		if wisataResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if ticketPurchase.Quantity <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Quantity must be greater than 0"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		checkinBookingTime, err := time.Parse("2006-01-02", ticketPurchase.CheckinBooking)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid checkin_booking date format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if checkinBookingTime.Before(time.Now().Truncate(24 * time.Hour)) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Checkin date must be today or later"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var discountPercentage int
		totalCost := wisata.Price * ticketPurchase.Quantity
		pointsEarned := totalCost / 10000
		var totalPotonganKodeVoucher int
		var totalPotonganPoints int

		var hargaSebelumDiskon int
		hargaSebelumDiskon = totalCost

		if ticketPurchase.KodeVoucher != "" {
			var promo model.Promo
			promoResult := db.Where("kode_voucher = ?", ticketPurchase.KodeVoucher).First(&promo)
			if promoResult.Error == nil {
				currentTime := time.Now()
				if promo.StatusAktif && currentTime.Before(promo.TanggalKadaluarsa) {
					discountPercentage = promo.JumlahPotonganPersen
					if discountPercentage > 0 {
						discount := (totalCost * discountPercentage) / 100
						totalCost -= discount
						totalPotonganKodeVoucher += discount
					}
					pointsEarned = 0
				} else {
					if !promo.StatusAktif {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Voucher belum aktif"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					} else if currentTime.After(promo.TanggalKadaluarsa) {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Voucher sudah expired"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					} else {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid kode voucher"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					}
				}
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid kode voucher"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		var usedPoints int

		if ticketPurchase.UseAllPoints {
			// Calculate the maximum points that can be used
			maxPoints := totalCost / 1000
			if maxPoints > user.Points {
				maxPoints = user.Points
			}

			usedPoints = maxPoints
			additionalDiscount := usedPoints * 1000
			totalCost -= additionalDiscount
			totalPotonganPoints += additionalDiscount

			// Deduct the used points from the user's account
			user.Points -= usedPoints

			if err := db.Save(&user).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user points"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		} else {
			usedPoints = ticketPurchase.UsedPoints
			additionalDiscount := usedPoints * 1000
			if additionalDiscount <= totalCost {
				totalCost -= additionalDiscount
				totalPotonganPoints += additionalDiscount
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough points to use"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if wisata.AvailableTickets < ticketPurchase.Quantity {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough available tickets"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		wisata.AvailableTickets -= ticketPurchase.Quantity
		if err := db.Save(&wisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update available tickets"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		carbonFootprint := CalculateCarbonFootprint(user, wisata)

		tenggatPembayaran := checkinBookingTime

		ticket := model.Ticket{
			WisataID:                 wisata.ID,
			UserID:                   user.ID,
			UsedPoints:               usedPoints,
			TotalCost:                totalCost,
			InvoiceNumber:            helper.GenerateInvoiceNumber(),
			KodeVoucher:              ticketPurchase.KodeVoucher,
			Quantity:                 ticketPurchase.Quantity,
			CheckinBooking:           &checkinBookingTime,
			PaidStatus:               false,
			PointsEarned:             pointsEarned,
			CarbonFootprint:          carbonFootprint,
			StatusOrder:              "pending", // Set nilai default
			TenggatPembayaran:        &tenggatPembayaran,
			TotalPotonganKodeVoucher: totalPotonganKodeVoucher,
			TotalPotonganPoints:      totalPotonganPoints,
			HargaSebelumDiskon:       hargaSebelumDiskon,
			UsedPointsOnPurchase:     usedPoints,
			UseAllPoints:             ticketPurchase.UseAllPoints,
		}

		if err := db.Create(&ticket).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create ticket"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		emailSubject := helper.GetEmailSubject(ticket)
		wisataName := wisata.Title
		emailBody := helper.GetEmailBody(ticket, totalCost, wisataName, ticketPurchase.KodeVoucher, pointsEarned, usedPoints, carbonFootprint)

		go func(email, subject, body string) {
			if err := helper.SendEmailToUser(email, subject, body); err != nil {
				fmt.Println("Failed to send email to user:", err)
			}
		}(user.Email, emailSubject, emailBody)

		tenggatPembayaranStr := ticket.TenggatPembayaran.Format("2006-01-02 15:04:05")

		if ticket.PaidStatus {
			ticket.StatusOrder = "success"
			ticket.TenggatPembayaran = nil // Set tenggat pembayaran menjadi kosong
		}

		pointMessage := "Points earned"
		if pointsEarned == 0 && ticketPurchase.KodeVoucher != "" {
			pointMessage = "Points not earned due to voucher"
		}

		var updatedTicket model.Ticket
		result = db.Where("invoice_number = ?", ticket.InvoiceNumber).First(&updatedTicket)
		if result.Error == nil && updatedTicket.PaidStatus && !user.IsAdmin {
			user.Points += pointsEarned

			if err := db.Save(&user).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user points"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		userData := map[string]interface{}{
			"name":         user.Name,
			"email":        user.Email,
			"photo_profil": user.PhotoProfil,
		}

		wisataData := map[string]interface{}{
			"title": wisata.Title,
		}

		responseData := map[string]interface{}{
			"error":                       false,
			"message":                     "Ticket Wisata purchased successfully",
			"ticket_id":                   ticket.ID,
			"invoice_number":              ticket.InvoiceNumber,
			"harga_sebelum_diskon":        hargaSebelumDiskon,
			"total_cost":                  totalCost,
			"kode_voucher":                ticketPurchase.KodeVoucher,
			"points_earned":               pointsEarned,
			"used_points":                 usedPoints,
			"carbon_footprint":            carbonFootprint,
			"point_message":               pointMessage,
			"user":                        userData,
			"wisata":                      wisataData,
			"checkin_booking":             ticketPurchase.CheckinBooking,
			"quantity":                    ticketPurchase.Quantity,
			"total_potongan_kode_voucher": totalPotonganKodeVoucher,
			"total_potongan_points":       totalPotonganPoints,
			"status_order":                ticket.StatusOrder,
			"tenggat_pembayaran":          tenggatPembayaranStr,
			"used_points_on_purchase":     usedPoints,
			"used_all_points":             ticketPurchase.UseAllPoints,
		}

		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Ticket Wisata purchased successfully",
			"data":    responseData,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func CancelTicket(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		invoiceNumber := c.Param("invoice_number")

		var ticket model.Ticket
		ticketResult := db.Where("user_id = ? AND invoice_number = ?", user.ID, invoiceNumber).First(&ticket)
		if ticketResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Ticket not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if ticket.StatusOrder == "dibatalkan" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Ticket has already been canceled"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if ticket.StatusOrder != "pending" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Cannot cancel ticket with current status"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		user.Points += ticket.UsedPointsOnPurchase
		db.Save(&user)

		ticket.StatusOrder = "dibatalkan"

		if err := db.Save(&ticket).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to cancel ticket"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Ticket canceled successfully",
		})
	}
}

func CheckTicketPrice(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var ticketPurchase struct {
			WisataID       uint   `json:"wisata_id"`
			KodeVoucher    string `json:"kode_voucher"`
			UseAllPoints   bool   `json:"use_all_points"`
			UsedPoints     int    `json:"used_points"`
			Quantity       int    `json:"quantity"`
			CheckinBooking string `json:"checkin_booking"`
		}

		if err := c.Bind(&ticketPurchase); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var wisata model.Wisata
		wisataResult := db.First(&wisata, ticketPurchase.WisataID)
		if wisataResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if ticketPurchase.Quantity <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Quantity must be greater than 0"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		checkinBookingTime, err := time.Parse("2006-01-02", ticketPurchase.CheckinBooking)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid checkin_booking date format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if checkinBookingTime.Before(time.Now().Truncate(24 * time.Hour)) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Checkin date must be today or later"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var discountPercentage int
		totalCost := wisata.Price * ticketPurchase.Quantity
		pointsEarned := totalCost / 10000
		var totalPotonganKodeVoucher int
		var totalPotonganPoints int

		hargaSebelumDiskon := totalCost

		if ticketPurchase.KodeVoucher != "" {
			var promo model.Promo
			promoResult := db.Where("kode_voucher = ?", ticketPurchase.KodeVoucher).First(&promo)
			if promoResult.Error == nil {
				currentTime := time.Now()
				if promo.StatusAktif && currentTime.Before(promo.TanggalKadaluarsa) {
					discountPercentage = promo.JumlahPotonganPersen
					if discountPercentage > 0 {
						discount := (totalCost * discountPercentage) / 100
						totalCost -= discount
						totalPotonganKodeVoucher += discount
					}
					pointsEarned = 0
				} else {
					if !promo.StatusAktif {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Voucher belum aktif"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					} else if currentTime.After(promo.TanggalKadaluarsa) {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Voucher sudah expired"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					} else {
						errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid kode voucher"}
						return c.JSON(http.StatusBadRequest, errorResponse)
					}
				}
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid kode voucher"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		var usedPoints int

		if ticketPurchase.UseAllPoints {
			// Calculate the maximum points that can be used
			maxPoints := totalCost / 1000
			if maxPoints > user.Points {
				maxPoints = user.Points
			}

			usedPoints = maxPoints
			additionalDiscount := usedPoints * 1000
			totalCost -= additionalDiscount
			totalPotonganPoints += additionalDiscount
		} else {
			usedPoints = ticketPurchase.UsedPoints
			additionalDiscount := usedPoints * 1000
			if additionalDiscount <= totalCost {
				totalCost -= additionalDiscount
				totalPotonganPoints += additionalDiscount
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough points to use"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if wisata.AvailableTickets < ticketPurchase.Quantity {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Not enough available tickets"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		carbonFootprint := CalculateCarbonFootprint(user, wisata)

		pointMessage := "Points earned"
		if pointsEarned == 0 && ticketPurchase.KodeVoucher != "" {
			pointMessage = "Points not earned due to voucher"
		}

		responseData := map[string]interface{}{
			"error":                       false,
			"message":                     "Ticket Wisata price checked successfully",
			"harga_sebelum_diskon":        hargaSebelumDiskon,
			"total_cost":                  totalCost,
			"kode_voucher":                ticketPurchase.KodeVoucher,
			"points_earned":               pointsEarned,
			"used_points":                 usedPoints,
			"carbon_footprint":            carbonFootprint,
			"point_message":               pointMessage,
			"quantity":                    ticketPurchase.Quantity,
			"total_potongan_kode_voucher": totalPotonganKodeVoucher,
			"total_potongan_points":       totalPotonganPoints,
		}

		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Ticket Wisata price checked successfully",
			"data":    responseData,
		}

		return c.JSON(http.StatusOK, response)
	}
}
