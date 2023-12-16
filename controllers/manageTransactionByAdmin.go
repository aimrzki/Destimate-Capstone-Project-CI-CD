package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"time"
)

type TicketResponse struct {
	Ticket      model.Ticket `json:"ticket"`
	UserProfile UserProfile  `json:"user_profile"`
}

type UserProfile struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	PhotoProfil string `json:"photo_profil"`
}

type TicketUserDetail struct {
	TicketID          uint       `json:"ticket_id"`
	UserID            uint       `json:"user_id"`
	Name              string     `json:"name"`
	PhotoProfil       string     `json:"photo_profil"`
	WisataID          uint       `json:"wisata_id"`
	WisataTitle       string     `json:"wisata_title"`
	TotalCost         int        `json:"total_cost"`
	InvoiceNumber     string     `json:"invoice_number"`
	CheckInBooking    *time.Time `json:"check_in_booking"`
	CreatedAt         *time.Time `json:"created_at"`
	KodeVoucher       string     `json:"kode_voucher"`
	PaidStatus        bool       `json:"paid_status"`
	StatusOrder       string     `json:"status_order"`
	TenggatPembayaran *time.Time `json:"tenggat_pembayaran"`
}

func GetAllTicketsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}
		searchQuery := c.QueryParam("search")
		page, perPage := helper.GetPaginationParams(c)

		query := db.Model(&model.Ticket{}).Order("created_at DESC")

		if searchQuery != "" {
			query = query.
				Joins("JOIN users ON tickets.user_id = users.id").
				Joins("JOIN wisata ON tickets.wisata_id = wisata.id").
				Where("tickets.invoice_number LIKE ? OR users.name LIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%").
				Or("wisata.title LIKE ?", "%"+searchQuery+"%")
		}

		var totalTickets int64
		query.Count(&totalTickets)

		// Calculate pagination information
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalTickets + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		// Retrieve paginated tickets
		var tickets []model.Ticket
		query.Offset((page - 1) * perPage).Limit(perPage).Find(&tickets)

		// Create a slice to store additional data in the response
		var ticketDetails []map[string]interface{}

		// Iterate through each ticket and add additional information
		for _, ticket := range tickets {
			var wisata model.Wisata
			db.First(&wisata, ticket.WisataID)

			var user model.User
			db.First(&user, ticket.UserID)

			ticketDetail := map[string]interface{}{
				"id":                 ticket.ID,
				"wisata_id":          ticket.WisataID,
				"wisata_title":       wisata.Title,
				"user_id":            ticket.UserID,
				"user_name":          user.Name,
				"photo_profil":       user.PhotoProfil,
				"used_points":        ticket.UsedPoints,
				"use_all_points":     ticket.UseAllPoints,
				"total_cost":         ticket.TotalCost,
				"invoice_number":     ticket.InvoiceNumber,
				"quantity":           ticket.Quantity,
				"created_at":         ticket.CreatedAt,
				"updated_at":         ticket.UpdatedAt,
				"carbon_footprint":   ticket.CarbonFootprint,
				"paid_status":        ticket.PaidStatus,
				"status_order":       ticket.StatusOrder,
				"tenggat_pembayaran": ticket.TenggatPembayaran,
				"checkin_booking":    ticket.CheckinBooking,
				"kode_voucher":       ticket.KodeVoucher,
				"points_earned":      ticket.PointsEarned,
			}

			ticketDetails = append(ticketDetails, ticketDetail)
		}

		if ticketDetails == nil {
			ticketDetails = []map[string]interface{}{}
		}

		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"tickets": ticketDetails,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(tickets),
				"total":        totalTickets,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetTicketByInvoiceNumber(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		invoiceNumber := c.Param("invoiceNumber")

		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var ticket model.Ticket
		result := db.Where("invoice_number = ?", invoiceNumber).First(&ticket)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Ticket not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var userDetail model.User
		userResult := db.First(&userDetail, ticket.UserID)
		if userResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var wisata model.Wisata
		wisataResult := db.First(&wisata, ticket.WisataID)
		if wisataResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch event data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		ticketDetail := TicketUserDetail{
			TicketID:          ticket.ID,
			UserID:            ticket.UserID,
			Name:              userDetail.Name,
			PhotoProfil:       userDetail.PhotoProfil,
			WisataID:          ticket.WisataID,
			WisataTitle:       wisata.Title,
			TotalCost:         ticket.TotalCost,
			InvoiceNumber:     ticket.InvoiceNumber,
			CheckInBooking:    ticket.CheckinBooking,
			CreatedAt:         ticket.CreatedAt,
			KodeVoucher:       ticket.KodeVoucher,
			PaidStatus:        ticket.PaidStatus,
			StatusOrder:       ticket.StatusOrder,
			TenggatPembayaran: ticket.TenggatPembayaran,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Ticket details retrieved successfully",
			"ticket_data": ticketDetail,
		})
	}
}

func UpdatePaidStatus(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		invoiceNumber := c.Param("invoiceId")

		var requestBody struct {
			PaidStatus bool `json:"paid_status"`
		}

		if err := c.Bind(&requestBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var ticket model.Ticket
		result := db.Where("invoice_number = ?", invoiceNumber).First(&ticket)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Ticket not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var user model.User
		result = db.First(&user, ticket.UserID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if ticket.StatusOrder == "dibatalkan" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Cannot update paid status for canceled ticket"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var notification model.Notification
		var wisata model.Wisata
		wisataResult := db.First(&wisata, ticket.WisataID)
		if wisataResult.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		notificationMessage := fmt.Sprintf("Tiket untuk wisata %s berhasil dibayar. Selamat liburan!", wisata.Title)
		notification = model.Notification{
			UserID:        user.ID,
			Message:       notificationMessage,
			Title:         "Transaksi Sukses",   // Tambahkan title
			InvoiceNumber: ticket.InvoiceNumber, // Tambahkan invoice_number
		}
		db.Create(&notification)

		if requestBody.PaidStatus {
			ticket.StatusOrder = "success"
			ticket.TenggatPembayaran = nil

			var wisata model.Wisata
			if ticket.PaidStatus {
				wisataResult := db.First(&wisata, ticket.WisataID)
				if wisataResult.Error != nil {
					errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
					return c.JSON(http.StatusNotFound, errorResponse)
				}

				notificationMessage := fmt.Sprintf("Tiket untuk wisata %s berhasil dibayar. Selamat liburan!", wisata.Title)
				notification = model.Notification{
					UserID:  user.ID,
					Message: notificationMessage,
				}
				db.Create(&notification)
			}

			if ticket.PointsEarned > 0 && requestBody.PaidStatus {
				user.Points += ticket.PointsEarned
				db.Save(&user)
			}
		} else {
			ticket.StatusOrder = "pending"
			ticket.TenggatPembayaran = ticket.CheckinBooking
		}

		ticket.PaidStatus = requestBody.PaidStatus
		db.Save(&ticket)

		var userProfile UserProfile
		db.Model(&user).Select("id as user_id, username, photo_profil").Scan(&userProfile)

		var ticketResponse TicketResponse
		ticketResponse.Ticket = ticket
		ticketResponse.UserProfile = userProfile

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Paid status updated successfully",
			"ticket":  ticketResponse,
		})
	}
}

func DeleteTicketByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}
		// Mendapatkan invoice_number dari parameter URL
		invoiceNumber := c.Param("invoice_number")
		// Menghapus tiket berdasarkan invoice_number
		result := db.Where("invoice_number = ?", invoiceNumber).Delete(&model.Ticket{})
		if result.Error != nil || result.RowsAffected == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Ticket not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Ticket deleted successfully"})
	}
}
