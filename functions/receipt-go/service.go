package main

import (
	"context"
	"fmt"
	"log"
)

// ReceiptExtractionService interface defines the receipt extraction operations
type ReceiptExtractionService interface {
	ExtractFromImage(ctx context.Context, imageData []byte) (*ExtractionResponse, error)
}

// RealReceiptExtractionService implements ReceiptExtractionService
type RealReceiptExtractionService struct {
	openAIClient OpenAIClient
}

// NewReceiptExtractionService creates a new receipt extraction service
func NewReceiptExtractionService(openAIClient OpenAIClient) *RealReceiptExtractionService {
	return &RealReceiptExtractionService{
		openAIClient: openAIClient,
	}
}

// ExtractFromImage extracts receipt data from an image
func (s *RealReceiptExtractionService) ExtractFromImage(ctx context.Context, imageData []byte) (*ExtractionResponse, error) {
	log.Printf("[INFO] ExtractFromImage called - imageSize: %d bytes", len(imageData))

	// Validate input
	if len(imageData) == 0 {
		log.Printf("[ERROR] Image data is empty")
		return &ExtractionResponse{
			Success: false,
			Error:   "image data is empty",
		}, fmt.Errorf("image data is empty")
	}

	// Extract receipt data using OpenAI
	receiptData, err := s.openAIClient.ExtractReceiptData(ctx, imageData)
	if err != nil {
		log.Printf("[ERROR] Failed to extract receipt data: %v", err)
		return &ExtractionResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to extract receipt data: %v", err),
		}, err
	}

	log.Printf("[INFO] Receipt extraction successful - merchant: %s, items: %d", receiptData.MerchantName, len(receiptData.Items))

	// Return successful response
	return &ExtractionResponse{
		Success: true,
		Data:    receiptData,
	}, nil
}

// ValidateReceiptData validates the extracted receipt data for completeness
func ValidateReceiptData(data *ReceiptData) []string {
	log.Printf("[INFO] Validating receipt data - merchant: %s, total: %.2f, items: %d",
		data.MerchantName, data.Total, len(data.Items))

	var errors []string

	if data.MerchantName == "" {
		log.Printf("[WARN] Validation failed: merchant name is missing")
		errors = append(errors, "merchant name is missing")
	}

	if data.Total <= 0 {
		log.Printf("[WARN] Validation failed: total amount is invalid or missing (total: %.2f)", data.Total)
		errors = append(errors, "total amount is invalid or missing")
	}

	if len(data.Items) == 0 {
		log.Printf("[WARN] Validation failed: no items found in receipt")
		errors = append(errors, "no items found in receipt")
	}

	// Validate that subtotal + tax approximately equals total (allow small rounding differences)
	calculatedTotal := data.Subtotal + data.Tax
	diff := calculatedTotal - data.Total
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.05 { // Allow up to 5 cents difference for rounding
		log.Printf("[WARN] Validation failed: total calculation mismatch - subtotal: %.2f, tax: %.2f, total: %.2f, diff: %.2f",
			data.Subtotal, data.Tax, data.Total, diff)
		errors = append(errors, fmt.Sprintf("total calculation mismatch: subtotal(%.2f) + tax(%.2f) != total(%.2f)",
			data.Subtotal, data.Tax, data.Total))
	}

	if len(errors) == 0 {
		log.Printf("[INFO] Receipt data validation successful")
	} else {
		log.Printf("[WARN] Receipt data validation completed with %d errors", len(errors))
	}

	return errors
}
