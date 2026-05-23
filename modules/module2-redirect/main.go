package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type URLItem struct {
	ShortCode   string `json:"ShortCode"`
	OriginalURL string `json:"OriginalURL"`
	CreatedAt   string `json:"CreatedAt"`
	Visits      int    `json:"Visits"`
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

func updateVisits(shortCode string) {
	updateExpr := "ADD Visits :inc"
	_, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"ShortCode": {S: aws.String(shortCode)},
		},
		UpdateExpression: aws.String(updateExpr),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":inc": {N: aws.String("1")},
		},
	})
	if err != nil {
		fmt.Printf("Error updating visits: %v\n", err)
	}
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	shortCode := req.PathParameters["code"]

	if shortCode == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Code is required"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
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
		}, nil
	}

	if result.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "URL not found"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	var item URLItem
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Parse error"}`,
		}, nil
	}

	go updateVisits(shortCode)

	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location": item.OriginalURL,
		},
	}, nil
}

func main() {
	lambda.Start(handler)
}
