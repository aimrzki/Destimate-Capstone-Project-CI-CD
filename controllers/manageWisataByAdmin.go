package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"io"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// nambahin data sesuai model wisata

func CreateWisata(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}
		kode := c.FormValue("kode")
		title := c.FormValue("title")
		location := c.FormValue("location")
		kota := c.FormValue("kota")
		description := c.FormValue("description")
		price := c.FormValue("price")
		availableTickets := c.FormValue("available_tickets")
		lat := c.FormValue("lat")
		long := c.FormValue("long")
		categoryName := c.FormValue("category_name")
		fasilitasStr := c.FormValue("fasilitas")
		priceInt, err := strconv.Atoi(price)

		if len(title) < 8 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Judul harus minimal 8 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(title) > 100 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Judul harus maksimal 100 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		existingWisata := model.Wisata{}
		result := db.Where("title = ?", title).First(&existingWisata)
		if result.RowsAffected > 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Judul sudah digunakan"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(kode) < 3 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode harus minimal 3 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(kode) > 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode harus maksimal 5 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		existingKode := model.Wisata{}
		result = db.Where("kode = ?", kode).First(&existingKode)
		if result.RowsAffected > 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode sudah digunakan"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(kota) < 4 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kota harus minimal 4 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(kota) > 30 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kota harus maksimal 30 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(location) < 8 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Lokasi harus minimal 8 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(location) > 200 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Lokasi harus maksimal 200 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(description) < 10 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi harus minimal 10 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(description) > 2000 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi harus maksimal 2000 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if priceInt <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Price must be greater than 0"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid price value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		availableTicketsInt, err := strconv.Atoi(availableTickets)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid available_tickets value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if availableTicketsInt <= 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Available Tickets harus lebih dari 0"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(fasilitasStr) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Fasilitas harus lebih dari 5 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(fasilitasStr) > 100 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Fasilitas harus maksimal 100 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if title == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Judul harus diisi"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if location == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Lokasi harus diisi"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if kota == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kota harus diisi"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if description == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Deskripsi harus diisi"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		latFloat, err := strconv.ParseFloat(lat, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid latitude value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		longFloat, err := strconv.ParseFloat(long, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid longitude value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var category model.Category
		result = db.Where("category_name = ?", categoryName).First(&category)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		//mengubah string menjadi potongan slice
		fasilitas := strings.Split(fasilitasStr, ",")
		if len(fasilitas) == 0 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Fasilitas harus di isi minimal satu "}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Menghapus karakter escape ganda dari awal dan akhir string
		for i, f := range fasilitas {
			fasilitas[i] = strings.Trim(f, "\"")
		}

		// Digunakan untuk mengubah atau konversi dari slice menjadi format slice
		fasilitasJSON, err := json.Marshal(fasilitas)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Gagal mengonversi fasilitas ke JSON"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		mapsLink := c.FormValue("maps_link")
		isopenStr := c.FormValue("is_open")
		isopen, err := strconv.ParseBool(isopenStr)
		descriptionIsOpen := c.FormValue("description_is_open")
		videoLink := c.FormValue("video_link")

		if isopenStr == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Field 'is_open' harus diisi"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid is_open value"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(mapsLink) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Link maps harus lebih dari 5 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(mapsLink) > 200 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Link maps harus maksimal 200 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if mapsLink == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Anda harus memasukan link maps_link"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(descriptionIsOpen) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "DescriptionIsOpen harus lebih dari 5 huruf"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(descriptionIsOpen) > 40 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "DescriptionIsOpen harus maksimal 40 karakter"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if descriptionIsOpen == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Anda harus memasukan description_is_open, keterangan buka jam berapa sampai jam berapa"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		randomString := helper.GenerateRandomString(10)

		createdWisata := model.Wisata{
			Kode:              kode,
			Title:             title,
			Location:          location,
			Kota:              kota,
			Description:       description,
			Price:             priceInt,
			AvailableTickets:  availableTicketsInt,
			Lat:               latFloat,
			Long:              longFloat,
			CreatedAt:         &[]time.Time{time.Now()}[0],
			CategoryID:        category.ID,
			MapsLink:          mapsLink,
			IsOpen:            isopen,
			DescriptionIsOpen: descriptionIsOpen,
			Fasilitas:         string(fasilitasJSON),
			VideoLink:         videoLink,
		}

		var imageUrls []string
		var photoNames []string
		for i := 1; i <= 3; i++ {
			imageFormField := fmt.Sprintf("photo_wisata%d", i)
			imageFile, err := c.FormFile(imageFormField)
			if err != nil {
				if i == 1 {
					errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Setidaknya satu gambar diperlukan."}
					return c.JSON(http.StatusBadRequest, errorResponse)
				}
				break
			}

			if !helper.IsImageFile(imageFile) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Jenis file tidak valid. Hanya file gambar yang diperbolehkan."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if helper.IsFileSizeExceeds(imageFile, 5*1024*1024) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Ukuran file melebihi batas yang diizinkan (5MB)."}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			src, err := imageFile.Open()
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Gagal membuka file gambar"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			defer src.Close()

			imageData, err := io.ReadAll(src)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Gagal membaca data gambar"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			timestamp := time.Now().Unix()
			imageName := fmt.Sprintf("wisatas/wisata/image_%s_%d_%d.jpg", randomString, timestamp, i)
			imageURL, err := helper.UploadImageToGCS(imageData, imageName)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal mengunggah gambar ke GCS"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}

			imageUrls = append(imageUrls, imageURL)
			photoNames = append(photoNames, imageName)
		}

		if len(imageUrls) >= 1 {
			createdWisata.PhotoWisata1 = imageUrls[0]
		}
		if len(imageUrls) >= 2 {
			createdWisata.PhotoWisata2 = imageUrls[1]
		}
		if len(imageUrls) >= 3 {
			createdWisata.PhotoWisata3 = imageUrls[2]
		}

		if err := db.Create(&createdWisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal menyimpan wisata"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if err := db.Preload("Category").First(&createdWisata, createdWisata.ID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal memuat data kategori"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Wisata berhasil dibuat",
			"wisataData": createdWisata,
		})
	}
}

func UpdateWisata(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		wisataID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid Wisata ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingWisata model.Wisata
		result := db.First(&existingWisata, wisataID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		kode := c.FormValue("kode")
		title := c.FormValue("title")
		location := c.FormValue("location")
		kota := c.FormValue("kota")
		description := c.FormValue("description")
		price := c.FormValue("price")
		lat := c.FormValue("lat")
		long := c.FormValue("long")
		availableTickets := c.FormValue("available_tickets")
		categoryName := c.FormValue("category_name")
		mapsLink := c.FormValue("maps_link")
		isOpenStr := c.FormValue("is_open")
		descriptionIsOpen := c.FormValue("description_is_open")
		fasilitasStr := c.FormValue("fasilitas")
		videoLink := c.FormValue("video_link")

		if fasilitasStr != "" {
			fasilitas := strings.Split(fasilitasStr, ",")
			for _, f := range fasilitas {
				if len(strings.Trim(f, "\"")) < 5 {
					errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Setiap fasilitas harus minimal terdiri dari 5 huruf"}
					return c.JSON(http.StatusBadRequest, errorResponse)
				}

				if len(fasilitasStr) > 100 {
					errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Fasilitas harus maksimal 100 karakter"}
					return c.JSON(http.StatusBadRequest, errorResponse)
				}
			}
			fasilitasJSON, err := json.Marshal(fasilitas)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Gagal mengonversi fasilitas ke JSON"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Fasilitas = string(fasilitasJSON)
		}

		if kode != "" {
			if len(kode) < 3 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode harus minimal terdiri dari 3 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			var existingKodeCount int64
			db.Model(&model.Wisata{}).Where("kode = ?", kode).Not("id = ?", wisataID).Count(&existingKodeCount)
			if existingKodeCount > 0 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kode sudah ada, pilih Kode lain"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Kode = kode
		}

		if title != "" {
			if len(title) < 8 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title harus minimal terdiri dari 8 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(title) > 100 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title harus maksimal 100 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			var existingTitleCount int64
			db.Model(&model.Wisata{}).Where("title = ?", title).Not("id = ?", wisataID).Count(&existingTitleCount)
			if existingTitleCount > 0 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Title sudah ada, pilih title lain"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Title = title
		}

		if location != "" {
			if len(location) < 8 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Location harus minimal terdiri dari 8 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(location) > 200 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Location harus maksimal 200 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if location != "" {
			existingWisata.Location = location
		}

		if kota != "" {
			if len(kota) < 4 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kota harus minimal terdiri dari 4 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(kota) > 30 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kota harus maksimal 30 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if kota != "" {
			existingWisata.Kota = kota
		}

		if description != "" {
			if len(description) < 10 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description harus minimal terdiri dari 10 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(description) > 2000 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description harus maksimal 2000 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if description != "" {
			existingWisata.Description = description
		}

		if price != "" {
			priceInt, err := strconv.Atoi(price)
			if err != nil || priceInt <= 0 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Price harus lebih besar dari 0"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Price = priceInt
		}

		if lat != "" {
			latFloat, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid latitude value"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Lat = latFloat
		}
		if long != "" {
			longFloat, err := strconv.ParseFloat(long, 64)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid longitude value"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.Long = longFloat
		}

		if availableTickets != "" {
			availableTicketsInt, err := strconv.Atoi(availableTickets)
			if err != nil || availableTicketsInt <= 0 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Available Tickets harus lebih dari 0"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.AvailableTickets = availableTicketsInt
		}

		if categoryName != "" {
			var existingCategory model.Category
			if err := db.Where("category_name = ?", categoryName).First(&existingCategory).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Category not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			existingWisata.Category = existingCategory
		}

		if mapsLink != "" {
			if len(mapsLink) < 5 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Maps Link harus minimal terdiri dari 5 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(mapsLink) > 200 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Maps Link harus maksimal 200 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if mapsLink != "" {
			existingWisata.MapsLink = mapsLink
		}

		if videoLink != "" {
			existingWisata.VideoLink = videoLink
		}

		if isOpenStr != "" {
			isOpen, err := strconv.ParseBool(isOpenStr)
			if err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid is_open value"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			existingWisata.IsOpen = isOpen
		}

		if descriptionIsOpen != "" {
			if len(descriptionIsOpen) < 5 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description Is Open harus minimal terdiri dari 5 huruf"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(descriptionIsOpen) > 40 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description Is Open harus maksimal 40 karakter"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		if descriptionIsOpen != "" {
			existingWisata.DescriptionIsOpen = descriptionIsOpen
		}

		for i := 1; i <= 3; i++ {
			imageFormField := fmt.Sprintf("photo_wisata%d", i)
			imageFile, err := c.FormFile(imageFormField)
			if err == nil {
				if helper.IsImageFile(imageFile) {
					if !helper.IsFileSizeExceeds(imageFile, 5*1024*1024) {
						src, err := imageFile.Open()
						if err == nil {
							defer src.Close()
							imageData, err := io.ReadAll(src)
							if err == nil {
								timestamp := time.Now().Unix()
								imageName := fmt.Sprintf("wisata%d/photo%d_%d.jpg", wisataID, i, timestamp)
								imageURL, err := helper.UploadImageToGCS(imageData, imageName)
								if err == nil {
									if i == 1 {
										existingWisata.PhotoWisata1 = imageURL
									} else if i == 2 {
										existingWisata.PhotoWisata2 = imageURL
									} else if i == 3 {
										existingWisata.PhotoWisata3 = imageURL
									}
								}
							}
						}
					}
				}
			}
		}

		if err := db.Save(&existingWisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to save changes to Wisata"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Preload Category
		result = db.Preload("Category").First(&existingWisata, wisataID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"message":    "Wisata updated successfully",
			"wisataData": existingWisata,
		})
	}
}

func DeleteWisata(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		// Mendapatkan ID wisata dari parameter URL
		wisataID := c.Param("id")

		// Cek apakah wisata dengan ID tersebut ada dalam basis data
		var existingWisata model.Wisata
		if err := db.Where("id = ?", wisataID).First(&existingWisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Menghapus wisata dari basis data
		if err := db.Delete(&existingWisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete Wisata"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengembalikan respons sukses jika berhasil
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Wisata deleted successfully",
		})
	}
}
