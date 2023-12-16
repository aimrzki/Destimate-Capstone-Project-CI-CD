package controllers

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"os"
	"strings"
)

type WisataResponse struct {
	Status string `json:"status"`
	Data   string `json:"data"`
}

type WisataUsecase interface {
	RecommendWisata(userInput, openAIKey string) (string, error)
}

type wisataUsecase struct{}

func NewWisataUsecase() WisataUsecase {
	return &wisataUsecase{}
}

func (uc *wisataUsecase) RecommendWisata(userInput, openAIKey string) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(openAIKey)
	model := openai.GPT3Dot5Turbo
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Halo, perkenalkan saya sistem untuk rekomendasi tempat wisata",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userInput,
		},
	}

	resp, err := uc.getCompletionFromMessages(ctx, client, messages, model)
	if err != nil {
		return "", err
	}
	answer := resp.Choices[0].Message.Content
	return answer, nil
}

func (uc *wisataUsecase) getCompletionFromMessages(
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

func RecommendWisata(c echo.Context, wisataUsecase WisataUsecase) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Authorization token is missing"})
	}

	authParts := strings.SplitN(tokenString, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": true, "message": "Invalid token format"})
	}

	tokenString = authParts[1]

	var requestData map[string]interface{}
	err := c.Bind(&requestData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid JSON format"})
	}

	userInput, ok := requestData["message"].(string)
	if !ok || userInput == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Invalid or missing 'message' in the request"})
	}

	// Check if the user input contains keywords related to tourism
	tourismKeywords := []string{"wisata", "tempat wisata", "rekomendasi wisata", "tourism", "tourist", "tourist attraction", "tourist spot", "tourist destination", "travel", "traveling", "traveler", "traveler attraction", "traveler spot", "traveler destination", "traveling destination", "traveling spot", "traveling attraction", "travel"}
	containsTourismKeyword := false
	for _, keyword := range tourismKeywords {
		if strings.Contains(strings.ToLower(userInput), keyword) {
			containsTourismKeyword = true
			break
		}
	}

	if !containsTourismKeyword {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": true, "message": "Sorry, this chatbot only serves questions about tempat wisata & rekomendasi wisata"})
	}

	userInput = fmt.Sprintf("Rekomendasi wisata: %s", userInput)

	answer, err := wisataUsecase.RecommendWisata(userInput, os.Getenv("OPENAI_API_KEY"))
	if err != nil {
		errorMessage := "Failed to generate wisata recommendations"
		if strings.Contains(err.Error(), "rate limits exceeded") {
			errorMessage = "Rate limits exceeded. Please try again later."
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": true, "message": errorMessage})
	}

	responseData := WisataResponse{
		Status: "success",
		Data:   answer,
	}

	return c.JSON(http.StatusOK, responseData)
}
