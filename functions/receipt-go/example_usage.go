package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// Example usage of the receipt extraction service
func ExampleUsage() {
	// 1. Read the receipt image file
	imageData, err := os.ReadFile("path/to/receipt-image.jpg")
	if err != nil {
		fmt.Printf("Failed to read image file: %v\n", err)
		return
	}

	// 2. Create OpenAI client
	// Make sure to set OPENAI_API_KEY environment variable
	openAIClient := NewOpenAIClient("")

	// 3. Create receipt extraction service
	service := NewReceiptExtractionService(openAIClient)

	// 4. Extract receipt data from image
	ctx := context.Background()
	response, err := service.ExtractFromImage(ctx, imageData)
	if err != nil {
		fmt.Printf("Failed to extract receipt data: %v\n", err)
		return
	}

	// 5. Check if extraction was successful
	if !response.Success {
		fmt.Printf("Extraction failed: %s\n", response.Error)
		return
	}

	// 6. Validate the extracted data
	validationErrors := ValidateReceiptData(response.Data)
	if len(validationErrors) > 0 {
		fmt.Println("Validation warnings:")
		for _, err := range validationErrors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// 7. Use the extracted data
	fmt.Println("Receipt extracted successfully!")
	fmt.Printf("Merchant: %s\n", response.Data.MerchantName)
	fmt.Printf("Date: %s\n", response.Data.TransactionDate)
	fmt.Printf("Total: $%.2f\n", response.Data.Total)
	fmt.Printf("Number of items: %d\n", len(response.Data.Items))

	// 8. Print items
	fmt.Println("\nItems:")
	for i, item := range response.Data.Items {
		fmt.Printf("%d. %s x%d - $%.2f (total: $%.2f)\n",
			i+1, item.Name, item.Quantity, item.Price, item.Total)
	}

	// 9. Convert to JSON for storage or API response
	jsonData, err := json.MarshalIndent(response.Data, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal to JSON: %v\n", err)
		return
	}

	fmt.Println("\nJSON output:")
	fmt.Println(string(jsonData))
}

// IntegrateWithLambdaHandler shows how to integrate the service with existing Lambda handler
func IntegrateWithLambdaHandler() {
	// This is an example of how you would modify the existing Handler function
	// to include receipt extraction functionality

	/*
		// In your Lambda handler, after uploading to S3:

		// Create OpenAI client and service
		openAIClient := NewOpenAIClient("")
		extractionService := NewReceiptExtractionService(openAIClient)

		// Extract receipt data
		extractionResponse, err := extractionService.ExtractFromImage(ctx, decodedFile)
		if err != nil {
			// Log error but continue - extraction is optional
			fmt.Printf("Failed to extract receipt data: %v\n", err)
		}

		// Create enhanced response that includes both S3 info and extracted data
		type EnhancedReceiptResponse struct {
			FileName       string        `json:"fileName"`
			FileSize       int64         `json:"fileSize"`
			S3Key          string        `json:"s3Key"`
			S3Bucket       string        `json:"s3Bucket"`
			Timestamp      int64         `json:"timestamp"`
			ExtractedData  *ReceiptData  `json:"extractedData,omitempty"`
			ExtractionError string       `json:"extractionError,omitempty"`
		}

		enhancedResponse := EnhancedReceiptResponse{
			FileName:  fileName,
			FileSize:  fileSize,
			S3Key:     s3Key,
			S3Bucket:  s3BucketName,
			Timestamp: time.Now().Unix(),
		}

		if extractionResponse != nil && extractionResponse.Success {
			enhancedResponse.ExtractedData = extractionResponse.Data
		} else if extractionResponse != nil {
			enhancedResponse.ExtractionError = extractionResponse.Error
		}

		// Return enhanced response
		responseBytes, err := json.Marshal(enhancedResponse)
		// ... rest of handler code
	*/
}
