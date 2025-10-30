package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

var (
	sheetsRepository SheetsRepository
	extractionService ReceiptExtractionService
)

// S3Uploader interface for uploading files to S3
type S3Uploader interface {
	Upload(ctx context.Context, fileData []byte, fileName string) (string, error)
}

// RealS3Uploader implements S3Uploader using AWS SDK
type RealS3Uploader struct{}

// Upload uploads file to S3 with date-based folder structure
func (u *RealS3Uploader) Upload(ctx context.Context, fileData []byte, fileName string) (string, error) {
	log.Printf("[INFO] Starting S3 upload - fileName: %s, size: %d bytes", fileName, len(fileData))

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Printf("[ERROR] Failed to create AWS session: %v", err)
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

	log.Printf("[INFO] Uploading to S3 - bucket: %s, key: %s", s3BucketName, s3Key)

	// Upload to S3
	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Key),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		log.Printf("[ERROR] S3 upload failed - bucket: %s, key: %s, error: %v", s3BucketName, s3Key, err)
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	log.Printf("[INFO] S3 upload successful - bucket: %s, key: %s", s3BucketName, s3Key)
	return s3Key, nil
}

var uploader S3Uploader = &RealS3Uploader{}

// initServices initializes Google Sheets repository and extraction service
func initServices(ctx context.Context) error {
	// Initialize Google Sheets repository if credentials are available
	credentialsJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	spreadsheetID := os.Getenv("GOOGLE_SPREADSHEET_ID")

	if credentialsJSON != "" && spreadsheetID != "" {
		log.Printf("[INFO] Initializing Google Sheets repository")
		repo, err := NewGoogleSheetsRepository(ctx, []byte(credentialsJSON), spreadsheetID)
		if err != nil {
			log.Printf("[ERROR] Failed to initialize Google Sheets repository: %v", err)
			return fmt.Errorf("failed to initialize Google Sheets repository: %w", err)
		}
		sheetsRepository = repo
		log.Printf("[INFO] Google Sheets repository initialized successfully")
	} else {
		log.Printf("[WARN] Google Sheets credentials not found, sheets integration disabled")
	}

	// Initialize OpenAI client and extraction service
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		log.Printf("[INFO] Initializing OpenAI extraction service")
		openAIClient := NewOpenAIClient(apiKey)
		extractionService = NewReceiptExtractionService(openAIClient)
		log.Printf("[INFO] OpenAI extraction service initialized successfully")
	} else {
		log.Printf("[WARN] OpenAI API key not found, extraction service disabled")
	}

	return nil
}

// Handler handles the Lambda function invocation
// Works with Lambda Function URLs
func Handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	log.Printf("[INFO] Received request - method: %s, path: %s", request.RequestContext.HTTP.Method, request.RequestContext.HTTP.Path)

	// Only accept POST method
	if request.RequestContext.HTTP.Method != "POST" {
		log.Printf("[WARN] Invalid HTTP method: %s, only POST is allowed", request.RequestContext.HTTP.Method)
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
		log.Printf("[ERROR] Failed to parse request body: %v", err)
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
		log.Printf("[ERROR] Missing or invalid 'file' field in request body")
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
	log.Printf("[INFO] Processing file - fileName: %s", fileName)

	// Decode base64 to get actual file size
	decodedFile, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		log.Printf("[ERROR] Failed to decode base64 file data: %v", err)
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
	log.Printf("[INFO] Decoded file - fileName: %s, size: %d bytes", fileName, fileSize)

	// Upload to S3
	s3Key, err := uploader.Upload(ctx, decodedFile, fileName)
	if err != nil {
		log.Printf("[ERROR] S3 upload failed - fileName: %s, error: %v", fileName, err)
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

	// Generate S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3BucketName, s3Key)

	// Extract receipt data if extraction service is available
	var receiptData *ReceiptData
	if extractionService != nil {
		log.Printf("[INFO] Extracting receipt data from image")
		extractionResp, err := extractionService.ExtractFromImage(ctx, decodedFile)
		if err != nil {
			log.Printf("[WARN] Failed to extract receipt data: %v (continuing without extraction)", err)
		} else if extractionResp.Success {
			receiptData = extractionResp.Data
			log.Printf("[INFO] Receipt extraction successful")

			// Save to Google Sheets if repository is available
			if sheetsRepository != nil && receiptData != nil {
				log.Printf("[INFO] Saving receipt data to Google Sheets")
				if err := sheetsRepository.SaveReceipt(ctx, receiptData, s3URL); err != nil {
					log.Printf("[ERROR] Failed to save to Google Sheets: %v (continuing)", err)
				} else {
					log.Printf("[INFO] Successfully saved receipt to Google Sheets")
				}
			}
		}
	}

	// Create response
	receiptResponse := ReceiptResponse{
		FileName:  fileName,
		FileSize:  fileSize,
		S3Key:     s3Key,
		S3Bucket:  s3BucketName,
		Timestamp: time.Now().Unix(),
	}

	log.Printf("[INFO] Request successful - fileName: %s, s3Key: %s, size: %d bytes", fileName, s3Key, fileSize)

	// Marshal to JSON
	responseBytes, err := json.Marshal(receiptResponse)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal response to JSON: %v", err)
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
	// Initialize services
	ctx := context.Background()
	if err := initServices(ctx); err != nil {
		log.Fatalf("[FATAL] Failed to initialize services: %v", err)
	}

	lambda.Start(Handler)
}
