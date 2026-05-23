package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type URLRequest struct {
	URL string `json:"url"`
}

type URLResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

func generateShortCode() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Lambda iniciada SIMPLE")

	var urlReq URLRequest
	if err := json.Unmarshal([]byte(req.Body), &urlReq); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid JSON"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	if urlReq.URL == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "URL is required"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	shortCode := generateShortCode()

	response := URLResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("http://localhost:4566/%s", shortCode),
		OriginalURL: urlReq.URL,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}

	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
