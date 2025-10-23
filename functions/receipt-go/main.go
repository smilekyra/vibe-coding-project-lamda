package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// ReceiptResponse represents the API response structure
type ReceiptResponse struct {
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	Timestamp int64  `json:"timestamp"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

// Handler handles the Lambda function invocation
// Works with Lambda Function URLs
func Handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Only accept POST method
	if request.RequestContext.HTTP.Method != "POST" {
		errorResponse := ErrorResponse{
			Error: "Method not allowed. Only POST is supported.",
		}
		errorBytes, _ := json.Marshal(errorResponse)

		return events.LambdaFunctionURLResponse{
			StatusCode: 405,
			Headers: map[string]string{
				"Content-Type": "application/json",
				"Allow":        "POST",
			},
			Body: string(errorBytes),
		}, nil
	}

	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
		errorResponse := ErrorResponse{
			Error: "Invalid request body. Expected JSON format.",
		}
		errorBytes, _ := json.Marshal(errorResponse)

		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(errorBytes),
		}, nil
	}

	// Get file data (expect base64 encoded file)
	fileData, ok := requestBody["file"].(string)
	if !ok || fileData == "" {
		errorResponse := ErrorResponse{
			Error: "Missing or invalid 'file' field in request body.",
		}
		errorBytes, _ := json.Marshal(errorResponse)

		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(errorBytes),
		}, nil
	}

	// Get file name (optional)
	fileName := "unknown"
	if name, ok := requestBody["fileName"].(string); ok && name != "" {
		fileName = name
	}

	// Decode base64 to get actual file size
	decodedFile, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		errorResponse := ErrorResponse{
			Error: "Failed to decode file data. Expected base64 encoded string.",
		}
		errorBytes, _ := json.Marshal(errorResponse)

		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(errorBytes),
		}, nil
	}

	// Calculate file size
	fileSize := int64(len(decodedFile))

	// Create response
	receiptResponse := ReceiptResponse{
		FileName:  fileName,
		FileSize:  fileSize,
		Timestamp: time.Now().Unix(),
	}

	// Marshal to JSON
	responseBytes, err := json.Marshal(receiptResponse)
	if err != nil {
		errorResponse := ErrorResponse{
			Error: "Failed to generate response",
		}
		errorBytes, _ := json.Marshal(errorResponse)

		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(errorBytes),
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
