package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// ReceiptResponse represents the API response structure
type ReceiptResponse struct {
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	S3Key     string `json:"s3Key"`
	S3Bucket  string `json:"s3Bucket"`
	Timestamp int64  `json:"timestamp"`
}

// ErrorResponse represents the error response structure
type ErrorResponse struct {
	Error string `json:"error"`
}

const (
	s3BucketName = "vibe-receipt-uploads-kyra"
)

// S3Uploader interface for uploading files to S3
type S3Uploader interface {
	Upload(ctx context.Context, fileData []byte, fileName string) (string, error)
}

// RealS3Uploader implements S3Uploader using AWS SDK
type RealS3Uploader struct{}

// Upload uploads file to S3 with date-based folder structure
func (u *RealS3Uploader) Upload(ctx context.Context, fileData []byte, fileName string) (string, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 client
	svc := s3.New(sess)

	// Get current date for folder structure (YYYY-MM-DD)
	now := time.Now()
	dateFolder := now.Format("2006-01-02")

	// Extract file extension
	ext := filepath.Ext(fileName)
	nameWithoutExt := strings.TrimSuffix(fileName, ext)

	// Generate unique filename to handle duplicates
	uniqueID := uuid.New().String()[:8]
	uniqueFileName := fmt.Sprintf("%s_%s%s", nameWithoutExt, uniqueID, ext)

	// Create S3 key with date-based folder structure
	s3Key := fmt.Sprintf("%s/%s", dateFolder, uniqueFileName)

	// Upload to S3
	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Key),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s3Key, nil
}

var uploader S3Uploader = &RealS3Uploader{}

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

	// Upload to S3
	s3Key, err := uploader.Upload(ctx, decodedFile, fileName)
	if err != nil {
		errorResponse := ErrorResponse{
			Error: fmt.Sprintf("Failed to upload file to S3: %v", err),
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

	// Create response
	receiptResponse := ReceiptResponse{
		FileName:  fileName,
		FileSize:  fileSize,
		S3Key:     s3Key,
		S3Bucket:  s3BucketName,
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
