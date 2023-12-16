// In controllers/getUserNotification.go

package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"strings"
	"time"
)

type NotificationResponse struct {
	ID            uint       `json:"id"`
	UserID        uint       `json:"user_id"`
	Label         string     `json:"label"`
	Message       string     `json:"message"`
	Status        string     `json:"status"`
	Title         string     `json:"title"`
	InvoiceNumber string     `json:"invoice_number"`
	CreatedAt     *time.Time `json:"created_at"`
	IsRead        bool       `json:"is_read"`
}

type PromoNotificationResponse struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Label     string     `json:"label"`
	Message   string     `json:"message"`
	PromoID   uint       `json:"promo_id"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"created_at"`
	IsRead    bool       `json:"is_read"`
}

func GetUserNotifications(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := middleware.ExtractUsernameFromToken(c, secretKey)

		var user model.User
		result := db.Where("username = ?", username).First(&user)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusInternalServerError, Message: "Failed to fetch user data"}
			return c.JSON(http.StatusInternalServerError, errorResponse)
		}

		var notifications []model.Notification
		db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&notifications)

		var notificationResponses []interface{}

		for _, notification := range notifications {
			isRead := notification.Status == "read"
			if !isRead {
				notification.Status = "unread"
			}

			if strings.HasPrefix(notification.Title, "promo ") {
				promoTitle := strings.TrimPrefix(notification.Title, "promo ")
				promoNotificationResponse := PromoNotificationResponse{
					ID:        notification.ID,
					Label:     "Promo",
					Title:     "Promo " + promoTitle,
					Message:   notification.Message,
					PromoID:   notification.PromoID,
					Status:    notification.Status,
					CreatedAt: notification.CreatedAt,
					IsRead:    isRead,
				}
				notificationResponses = append(notificationResponses, promoNotificationResponse)
			} else {
				notificationResponse := NotificationResponse{
					ID:            notification.ID,
					UserID:        notification.UserID,
					Label:         "Pembayaran",
					Message:       notification.Message,
					Status:        notification.Status,
					Title:         notification.Title,
					InvoiceNumber: notification.InvoiceNumber,
					CreatedAt:     notification.CreatedAt,
					IsRead:        isRead,
				}
				notificationResponses = append(notificationResponses, notificationResponse)
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":          http.StatusOK,
			"error":         false,
			"message":       "User notifications retrieved successfully",
			"notifications": notificationResponses,
		})
	}
}

func MarkNotificationAsRead(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		notificationID := c.Param("id")

		var notification model.Notification
		result := db.First(&notification, notificationID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Notification not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		notification.Status = "read"
		db.Save(&notification)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":    http.StatusOK,
			"error":   false,
			"message": "Notification marked as read successfully",
		})
	}
}
