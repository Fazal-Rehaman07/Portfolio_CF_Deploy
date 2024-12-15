package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Define the structure for storing visitor logs
type VisitorLog struct {
	IP        string `json:"IP"`
	Timestamp string `json:"Timestamp"`
}

var dynamoClient *dynamodb.DynamoDB

func init() {
	// Initialize DynamoDB session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // Replace with your AWS region
	}))
	dynamoClient = dynamodb.New(sess)
}

func handler(ctx context.Context, event map[string]interface{}) (map[string]interface{}, error) {
	// Extract body from the event
	body, ok := event["body"].(string)
	if !ok {
		log.Printf("Error: no body found in the event")
		return nil, fmt.Errorf("invalid request body")
	}

	// Parse the incoming event body
	var visitor VisitorLog
	err := json.Unmarshal([]byte(body), &visitor)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return nil, fmt.Errorf("invalid request body")
	}

	// Add timestamp to the visitor log
	visitor.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	// Convert to DynamoDB item
	item, err := dynamodbattribute.MarshalMap(visitor)
	if err != nil {
		log.Printf("Failed to marshal visitor log: %v", err)
		return nil, fmt.Errorf("intenal server error")
	}

	// Store item in DynamoDB
	input := &dynamodb.PutItemInput{
		TableName: aws.String("VisitorLogs"), // Replace with your table name
		Item:      item,
	}
	_, err = dynamoClient.PutItem(input)
	if err != nil {
		log.Printf("Failed to put item in DynamoDB: %v", err)
		return nil, fmt.Errorf("intenal server error")
	}

	// Return a success message with CORS headers
	return map[string]interface{}{
		"statusCode": 200,
		"body":       "Visitor log stored successfully",
		"headers": map[string]string{
			"Access-Control-Allow-Origin":  "*",                  // Allows all domains. Replace with specific domains for tighter security
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS", // Allowed HTTP methods
			"Access-Control-Allow-Headers": "Content-Type",       // Allowed headers
		},
	}, nil
}

func main() {
	log.Println("Lambda Execution Started")
	// Start Lambda function
	lambda.Start(handler)
}
