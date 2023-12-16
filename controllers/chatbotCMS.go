package controllers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
	"myproject/helper"
	"myproject/middleware"
	"myproject/model"
	"net/http"
	"os"
	"strings"
	"time"
)

type PromoChatbotResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

type PromoChatbotUsecase interface {
	RecommendPromo(userInput, openAIKey string, db *gorm.DB) (string, error)
}

type promoChatbotUsecase struct{}

func NewPromoChatbotUsecase() PromoChatbotUsecase {
	return &promoChatbotUsecase{}
}

func (uc *promoChatbotUsecase) RecommendPromo(userInput, openAIKey string, db *gorm.DB) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(openAIKey)
	model := openai.GPT3Dot5Turbo
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "Halo, perkenalkan saya sistem untuk rekomendasi promo"},
		{Role: openai.ChatMessageRoleUser, Content: userInput},
	}

	resp, err := uc.getCompletionFromMessages(ctx, client, messages, model)
	if err != nil {
		return "", err
	}
	answer := resp.Choices[0].Message.Content

	promoRecommendation, err := uc.getPromoRecommendation(db)
	if err != nil {
		return "", err
	}

	newPromoRecommendation, err := uc.generateNewPromoRecommendation(userInput, openAIKey)
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("%s\n\nPromo tiket wisata yang sudah ada:\n%s\n\nRekomendasi Promo Tiket Wisata Baru:\n%s", answer, promoRecommendation, newPromoRecommendation)

	return result, nil
}

func (uc *promoChatbotUsecase) generateNewPromoRecommendation(userInput, openAIKey string) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(openAIKey)
	model := openai.GPT3Dot5Turbo

	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: "Halo, perkenalkan saya sistem untuk rekomendasi promo tiket wisata baru"},
		{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("Rekomendasi promo tiket wisata baru berdasarkan input pengguna: %s", userInput)},
	}

	resp, err := uc.getCompletionFromMessages(ctx, client, messages, model)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (uc *promoChatbotUsecase) getPromoRecommendation(db *gorm.DB) (string, error) {
	var promos []model.Promo
	err := db.Where("status_aktif = ? AND tanggal_kadaluarsa > ?", true, time.Now()).Find(&promos).Error
	if err != nil {
		return "", err
	}

	var promoRecommendation string
	for _, promo := range promos {
		promoRecommendation += fmt.Sprintf("- %s: %s\n", promo.NamaPromo, promo.Deskripsi)
	}

	return promoRecommendation, nil
}

func (uc *promoChatbotUsecase) getCompletionFromMessages(
	ctx context.Context,
	client *openai.Client,
	messages []openai.ChatCompletionMessage,
	model string,
) (openai.ChatCompletionResponse, error) {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)
	return resp, err
}

func RecommendPromoChatbot(c echo.Context, promoChatbotUsecase PromoChatbotUsecase, db *gorm.DB, secretKey []byte) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Token autorisasi tidak ditemukan"})
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Kesalahan format token"})
	}

	tokenString = authParts[1]

	username, err := middleware.VerifyToken(tokenString, secretKey)
	if err != nil {
		errorResponse := helper.ErrorResponse{Code: http.StatusUnauthorized, Message: "Kesalahan token"}
		return c.JSON(http.StatusUnauthorized, errorResponse)
	}

	var adminUser model.User
	result := db.Where("username = ?", username).First(&adminUser)
	if result.Error != nil {
		errorResponse := helper.ErrorResponse{Code: http.StatusNotFound, Message: "Admin tidak ditemukan"}
		return c.JSON(http.StatusNotFound, errorResponse)
	}

	if !adminUser.IsAdmin {
		errorResponse := helper.ErrorResponse{Code: http.StatusForbidden, Message: "Akses tidak diberikan, anda bukan admin"}
		return c.JSON(http.StatusForbidden, errorResponse)
	}

	var requestData map[string]interface{}
	err = c.Bind(&requestData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Kesalahan format JSON"})
	}

	userInput, ok := requestData["message"].(string)
	if !ok || userInput == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Kesalahan atau tidak ada 'message' dalam request"})
	}

	if !containsPromoKeyword(userInput) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Pertanyaan tidak berkaitan dengan promo tiket wisata"})
	}

	userInput = fmt.Sprintf("Rekomendasi promo tiket wisata: %s", userInput)

	answer, err := promoChatbotUsecase.RecommendPromo(userInput, os.Getenv("OPENAI_API_KEY"), db)
	if err != nil {
		errorMessage := "Failed to generate promo recommendations"
		if strings.Contains(err.Error(), "rate limits exceeded") {
			errorMessage = "Rate limits exceeded. Please try again later."
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": true, "message": errorMessage})
	}

	responseData := PromoChatbotResponse{
		Status: "success",
		Data:   answer,
	}

	return c.JSON(http.StatusOK, responseData)
}

func containsPromoKeyword(input string) bool {
	keywords := []string{"promo", "diskon", "potongan", "voucher", "kupon", "kode promo", "kode voucher", "kode kupon", "kode diskon", "potongan harga", "potongan tiket", "potongan harga tiket", "potongan tiket wisata"}
	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(input), keyword) {
			return true
		}
	}
	return false
}
