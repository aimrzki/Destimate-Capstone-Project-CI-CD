package controllers

import (
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
)

func CreateCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		var category model.Category
		if err := c.Bind(&category); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(category.CategoryName) < 5 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Category name must be at least 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(category.CategoryName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Category name maximum at least 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		existingCategory := model.Category{}
		if err := db.Where("category_name = ?", category.CategoryName).First(&existingCategory).Error; err == nil {
			// Category dengan nama tersebut sudah ada
			errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Category with this name already exists"}
			return c.JSON(http.StatusConflict, errorResponse)
		}

		db.Create(&category)

		successResponse := helper.Response{Code: http.StatusCreated, Error: false, Message: "Category created successfully", Category: &category}
		return c.JSON(http.StatusCreated, successResponse)
	}
}

func UpdateCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		categoryID, err := helper.ConvertParamToUint(c.Param("id"))
		if err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingCategory model.Category
		result := db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.Response{Code: http.StatusNotFound, Error: true, Message: "Category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		var updatedCategory model.Category
		if err := c.Bind(&updatedCategory); err != nil {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Invalid request body"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(updatedCategory.CategoryName) < 5 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Category name must be at least 5 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if len(updatedCategory.CategoryName) > 30 {
			errorResponse := helper.Response{Code: http.StatusBadRequest, Error: true, Message: "Category maximum at least 30 characters"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		if updatedCategory.CategoryName != existingCategory.CategoryName {
			var categoryWithSameName model.Category
			if err := db.Where("category_name = ?", updatedCategory.CategoryName).First(&categoryWithSameName).Error; err == nil {
				// Category dengan nama tersebut sudah ada
				errorResponse := helper.Response{Code: http.StatusConflict, Error: true, Message: "Category with this name already exists"}
				return c.JSON(http.StatusConflict, errorResponse)
			}
		}

		// Update only the specified fields
		db.Model(&existingCategory).Updates(updatedCategory)

		successResponse := helper.Response{Code: http.StatusOK, Error: false, Message: "Category updated successfully", Category: &existingCategory}
		return c.JSON(http.StatusOK, successResponse)
	}
}

func DeleteCategoryByAdmin(db *gorm.DB, secretKey []byte) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := middleware.AuthenticateAndAuthorize(c, db, secretKey)
		if err != nil {
			return err
		}

		categoryID, err := helper.ConvertParamToUint(c.Param("id"))
		if err != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusBadRequest, Message: "Invalid category ID"}
			return c.JSON(http.StatusBadRequest, errorResponse)
		}

		var existingCategory model.Category
		result := db.First(&existingCategory, categoryID)
		if result.Error != nil {
			errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Category not found"}
			return c.JSON(http.StatusNotFound, errorResponse)
		}

		db.Delete(&existingCategory)

		return c.JSON(http.StatusOK, map[string]interface{}{"code": http.StatusOK, "error": false, "message": "Category deleted successfully"})
	}
}
