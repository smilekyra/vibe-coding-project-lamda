package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// HelloResponse represents the API response structure
type HelloResponse struct {
	Message   string `json:"message"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

// Handler handles the Lambda function invocation
// Works with Lambda Function URLs
func Handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Only accept GET method
	if request.RequestContext.HTTP.Method != "GET" {
		return events.LambdaFunctionURLResponse{
			StatusCode: 405,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Allow":        "GET",
			},
			Body: `{"error":"Method not allowed. Only GET is supported."}`,
		}, nil
	}

	// Get name from query parameters
	name := "World"
	if request.QueryStringParameters != nil {
		if n, ok := request.QueryStringParameters["name"]; ok && n != "" {
			name = n
		}
	}

	// Create response
	helloResponse := HelloResponse{
		Message:   "Hello, " + name + "!",
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
	}

	// Marshal to JSON
	responseBytes, err := json.Marshal(helloResponse)
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Failed to generate response"}`,
		}, nil
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBytes),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
