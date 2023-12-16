package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func GetAllUsersByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan nilai query parameter "name"
		name := c.QueryParam("name")

		// Mendapatkan nilai query parameter "page" dan "per_page"
		page, perPage := helper.GetPaginationParams(c)

		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var users []model.User
		query := db.Where("is_admin = ?", false)
		if name != "" {
			// Jika parameter "name" diisi, lakukan pencarian berdasarkan nama user
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		var totalUsers int64
		query.Model(&model.User{}).Count(&totalUsers)

		// Hitung informasi pagination
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalUsers + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		query.Offset((page - 1) * perPage).Limit(perPage).Find(&users)

		// Buat slice untuk menyimpan data yang akan dikirim sebagai respons
		var userResponses []helper.UserResponse

		// Konversi data dari model User ke UserResponse
		for _, user := range users {
			userResponse := helper.UserResponse{
				ID:               user.ID,
				Name:             user.Name,
				Username:         user.Username,
				Email:            user.Email,
				PhoneNumber:      user.PhoneNumber,
				Points:           user.Points,
				IsVerified:       user.IsVerified,
				CreatedAt:        user.CreatedAt,
				CategoryID:       user.CategoryID,
				CategoryKesukaan: user.CategoryKesukaan,
				PhotoProfil:      user.PhotoProfil,
			}
			userResponses = append(userResponses, userResponse)
		}

		// Jika tidak ada hasil, atur userResponses menjadi slice kosong
		if len(userResponses) == 0 {
			userResponses = []helper.UserResponse{}
		}

		response := map[string]interface{}{
			"code":  http.StatusOK,
			"error": false,
			"users": userResponses,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(userResponses),
				"total":        totalUsers,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetAllAdminsByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Mendapatkan nilai query parameter "name"
		name := c.QueryParam("name")

		// Mendapatkan nilai query parameter "page" dan "per_page"
		page, perPage := helper.GetPaginationParams(c)

		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var users []model.User
		query := db.Where("is_admin = ?", true)
		if name != "" {
			// Jika parameter "name" diisi, lakukan pencarian berdasarkan nama user
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		var totalUsers int64
		query.Model(&model.User{}).Count(&totalUsers)

		// Hitung informasi pagination
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalUsers + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		query.Offset((page - 1) * perPage).Limit(perPage).Find(&users)

		// Buat slice untuk menyimpan data yang akan dikirim sebagai respons
		var userResponses []helper.UserResponse

		// Konversi data dari model User ke UserResponse
		for _, user := range users {
			userResponse := helper.UserResponse{
				ID:               user.ID,
				Name:             user.Name,
				Username:         user.Username,
				Email:            user.Email,
				PhoneNumber:      user.PhoneNumber,
				Points:           user.Points,
				IsVerified:       user.IsVerified,
				CreatedAt:        user.CreatedAt,
				CategoryID:       user.CategoryID,
				CategoryKesukaan: user.CategoryKesukaan,
				PhotoProfil:      user.PhotoProfil,
			}
			userResponses = append(userResponses, userResponse)
		}

		// Jika tidak ada hasil, atur userResponses menjadi slice kosong
		if len(userResponses) == 0 {
			userResponses = []helper.UserResponse{}
		}

		response := map[string]interface{}{
			"code":  http.StatusOK,
			"error": false,
			"users": userResponses,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(userResponses),
				"total":        totalUsers,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

// EditUserByAdmin handles the logic for an admin to edit user data.
func EditUserByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		userID := c.Param("id")

		var user model.User
		result := db.First(&user, userID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var req model.User
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validate username
		if req.Username != "" && req.Username != user.Username {
			if len(req.Username) < 5 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username must be at least 5 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			var existingUserByUsername model.User
			if db.Where("username = ?", req.Username).Not("id = ?", user.ID).First(&existingUserByUsername).Error == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Username is already taken"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
			user.Username = req.Username
		}

		// Validate phone_number
		if req.PhoneNumber != "" && req.PhoneNumber != user.PhoneNumber {
			if !helper.IsValidPhoneNumber(req.PhoneNumber) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid phone number must have minimum 10 digits of numbers only"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			var existingUserByPhone model.User
			if db.Where("phone_number = ?", req.PhoneNumber).Not("id = ?", user.ID).First(&existingUserByPhone).Error == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Phone number is already taken"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
			user.PhoneNumber = req.PhoneNumber
		}

		// Validate email
		if req.Email != "" && req.Email != user.Email {
			if !helper.IsValidEmail(req.Email) {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid email format"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			var existingUserByEmail model.User
			if db.Where("email = ?", req.Email).Not("id = ?", user.ID).First(&existingUserByEmail).Error == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Email is already taken"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
			user.Email = req.Email
		}

		// Validate name
		if req.Name != "" && req.Name != user.Name {
			if len(req.Name) < 3 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name must be at least 3 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}
			user.Name = req.Name
		}

		// Update user data
		if err := db.Save(&user).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to update user"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "User updated successfully"})
	}
}

func DeleteUserByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		userID := c.Param("id")

		var user model.User
		result := db.First(&user, userID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "User not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&user)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "User deleted successfully"})
	}
}

func DeleteAdminByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		userID := c.Param("id")

		var user model.User
		result := db.First(&user, userID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&user)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Admin deleted successfully"})
	}
}
