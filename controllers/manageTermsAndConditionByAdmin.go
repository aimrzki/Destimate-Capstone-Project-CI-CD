package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func CreateTermCondition(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var term model.TermCondition
		if err := c.Bind(&term); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validasi panjang karakter input
		if len(term.Name) < 5 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name should be at least 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(term.Name) > 100 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name maximum at least 100 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(term.Description) < 10 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description should be at least 10 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(term.Description) > 2000 {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description maximum at least 2000 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Check if tnc_name already exists
		var existingTerm model.TermCondition
		if err := db.Where("name = ?", term.Name).First(&existingTerm).Error; err == nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Term and condition with the same name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		// Create the new term and condition
		db.Create(&term)

		return c.JSON(http.StatusCreated, map[string]interface{}{"code": http.StatusCreated, "error": false, "term_condition": term})
	}
}

func EditTermCondition(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		tncID := c.Param("id")

		term := model.TermCondition{}
		result := db.First(&term, tncID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Term and condition not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var req model.TermCondition
		if err := c.Bind(&req); err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		// Validation for Name
		if req.Name != "" {
			if len(req.Name) < 5 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name should be at least 5 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(req.Name) > 100 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Name maximum at least 100 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			// Check if the new name already exists
			var existingTerm model.TermCondition
			if err := db.Where("name = ? AND id != ?", req.Name, tncID).First(&existingTerm).Error; err == nil {
				errorResponse := helper.ErrorResponse{Code: http.StatusConflict, Message: "Term and condition with the same name already exists"}
				return c.JSON(http.StatusConflict, errorResponse)
			}

			term.Name = req.Name
		}

		// Validation for Description
		if req.Description != "" {
			if len(req.Description) < 10 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description should be at least 10 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			if len(req.Description) > 2000 {
				errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Description maximum at least 2000 characters"}
				return c.JSON(http.StatusBadRequest, errorResponse)
			}

			term.Description = req.Description
		}

		db.Save(&term)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Term and condition updated successfully"})
	}
}

func DeleteTermCondition(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		tncID := c.Param("id")

		term := model.TermCondition{}
		result := db.First(&term, tncID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Term and condition not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&term)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Term and condition deleted successfully"})
	}
}
