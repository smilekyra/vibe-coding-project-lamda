package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// TimeResponse represents the API response structure
type TimeResponse struct {
	Message     string `json:"message"`
	CurrentTime string `json:"current_time"`
	Timezone    string `json:"timezone"`
	Timestamp   int64  `json:"timestamp"`
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

	// Load JST timezone
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"Failed to load timezone"}`,
		}, nil
	}

	// Get current time in JST
	currentTime := time.Now().In(jst)

	// Create response
	timeResponse := TimeResponse{
		Message:     "Current time in Japan",
		CurrentTime: currentTime.Format("2006-01-02 15:04:05"),
		Timezone:    "JST (Asia/Tokyo)",
		Timestamp:   currentTime.Unix(),
	}

	// Marshal to JSON
	responseBytes, err := json.Marshal(timeResponse)
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
