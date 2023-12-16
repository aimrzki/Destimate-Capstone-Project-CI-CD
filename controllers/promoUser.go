package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strconv"
	"time"
)

func GetPromos(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		// Get query parameters
		namaPromo := c.QueryParam("nama_promo")
		page, perPage := helper.GetPaginationParams(c)

		// Build the base query
		query := db.Model(&model.Promo{})

		if namaPromo != "" {
			query = query.Where("nama_promo LIKE ?", "%"+namaPromo+"%")
		}

		var totalPromos int64
		query.Count(&totalPromos)

		// Calculate pagination information
		var totalPages int
		if perPage > 0 {
			totalPages = int((totalPromos + int64(perPage) - 1) / int64(perPage))
		} else {
			totalPages = 0
		}

		// Retrieve paginated promos
		var promos []model.Promo
		query.Offset((page - 1) * perPage).Limit(perPage).Find(&promos)

		currentTime := time.Now()
		for i := range promos {
			if promos[i].TanggalKadaluarsa.Before(currentTime) {
				promos[i].StatusAktif = false

				// Update the status_aktif in the database
				db.Save(&promos[i])
			}
		}

		if promos == nil {
			promos = []model.Promo{}
		}

		response := map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"username": username, // Mengirimkan username pengguna (kosong jika token tidak valid)
			"promos":   promos,
			"pagination": map[string]interface{}{
				"current_page": page,
				"from":         (page-1)*perPage + 1,
				"last_page":    totalPages,
				"per_page":     perPage,
				"to":           (page-1)*perPage + len(promos),
				"total":        totalPromos,
			},
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetPromoByID(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		// Mendapatkan ID promo dari parameter URL
		promoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid promo ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var promo model.Promo
		if err := db.First(&promo, promoID).Error; err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Promo not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     http.StatusOK,
			"error":    false,
			"username": username,
			"promo":    promo,
		})
	}
}
