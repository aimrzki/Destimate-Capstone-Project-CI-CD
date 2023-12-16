package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
)

func GetWisatas(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
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
		searchQuery := c.QueryParam("search")

		query := db.Model(&model.Wisata{}).
			Preload("Category").
			Joins("JOIN categories ON wisata.category_id = categories.id")

		// Add searching condition
		if searchQuery != "" {
			query = query.Where("title LIKE ? OR kota LIKE ? OR category_name LIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%", "%"+searchQuery+"%")
		}

		var totalWisatas int64
		query.Count(&totalWisatas)

		// Calculate pagination information
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalWisatas + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		// Retrieve paginated wisatas
		var wisatas []model.Wisata
		query.Offset((page - 1) * perPage).Limit(perPage).Find(&wisatas)

		if wisatas == nil {
			wisatas = []model.Wisata{}
		}

		response := map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"wisatas": wisatas,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(wisatas),
				"total":        totalWisatas,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetWisataByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		wisataIDParam := c.Param("id")

		wisataID, err := strconv.ParseUint(wisataIDParam, 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid wisata ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var wisata model.Wisata
		if err := db.Preload("Category").First(&wisata, wisataID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
				return c.JSON(http.StatusNotFound, errorResponse)
			}
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch wisata"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var totalCarbonFootprint float64
		if err := db.Model(&model.Ticket{}).Select("COALESCE(SUM(carbon_footprint), 0)").Where("wisata_id = ? AND paid_status = ?", wisataID, true).Row().Scan(&totalCarbonFootprint); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to calculate total carbon footprint"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		response := map[string]interface{}{
			"code":                   http.StatusOK,
			"error":                  false,
			"wisata":                 wisata,
			"total_carbon_footprint": totalCarbonFootprint,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetTotalCarbonFootprintByWisataID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		wisataID := c.Param("wisata_id")
		if wisataID == "" {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Wisata ID is required"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var wisata model.Wisata
		if err := db.Where("id = ?", wisataID).First(&wisata).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Wisata not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var totalCarbonFootprint float64
		if err := db.Model(&model.Ticket{}).Select("COALESCE(SUM(carbon_footprint), 0)").Where("wisata_id = ? AND paid_status = ?", wisataID, true).Row().Scan(&totalCarbonFootprint); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to calculate total carbon footprint"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		if totalCarbonFootprint == 0 {
			return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "total_carbon_footprint": 0})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "total_carbon_footprint": totalCarbonFootprint})
	}
}

func GetWisataByCategoryKesukaan(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		// Dapatkan wisata berdasarkan category_kesukaan user
		var wisatas []model.Wisata

		if user.CategoryID == 0 || user.CategoryKesukaan == "" {
			// Jika category_kesukaan user kosong, tampilkan semua wisata
			if err := db.Preload("Category").Find(&wisatas).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch all wisatas"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		} else {
			// Jika category_kesukaan user tidak kosong, tampilkan wisata berdasarkan category_kesukaan
			if err := db.Preload("Category").Where("category_id = ?", user.CategoryID).Find(&wisatas).Error; err != nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch wisatas by category_kesukaan"}
				return c.JSON(http.StatusInternalServerError, errorResponse)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"wisatas": wisatas,
		})
	}
}
