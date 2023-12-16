package middleware

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/model"
	"net/http"
	"strings"
)

func AuthenticateAndAuthorize(c echo.Context, db *gorm.DB, secretKey []byte) (*model.User, error) {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Authorization token is missing")
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
	}

	tokenString = authParts[1]

	username, err := VerifyToken(tokenString, secretKey)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	var user model.User
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	if !user.IsAdmin {
		return nil, echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	return &user, nil
}

func ExtractUsernameFromToken(c echo.Context, secretKey []byte) string {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return ""
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return ""
	}

	tokenString = authParts[1]

	username, err := VerifyToken(tokenString, secretKey)
	if err != nil {
		// Token tidak valid, tetapkan username menjadi string kosong
		username = ""
	}

	return username
}
