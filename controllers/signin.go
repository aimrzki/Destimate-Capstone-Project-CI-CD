package controllers

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func AdminSignin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user model.User
		if err := c.Bind(&user); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi apakah username dan password telah diisi
		if user.Username == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if user.Password == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Password is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username ada dalam database
		var existingUser model.User
		result := db.Where("username = ?", user.Username).First(&existingUser)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username"}
				return c.JSON(http.StatusUnauthorized, errorResponse)
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Membandingkan password yang dimasukkan dengan password yang di-hash
		err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if !existingUser.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied. User is not an admin."}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		if !existingUser.IsVerified {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Account not verified. Please verify your email before logging in."}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(existingUser.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mendapatkan informasi waktu kadaluwarsa token
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to parse token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Goroutine untuk mengirimkan notifikasi login ke email
		go func(email, username string) {
			if err := helper.SendLoginNotification(email, username); err != nil {
				fmt.Println("Failed to send notification email:", err)
			}
		}(existingUser.Email, existingUser.Username)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":        http.StatusOK,
			"error":       false,
			"message":     "Admin login successful",
			"token":       tokenString,
			"id":          existingUser.ID,
			"name":        existingUser.Name,
			"email":       existingUser.Email,
			"username":    existingUser.Username,
			"photoProfil": existingUser.PhotoProfil,
			"expires_at":  token.Claims.(jwt.MapClaims)["exp"],
		})
	}
}

func Signin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user model.User
		if err := c.Bind(&user); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi apakah username dan password telah diisi
		if user.Username == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Username is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}
		if user.Password == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Password is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Mengecek apakah username ada dalam database
		var existingUser model.User
		result := db.Where("username = ?", user.Username).First(&existingUser)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid username"}
				return c.JSON(http.StatusUnauthorized, errorResponse)
			} else {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to check username"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		// Membandingkan password yang dimasukkan dengan password yang di-hash
		err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid password"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		if existingUser.IsAdmin {
			errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Access denied. Admin cannot use this endpoint."}
			return c.JSON(http.StatusForbidden, errorResponse)
		}

		if !existingUser.IsVerified {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Account not verified. Please verify your email before logging in."}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		// Generate JWT token
		tokenString, err := middleware.GenerateToken(existingUser.Username, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to generate token"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Goroutine untuk mengirimkan email notifikasi
		go func(email, username string) {
			if err := helper.SendLoginNotification(email, username); err != nil {
				fmt.Println("Failed to send notification email:", err)
			}
		}(existingUser.Email, existingUser.Username)

		// Menyertakan ID pengguna dalam respons
		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "User login successful", "token": tokenString, "id": existingUser.ID})
	}
}
