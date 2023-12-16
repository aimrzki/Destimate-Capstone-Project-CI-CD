package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"time"
)

func CreatePromo(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}
		title := c.FormValue("title")
		namaPromo := c.FormValue("nama_promo")
		kodeVoucher := c.FormValue("kode_voucher")
		jumlahPotonganPersenStr := c.FormValue("jumlah_potongan_persen")
		statusAktifStr := c.FormValue("status_aktif")
		tanggalKadaluarsaStr := c.FormValue("tanggal_kadaluarsa")
		deskripsi := c.FormValue("deskripsi")
		peraturan := c.FormValue("peraturan")

		if len(title) < 5 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title must be at least 5 characters"})
		}

		if len(title) > 100 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title cannot exceed 100 characters"})
		}

		if len(namaPromo) < 5 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Nama Promo must be at least 5 characters"})
		}

		if len(namaPromo) > 100 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Nama Promo cannot exceed 100 characters"})
		}

		if len(kodeVoucher) < 5 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode voucher must be at least 5 characters"})
		}

		if len(kodeVoucher) > 40 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode Voucher cannot exceed 40 characters"})
		}

		if len(deskripsi) < 10 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi must be at least 10 characters"})
		}

		if len(deskripsi) > 2000 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi cannot exceed 2000 characters"})
		}

		// Validasi minimal 10 huruf untuk peraturan
		if len(peraturan) < 10 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Peraturan must be at least 10 characters"})
		}

		if len(peraturan) > 2000 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Peraturan cannot exceed 2000 characters"})
		}

		// Validasi agar tidak menggunakan kode_voucher yang sudah ada pada promo_model.go
		existingKodeVoucher := model.Promo{}
		if err := db.Where("kode_voucher = ?", kodeVoucher).First(&existingKodeVoucher).Error; err == nil {
			return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this name for kode_voucher already exists"})
		}

		// Validasi jumlah_potongan_persen harus di atas 0
		jumlahPotongan, err := strconv.Atoi(jumlahPotonganPersenStr)
		if err != nil || jumlahPotongan <= 0 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid jumlah_potongan_persen"})
		}

		if jumlahPotongan > 100 {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid jumlah_potongan_persen. It must not exceed 100"})
		}

		statusAktif, err := strconv.ParseBool(statusAktifStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid status_aktif"})
		}

		tanggalKadaluarsa, err := time.Parse("2006-01-02", tanggalKadaluarsaStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid tanggal_kadaluarsa format. Use YYYY-MM-DD"})
		}

		// Validasi agar tidak menggunakan title yang sudah ada pada promo_model.go
		existingTitle := model.Promo{}
		if err := db.Where("title = ?", title).First(&existingTitle).Error; err == nil {
			return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this title already exists"})
		}

		// Validasi agar tidak menggunakan nama_promo yang sudah ada pada promo_model.go
		existingNamaPromo := model.Promo{}
		if err := db.Where("nama_promo = ?", namaPromo).First(&existingNamaPromo).Error; err == nil {
			return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this nama_promo already exists"})
		}

		// Validasi tanggal_kadaluarsa tidak boleh tanggal yang sudah lewat
		if time.Now().After(tanggalKadaluarsa) {
			return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid tanggal_kadaluarsa. It must be a future date"})
		}

		randomString := helper.GenerateRandomString(10)
		newPromo := model.Promo{
			Title:                title,
			KodeVoucher:          kodeVoucher,
			JumlahPotonganPersen: jumlahPotongan,
			NamaPromo:            namaPromo,
			StatusAktif:          statusAktif,
			TanggalKadaluarsa:    tanggalKadaluarsa,
			Deskripsi:            deskripsi,
			Peraturan:            peraturan,
		}

		imageFile, err := c.FormFile("image_voucher")
		if err == nil {
			if !helper.IsImageFile(imageFile) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid image format"})
			}

			if !helper.IsFileSizeExceeds(imageFile, 5*1024*1024) {
				src, err := imageFile.Open()
				if err == nil {
					defer src.Close()
					imageData, err := io.ReadAll(src)
					if err == nil {
						timestamp := time.Now().Unix()
						imageName := fmt.Sprintf("promos/promo/image_%s_%d.jpg", randomString, timestamp)
						imageURL, err := helper.UploadImageToGCS(imageData, imageName)
						if err == nil {
							newPromo.ImageVoucher = imageURL
						}
					}
				}
			}
		}

		if err := db.Create(&newPromo).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to create promo"})
		}

		// Notifikasi kepada semua pengguna
		var users []model.User
		db.Find(&users)

		notificationMessage := fmt.Sprintf("Ada promo menarik buat kamu yang suka healing")

		for _, user := range users {
			notification := model.Notification{
				UserID:  user.ID,
				Message: notificationMessage,
				Title:   fmt.Sprintf("promo %s", newPromo.Title),
				IsRead:  false,
				Status:  "unread",
				PromoID: newPromo.ID, // Tambahkan ID promo ke notifikasi
			}
			db.Create(&notification)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Promo created successfully",
			"promo_data": newPromo,
		})
	}
}

func EditPromo(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}
		promoID := c.Param("id")

		var existingPromo model.Promo
		if err := db.First(&existingPromo, promoID).Error; err != nil {
			return c.JSON(http.StatusNotFound, helper.ErrorResponse{Code: http.StatusNotFound, Message: "Promo not found"})
		}

		title := c.FormValue("title")
		kodeVoucher := c.FormValue("kode_voucher")
		jumlahPotonganPersen := c.FormValue("jumlah_potongan_persen")
		namaPromo := c.FormValue("nama_promo")
		statusAktifStr := c.FormValue("status_aktif")
		tanggalKadaluarsaStr := c.FormValue("tanggal_kadaluarsa")
		deskripsi := c.FormValue("deskripsi")
		peraturan := c.FormValue("peraturan")

		if title != "" && title != existingPromo.Title {
			if len(title) < 5 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title must have at least 5 characters"})
			}

			if len(title) > 100 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title cannot exceed 100 characters"})
			}

			existingTitle := model.Promo{}
			if err := db.Where("title = ?", title).First(&existingTitle).Error; err == nil {
				return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this title already exists"})
			}

			existingPromo.Title = title
		}

		if namaPromo != "" && namaPromo != existingPromo.NamaPromo {
			if len(namaPromo) < 5 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "NamaPromo must have at least 5 characters"})
			}

			if len(namaPromo) > 100 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Nama Promo cannot exceed 100 characters"})
			}

			existingNamaPromo := model.Promo{}
			if err := db.Where("nama_promo = ?", namaPromo).First(&existingNamaPromo).Error; err == nil {
				return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this nama_promo already exists"})
			}

			existingPromo.NamaPromo = namaPromo
		}

		if kodeVoucher != "" && kodeVoucher != existingPromo.KodeVoucher {
			if len(kodeVoucher) < 5 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "KodeVoucher must have at least 5 characters"})
			}

			if len(kodeVoucher) > 40 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode Voucher cannot exceed 40 characters"})
			}

			existingKodeVoucher := model.Promo{}
			if err := db.Where("kode_voucher = ?", kodeVoucher).First(&existingKodeVoucher).Error; err == nil {
				return c.JSON(http.StatusConflict, helper.ErrorResponse{Code: http.StatusConflict, Message: "Promo with this kode_voucher already exists"})
			}

			existingPromo.KodeVoucher = kodeVoucher
		}

		if jumlahPotonganPersen != "" {
			jumlahPotongan, err := strconv.Atoi(jumlahPotonganPersen)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid jumlah_potongan_persen"})
			}

			if jumlahPotongan <= 0 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "JumlahPotonganPersen must be greater than 0"})
			}

			if jumlahPotongan > 100 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid jumlah_potongan_persen. It must not exceed 100"})
			}

			existingPromo.JumlahPotonganPersen = jumlahPotongan
		}

		if deskripsi != "" && deskripsi != existingPromo.Deskripsi {
			if len(deskripsi) < 10 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi must be at least 10 characters"})
			}

			if len(deskripsi) > 2000 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi cannot exceed 2000 characters"})
			}

			existingPromo.Deskripsi = deskripsi
		}

		if peraturan != "" && peraturan != existingPromo.Peraturan {
			if len(peraturan) < 10 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Peraturan must be at least 10 characters"})
			}

			if len(peraturan) > 2000 {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Peraturan cannot exceed 2000 characters"})
			}
			existingPromo.Peraturan = peraturan
		}

		if tanggalKadaluarsaStr != "" {
			tanggalKadaluarsa, err := time.Parse("2006-01-02", tanggalKadaluarsaStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid tanggal_kadaluarsa format. Use YYYY-MM-DD"})
			}
			if time.Now().After(tanggalKadaluarsa) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid tanggal_kadaluarsa. It must be a future date"})
			}
			existingPromo.TanggalKadaluarsa = tanggalKadaluarsa
		}

		if statusAktifStr != "" {
			statusAktif, err := strconv.ParseBool(statusAktifStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid status_aktif"})
			}
			existingPromo.StatusAktif = statusAktif
		}

		imageFile, err := c.FormFile("image_voucher")
		if err == nil {
			if !helper.IsImageFile(imageFile) {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid image format"})
			}

			if !helper.IsFileSizeExceeds(imageFile, 5*1024*1024) {
				src, err := imageFile.Open()
				if err == nil {
					defer src.Close()
					imageData, err := io.ReadAll(src)
					if err == nil {
						timestamp := time.Now().Unix()
						imageName := fmt.Sprintf("promos/promo%d/image_%d.jpg", existingPromo.ID, timestamp)
						imageURL, err := helper.UploadImageToGCS(imageData, imageName)
						if err == nil {
							existingPromo.ImageVoucher = imageURL
						}
					}
				}
			}
		}

		if err := db.Save(&existingPromo).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update promo"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Promo updated successfully",
			"promo_data": existingPromo,
		})
	}
}

func DeletePromoByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		promoID, err := helper.ConvertParamToUint(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid promo ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingPromo model.Promo
		result := db.First(&existingPromo, promoID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Promo not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingPromo)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Promo deleted successfully"})
	}
}
