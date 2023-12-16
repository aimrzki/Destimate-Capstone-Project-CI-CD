package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"time"
)

func GetUserDataByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		userIDStr := c.Param("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid user ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var authUser model.User
		if err := db.Where("username = ?", username).First(&authUser).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if authUser.IsAdmin || uint(userID) == authUser.ID {
			var user model.User
			if err := db.First(&user, uint(userID)).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}

			userResponse := helper.UserResponse{
				ID:               user.ID,
				Name:             user.Name,
				Username:         user.Username,
				Email:            user.Email,
				PhoneNumber:      user.PhoneNumber,
				Points:           user.Points,
				IsVerified:       user.IsVerified,
				PhotoProfil:      user.PhotoProfil,
				CategoryKesukaan: user.CategoryKesukaan,
				CategoryID:       user.CategoryID,
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"code":    http.StatusOK,
				"error":   false,
				"message": "User data retrieved successfully",
				"user":    userResponse,
			})
		}

		errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
		return c.JSON(http.StatusForbidden, errorResponse)
	}
}

func GetProfile(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		response := map[string]interface{}{
			"code":  http.StatusOK,
			"error": false,
			"profile": map[string]interface{}{
				"id":          user.ID,
				"name":        user.Name,
				"email":       user.Email,
				"username":    user.Username,
				"photoProfil": user.PhotoProfil,
				"createdAt":   user.CreatedAt,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func EditUser(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		userID := c.Param("id")

		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if user.Username != username {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized to edit this user"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		err := c.Request().ParseMultipartForm(10 << 20)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Failed to parse form data"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		newName := c.FormValue("name")
		if newName != "" && len(newName) < 3 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name should be at least 3 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		newUsername := c.FormValue("username")
		if newUsername != "" && len(newUsername) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username should be at least 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if newUsername != "" && newUsername != user.Username {
			var existingUser model.User
			if err := db.Where("username = ?", newUsername).First(&existingUser).Error; err == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username already exists"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		newEmail := c.FormValue("email")
		if newEmail != "" && !helper.IsValidEmail(newEmail) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid email format"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if newEmail != "" && newEmail != user.Email {
			var existingUser model.User
			if err := db.Where("email = ?", newEmail).First(&existingUser).Error; err == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Email already exists"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
		}

		newPhoneNumber := c.FormValue("phone_number")
		if newPhoneNumber != "" {

			// Validasi: phone_number harus memiliki minimal 10 digit
			if len(newPhoneNumber) < 10 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number should be at least 10 digits"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			// Validasi: phone_number hanya boleh mengandung digit
			if !helper.IsValidPhoneNumber(newPhoneNumber) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid phone number format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			// Periksa apakah phone_number sudah digunakan oleh pengguna lain
			if newPhoneNumber != user.PhoneNumber {
				var existingUser model.User
				if err := db.Where("phone_number = ?", newPhoneNumber).First(&existingUser).Error; err == nil {
					errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Phone number already exists"}
					return c.JSON(http.StatusBadRequest, errorResponse)
				}
			}
		}

		profileImageFile, err := c.FormFile("profile_image")
		if err == nil {
			if !helper.IsImageFile(profileImageFile) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid image format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if !helper.IsFileSizeExceeds(profileImageFile, 5*1024*1024) {
				src, err := profileImageFile.Open()
				if err == nil {
					defer src.Close()
					imageData, err := io.ReadAll(src)
					if err == nil {
						timestamp := time.Now().Unix()
						imageName := fmt.Sprintf("users/user%d/profile_%d.jpg", user.ID, timestamp)
						imageURL, err := helper.UploadImageToGCS(imageData, imageName)
						if err == nil {
							user.PhotoProfil = imageURL
						}
					}
				}
			}
		}

		name := c.FormValue("name")
		username = c.FormValue("username")
		email := c.FormValue("email")
		phoneNumber := c.FormValue("phone_number")
		categoryKesukaan := c.FormValue("category_kesukaan")

		if categoryKesukaan != "" {
			var category model.Category
			result := db.Where("category_name = ?", categoryKesukaan).First(&category)
			if result.Error != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Category not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			user.CategoryKesukaan = categoryKesukaan
			user.CategoryID = category.ID
			user.StatusCategory = true
		}

		if name != "" {
			user.Name = name
		}
		if username != "" {
			user.Username = username
		}
		if email != "" {
			user.Email = email
		}
		if phoneNumber != "" {
			user.PhoneNumber = phoneNumber
		}

		if err := db.Save(&user).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		newTokenString, err := middleware.GenerateToken(user.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		userResponse := helper.EditUserResponse{
			ID:               user.ID,
			Name:             user.Name,
			Username:         user.Username,
			Email:            user.Email,
			PhoneNumber:      user.PhoneNumber,
			PhotoProfil:      user.PhotoProfil,
			CategoryKesukaan: user.CategoryKesukaan,
			CategoryID:       user.CategoryID,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "User updated successfully",
			"user":    userResponse,
			"token":   newTokenString, // Include the new token in the response
		})
	}
}

func DeleteUserProfilePhoto(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		// Mendapatkan ID pengguna dari parameter URL
		userID := c.Param("id")

		// Mengambil data pengguna dari database berdasarkan ID
		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memastikan bahwa pengguna yang meminta penghapusan adalah pemilik akun
		if user.Username != username {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized to delete photo for this user"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Menghapus foto profil pengguna
		user.PhotoProfil = ""
		if err := db.Save(&user).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to delete user profile photo"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "User profile photo deleted successfully",
		})
	}
}

func ChangePassword(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		userID := c.Param("id")

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		if userID != fmt.Sprint(user.ID) {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		var req model.ChangePasswordRequest
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi password baru
		if len(req.NewPassword) < 8 || !helper.IsValidPassword(req.NewPassword) {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New password must be at least 8 characters and contain a combination of letters and numbers"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi konfirmasi password
		if req.NewPassword != req.ConfirmPassword {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Confirmation password does not match the new password"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi password baru tidak sama dengan password sekarang
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.NewPassword))
		if err == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "New password must be different from the current password"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi password sekarang
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Incorrect current password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Mengenkripsi password baru
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to hash new password"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Menyimpan password baru yang dienkripsi ke database
		user.Password = string(hashedNewPassword)
		db.Save(&user)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Password updated successfully"})
	}
}

func GetTotalCarbonFootprintByUser(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		var currentUser model.User
		result := db.Where("username = ?", username).First(&currentUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		userStrID := c.Param("user_id")
		userID, err := strconv.ParseUint(userStrID, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid user ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if currentUser.IsAdmin || currentUser.ID == uint(userID) {
			var totalCarbonFootprint float64
			if err := db.Model(&model.Ticket{}).Where("user_id = ? AND paid_status = ?", userID, true).Select("COALESCE(SUM(carbon_footprint), 0)").Row().Scan(&totalCarbonFootprint); err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to calculate total carbon footprint"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}

			// Rumus menulis aplikasi trip it 0,36tco2 setara dengan listrik 1 rumah selama 1 bulan
			// setelah di kalkulasi didapatkan bahwa satu rumah membutuhkan 1000 gram co2 untuk daya listrik/jam

			carbonEquivalentHours := int(totalCarbonFootprint / 1000) // 1 jam = 1000 mg CO2

			// Membulatkan nilai totalCarbonFootprint ke integer
			roundedTotalCarbonFootprint := int(totalCarbonFootprint + 0.5)

			responseData := map[string]interface{}{
				"code":                               http.StatusOK,
				"error":                              false,
				"rounded_total_carbon_footprint":     roundedTotalCarbonFootprint,
				"equivalent_powering_house_in_hours": carbonEquivalentHours,
			}

			return c.JSON(http.StatusOK, responseData)
		}

		errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied"}
		return c.JSON(http.StatusForbidden, errorResponse)
	}
}

func EditUserLocation(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		// Mendapatkan ID pengguna yang akan diedit
		userID := c.Param("id")

		var user model.User
		if err := db.First(&user, userID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		// Memeriksa otorisasi
		if user.Username != username {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Unauthorized to edit this user"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var updateLocation model.User
		if err := c.Bind(&updateLocation); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Memperbarui data Lat dan Long
		if updateLocation.Lat != 0 {
			user.Lat = updateLocation.Lat
		}
		if updateLocation.Long != 0 {
			user.Long = updateLocation.Long
		}

		if err := db.Save(&user).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user location"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		userResponse := helper.EditUserLocation{
			ID:   user.ID,
			Lat:  user.Lat,
			Long: user.Long,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "User location updated successfully",
			"user":    userResponse,
		})
	}
}

func GetUserPoints(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		// Mendapatkan informasi pengguna yang diautentikasi
		var authUser model.User
		result := db.Where("username = ?", username).First(&authUser)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mengembalikan respons dengan jumlah poin yang dimiliki oleh pengguna
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "User's points retrieved successfully",
			"points":  authUser.Points,
		})
	}
}
