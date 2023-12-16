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

const adminEmail = "hidestimate@gmail.com"

func CreateCooperationMessage(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cooperationMessage model.CooperationMessage

		if err := c.Bind(&cooperationMessage); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi name
		if len(cooperationMessage.FirstName) < 3 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "first name harus minimal 3 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi email format
		if !helper.IsValidEmail(cooperationMessage.Email) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Format email tidak valid"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if !helper.IsValidPhoneNumber(cooperationMessage.PhoneNumber) {
			if !helper.ContainsOnlyDigits(cooperationMessage.PhoneNumber) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number harus mengandung angka semua"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			} else if len(cooperationMessage.PhoneNumber) < 10 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number kurang dari 10 digit"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		// Validasi message
		if len(cooperationMessage.Message) < 10 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Message harus minimal 10 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Set waktu saat ini sebagai waktu pembuatan
		cooperationMessage.CreatedAt = time.Now()

		// Simpan pesan ke database
		db.Create(&cooperationMessage)

		// Kirim email ke admin
		adminEmailSubject := "New Cooperation Message"
		adminEmailBody := helper.GetCooperationEmailBody(cooperationMessage)
		if err := helper.SendEmailToUser(adminEmail, adminEmailSubject, adminEmailBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send email to admin"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Kirim email balasan ke pengirim pesan
		userEmailSubject := "Your Cooperation Message"
		userEmailBody := helper.GetUserCooperationEmailBody(cooperationMessage)
		if err := helper.SendEmailToUser(cooperationMessage.Email, userEmailSubject, userEmailBody); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to send email to user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusCreated, map[string]interface{}{"code": http.StatusCreated, "error": false, "message": "Pesan kerjasama berhasil dikirim"})
	}
}

func GetCooperationMessagesByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var cooperationMessages []model.CooperationMessage
		result := db.Find(&cooperationMessages)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Internal Server Error"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "cooperation_messages": cooperationMessages})
	}
}
