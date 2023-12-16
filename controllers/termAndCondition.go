package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func GetAllTermCondition(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Get query parameters
		page, perPage := helper.GetPaginationParams(c)

		// Build the base query
		query := db.Model(&model.TermCondition{})

		var totalTermConditions int64
		query.Count(&totalTermConditions)

		// Calculate pagination information
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalTermConditions + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		// Retrieve paginated term conditions
		var termConditions []model.TermCondition
		query.Offset((page - 1) * perPage).Limit(perPage).Find(&termConditions)

		if termConditions == nil {
			termConditions = []model.TermCondition{}
		}

		response := map[string]interface{}{
			"code":            http.StatusOK,
			"error":           false,
			"term_conditions": termConditions,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(termConditions),
				"total":        totalTermConditions,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetTermConditionByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)
		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Mendapatkan ID dari parameter URL
		id := c.Param("id")

		var term model.TermCondition
		if err := db.First(&term, id).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "TermCondition not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "term_condition": term})
	}
}
