package helper

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

func decodeBase64Credential() ([]byte, error) {
	credentialsBase64 := os.Getenv("CREDENTIALS")
	credentialsBytes, err := base64.StdEncoding.DecodeString(credentialsBase64)
	if err != nil {
		fmt.Println("Error decoding base64 credential")
		return nil, err
	}
	return credentialsBytes, nil
}

func UploadImageToGCS(imageData []byte, imageName string) (string, error) {
	ctx := context.Background()

	credentialsBytes, err := decodeBase64Credential()
	if err != nil {
		return "", err
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentialsBytes))
	if err != nil {
		return "", err
	}

	bucketName := "destimate"

	object := client.Bucket(bucketName).Object(imageName)
	wc := object.NewWriter(ctx)

	wc.ContentType = "image/jpeg"

	if _, err := io.Copy(wc, bytes.NewReader(imageData)); err != nil {
		wc.Close()
		return "", err
	}

	if err := wc.Close(); err != nil {
		return "", err
	}

	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, imageName)
	return imageURL, nil
}

func IsImageFile(file *multipart.FileHeader) bool {
	return strings.HasPrefix(file.Header.Get("Content-Type"), "image/")
}

func IsFileSizeExceeds(file *multipart.FileHeader, maxSize int64) bool {
	return file.Size > maxSize
}
