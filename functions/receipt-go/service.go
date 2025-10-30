package main

import (
	"context"
	"fmt"
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
	// Validate input
	if len(imageData) == 0 {
		return &ExtractionResponse{
			Success: false,
			Error:   "image data is empty",
		}, fmt.Errorf("image data is empty")
	}

	// Extract receipt data using OpenAI
	receiptData, err := s.openAIClient.ExtractReceiptData(ctx, imageData)
	if err != nil {
		return &ExtractionResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to extract receipt data: %v", err),
		}, err
	}

	// Return successful response
	return &ExtractionResponse{
		Success: true,
		Data:    receiptData,
	}, nil
}

// ValidateReceiptData validates the extracted receipt data for completeness
func ValidateReceiptData(data *ReceiptData) []string {
	var errors []string

	if data.MerchantName == "" {
		errors = append(errors, "merchant name is missing")
	}

	if data.Total <= 0 {
		errors = append(errors, "total amount is invalid or missing")
	}

	if len(data.Items) == 0 {
		errors = append(errors, "no items found in receipt")
	}

	// Validate that subtotal + tax approximately equals total (allow small rounding differences)
	calculatedTotal := data.Subtotal + data.Tax
	diff := calculatedTotal - data.Total
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.05 { // Allow up to 5 cents difference for rounding
		errors = append(errors, fmt.Sprintf("total calculation mismatch: subtotal(%.2f) + tax(%.2f) != total(%.2f)",
			data.Subtotal, data.Tax, data.Total))
	}

	return errors
}
