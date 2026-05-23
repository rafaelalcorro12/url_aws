package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type StatsResponse struct {
	ShortCode   string         `json:"short_code"`
	OriginalURL string         `json:"original_url"`
	TotalVisits int            `json:"total_visits"`
	CreatedAt   string         `json:"created_at"`
	DailyStats  map[string]int `json:"daily_stats"`
}

type URLItem struct {
	ShortCode   string         `json:"ShortCode"`
	OriginalURL string         `json:"OriginalURL"`
	CreatedAt   string         `json:"CreatedAt"`
	Visits      int            `json:"Visits"`
	DailyStats  map[string]int `json:"DailyStats"`
}

var db *dynamodb.DynamoDB
var tableName = "ShortUrls"

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("http://172.17.0.1:4566"),
	}))
	db = dynamodb.New(sess)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	shortCode := req.PathParameters["code"]

	if shortCode == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Code is required"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	result, err := db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ShortCode": {S: aws.String(shortCode)},
		},
	})

	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Database error"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	if result.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "URL not found"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	var item URLItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Parse error"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	response := StatsResponse{
		ShortCode:   item.ShortCode,
		OriginalURL: item.OriginalURL,
		TotalVisits: item.Visits,
		CreatedAt:   item.CreatedAt,
		DailyStats:  item.DailyStats,
	}

	if response.DailyStats == nil {
		response.DailyStats = make(map[string]int)
	}

	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(responseBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func main() {
	lambda.Start(handler)
}
