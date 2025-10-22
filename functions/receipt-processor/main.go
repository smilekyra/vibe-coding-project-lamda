package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"vibe-coding-project-lambda/functions/receipt-processor/handler"
	"vibe-coding-project-lambda/functions/receipt-processor/service"
	"vibe-coding-project-lambda/shared/openai"
	"vibe-coding-project-lambda/shared/repository"
)

const (
	defaultBucketName = "lambda-file-uploads"
	defaultRegion     = "ap-northeast-1"
	defaultSheetName  = "가계부" // Default sheet name for household ledger
)

var receiptHandler *handler.ReceiptHandler

// init initializes all dependencies
func init() {
	ctx := context.Background()

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(defaultRegion))
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	// Initialize S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Get bucket name from environment variable or use default
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = defaultBucketName
	}

	// Create repository layer
	s3Repo := repository.NewS3Repository(s3Client, bucketName, defaultRegion)

	// Create OpenAI service (optional - gracefully handle if API key is missing)
	var openaiService *openai.Service
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		openaiService, err = openai.NewService(openai.ServiceConfig{
			APIKey:          apiKey,
			DefaultCurrency: "JPY",
			DefaultLanguage: "ja",
			DefaultTimezone: "Asia/Tokyo",
			VisionModel:     "gpt-4o",
			MaxTokens:       4096,
			Temperature:     0.1,
		})
		if err != nil {
			log.Printf("Warning: Failed to initialize OpenAI service: %v", err)
			openaiService = nil
		} else {
			log.Printf("OpenAI service initialized successfully")
		}
	} else {
		log.Printf("Warning: OPENAI_API_KEY not set, receipt OCR will be disabled")
	}

	// Create Google Sheets service (optional - gracefully handle if credentials are missing)
	var sheetsService *service.SheetsService
	serviceAccountJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON")
	spreadsheetID := os.Getenv("GOOGLE_SPREADSHEET_ID")

	if serviceAccountJSON != "" && spreadsheetID != "" {
		log.Printf("Initializing Google Sheets integration...")

		// Parse service account JSON
		jsonBytes, err := repository.ParseServiceAccountJSON(serviceAccountJSON)
		if err != nil {
			log.Printf("Warning: Failed to parse service account JSON: %v", err)
		} else {
			// Create Sheets repository
			sheetsRepo, err := repository.NewSheetsRepository(ctx, repository.SheetsConfig{
				ServiceAccountJSON: jsonBytes,
				SpreadsheetID:      spreadsheetID,
			})
			if err != nil {
				log.Printf("Warning: Failed to initialize Google Sheets repository: %v", err)
			} else {
				// Create Sheets service
				sheetsService = service.NewSheetsService(service.SheetsServiceConfig{
					SheetsRepo: sheetsRepo,
					SheetName:  defaultSheetName,
				})

				// Initialize spreadsheet with headers if needed
				if err := sheetsService.InitializeSpreadsheet(ctx); err != nil {
					log.Printf("Warning: Failed to initialize spreadsheet headers: %v", err)
				} else {
					log.Printf("Google Sheets service initialized successfully (Sheet: %s)", defaultSheetName)
				}
			}
		}
	} else {
		log.Printf("Warning: Google Sheets credentials not configured, spreadsheet integration disabled")
		if serviceAccountJSON == "" {
			log.Printf("  - GOOGLE_SERVICE_ACCOUNT_JSON is not set")
		}
		if spreadsheetID == "" {
			log.Printf("  - GOOGLE_SPREADSHEET_ID is not set")
		}
	}

	// Create service layer
	receiptService := service.NewReceiptService(s3Repo, openaiService)

	// Create handler layer with optional sheets service
	receiptHandler = handler.NewReceiptHandler(receiptService)

	// Set sheets service if available
	if sheetsService != nil {
		receiptHandler.SetSheetsService(sheetsService)
	}
}

func main() {
	lambda.Start(receiptHandler.Handle)
}
