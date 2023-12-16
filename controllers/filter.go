package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strings"
)

func GetCities(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")
		if tokenString == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Authorization token is missing"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		authParts := strings.SplitN(tokenString, " ", 2)
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token format"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		tokenString = authParts[1]

		_, err := middleware.VerifyToken(tokenString, secretKey)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid token"}
			return c.JSON(http.StatusUnauthorized, errorResponse)
		}

		var cities []string

		// Menggunakan COALESCE untuk mengatasi nilai NULL
		if err := db.Model(&model.Wisata{}).Distinct("COALESCE(kota, 'Unknown')").Pluck("COALESCE(kota, 'Unknown')", &cities).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch cities"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "cities": cities})
	}
}

func GetCategories(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get query parameters
		categoryName := c.QueryParam("category_name")
		page, perPage := helper.GetPaginationParams(c)

		// Build the base query
		query := db.Model(&model.Category{})

		if categoryName != "" {
			query = query.Where("category_name LIKE ?", "%"+categoryName+"%")
		}

		var totalCategories int64
		query.Count(&totalCategories)

		// Calculate pagination information
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalCategories + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		// Retrieve paginated categories
		var categories []model.Category
		query.Offset((page - 1) * perPage).Limit(perPage).Find(&categories)

		if categories == nil {
			categories = []model.Category{}
		}

		response := map[string]interface{}{
			"code":       http.StatusOK,
			"error":      false,
			"categories": categories,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(categories),
				"total":        totalCategories,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}
