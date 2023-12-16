package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"time"
)

func GetAdminDashboardData(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		startDateStr := c.QueryParam("start_date")
		endDateStr := c.QueryParam("end_date")

		var startDate, endDate time.Time
		var errDate error

		if startDateStr == "" || endDateStr == "" {
			currentDate := time.Now()
			startDate = currentDate.AddDate(0, -5, 0) // 5 bulan ke belakang
			endDate = currentDate
		} else {
			if startDate, errDate = time.Parse("2006-01-02", startDateStr); errDate != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kesalahan format pada start date. Gunakan YYYY-MM-DD"})
			}

			if endDate, errDate = time.Parse("2006-01-02", endDateStr); errDate != nil {
				return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kesalahan format end date. Gunakan YYYY-MM-DD"})
			}

			if startDateStr != "" && endDateStr != "" {
				endDate = endDate.AddDate(0, 0, 0)
			}

			// Setelah menetapkan endDate, sekarang tentukan startDate
			if startDateStr == "" || endDateStr == "" {
				currentDate := time.Now()
				startDate = currentDate.AddDate(0, -5, 0) // 5 bulan ke belakang
			} else {
				if startDate, errDate = time.Parse("2006-01-02", startDateStr); errDate != nil {
					return c.JSON(http.StatusBadRequest, helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Kesalahan format pada start date. Gunakan YYYY-MM-DD"})
				}
			}

		}

		// Mengubah endDate untuk menyertakan waktu akhir hari
		endDate = endDate.AddDate(0, 0, 1).Add(-time.Second)

		var totalUserCount int64
		if err := db.Model(&model.User{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&totalUserCount).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung jumlah total user"})
		}

		var totalWisataCount int64
		if err := db.Model(&model.Wisata{}).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&totalWisataCount).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung tempat wisata yang tersedia"})
		}

		var totalVisitors int
		if err := db.Model(&model.Ticket{}).Where("paid_status = ? AND created_at BETWEEN ? AND ?", true, startDate, endDate).Select("COALESCE(SUM(quantity), 0) as total_visitors").Scan(&totalVisitors).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung pengunjung"})
		}

		var totalTicketPurchaseCount int64
		if err := db.Model(&model.Ticket{}).Where("paid_status = ? AND created_at BETWEEN ? AND ?", true, startDate, endDate).Count(&totalTicketPurchaseCount).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung order penjualan"})
		}

		var totalIncome int
		query := db.Model(&model.Ticket{}).
			Select("COALESCE(SUM(total_cost), 0) as total_income").
			Where("paid_status = ?", true).
			Where("created_at BETWEEN ? AND ?", startDate, endDate)

		if err := query.Scan(&totalIncome).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung total pendapatan"})
		}

		startMonth := time.Now().AddDate(0, -5, 0)
		currentDate := time.Now()
		monthlyIncome := make([]map[string]interface{}, 0)
		var totalSixMonthsIncome int

		for i := 0; i < 6; i++ {
			currentMonth := startMonth.AddDate(0, i, 0)
			startOfMonth := time.Date(currentMonth.Year(), currentMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

			if endOfMonth.After(currentDate) {
				endOfMonth = currentDate
			}

			var monthlyTotalIncome int
			monthlyQuery := db.Model(&model.Ticket{}).
				Select("COALESCE(SUM(total_cost), 0) as total_income").
				Where("paid_status = ?", true).
				Where("created_at BETWEEN ? AND ?", startOfMonth, endOfMonth)

			if err := monthlyQuery.Scan(&monthlyTotalIncome).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Gagal dalam menghitung pendapatan bulanan"})
			}

			monthlyIncome = append(monthlyIncome, map[string]interface{}{
				"month":  startOfMonth.Format("January"),
				"income": monthlyTotalIncome,
			})
			totalSixMonthsIncome += monthlyTotalIncome
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":                     http.StatusOK,
			"error":                    false,
			"totalUserCount":           totalUserCount,
			"totalWisataCount":         totalWisataCount,
			"totalVisitors":            totalVisitors,
			"totalTicketPurchaseCount": totalTicketPurchaseCount,
			"totalIncomeForTimeRange":  totalIncome,
			"timeRangeStart":           startDate.Format("2006-01-02"),
			"timeRangeEnd":             endDate.Format("2006-01-02"),
			"monthlyIncome":            monthlyIncome,
		})
	}
}

func GetTopEmition(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		type ResultQuery struct {
			Name         string  `json:"user_name"`
			Profile      string  `json:"user_profile"`
			Purchassed   int     `json:"purchassed"`
			TotalEmition float64 `json:"total_emition"`
		}
		var resQuery []ResultQuery

		err = db.Model(model.User{}).Joins("JOIN tickets ON users.id = tickets.user_id AND tickets.paid_status = ?", true).
			Group("users.name").Order("total_emition asc").Limit(4).
			Select("users.name AS name, MIN(users.photo_profil) AS profile, COUNT(tickets.id) AS purchassed, SUM(tickets.carbon_footprint) AS total_emition").
			Scan(&resQuery).Error

		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "data": []interface{}{}})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "data": resQuery})
	}
}

func GetTopWisata(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		type ResultQuery struct {
			Title        string `json:"destination_title"`
			TotalTicket  int64  `json:"total_ticket"`
			PhotoWisata1 string `json:"photo_wisata1"`
			PhotoWisata2 string `json:"photo_wisata2"`
			PhotoWisata3 string `json:"photo_wisata3"`
			Kota         string `json:"kota"`
		}
		var resQuery []ResultQuery

		var totalTicketsSold int64
		err = db.Model(&model.Ticket{}).Where("paid_status = ?", true).Count(&totalTicketsSold).Error
		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "data": []interface{}{}})
		}

		err = db.Model(&model.Wisata{}).Select("wisata.title AS title, wisata.kota, COUNT(tickets.id) AS total_ticket, wisata.photo_wisata1, wisata.photo_wisata2, wisata.photo_wisata3").
			Joins("JOIN tickets ON wisata.id = tickets.wisata_id").
			Where("tickets.paid_status = ?", true).
			Group("wisata.title, wisata.kota, wisata.photo_wisata1, wisata.photo_wisata2, wisata.photo_wisata3").
			Order("total_ticket desc").Limit(3).Scan(&resQuery).Error

		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusInternalServerError, "error": true, "data": []interface{}{}})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "data": resQuery})
	}
}
