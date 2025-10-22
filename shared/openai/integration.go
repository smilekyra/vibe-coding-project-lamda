package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ProcessReceiptFromS3URL processes a receipt image from an S3 URL
// This is a convenience function for AWS Lambda integration
func (s *Service) ProcessReceiptFromS3URL(ctx context.Context, s3URL string) (*ReceiptData, error) {
	response, err := s.ExtractReceiptDataFromURL(ctx, s3URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract receipt: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("extraction failed: %s", response.Error)
	}

	return response.Data, nil
}

// ProcessReceiptFromBase64 processes a receipt image from base64 encoded data
// This is useful when receiving base64 data from API requests
func (s *Service) ProcessReceiptFromBase64(ctx context.Context, base64Data string) (*ReceiptData, error) {
	response, err := s.ExtractReceiptDataFromBase64(ctx, base64Data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to extract receipt: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("extraction failed: %s", response.Error)
	}

	return response.Data, nil
}

// ProcessReceiptWithContext processes a receipt with additional context hints
func (s *Service) ProcessReceiptWithContext(ctx context.Context, imageSource string, hints map[string]string) (*ReceiptData, error) {
	var response *ReceiptExtractionResponse
	var err error

	// Determine if imageSource is URL or base64
	if len(imageSource) > 8 && (imageSource[:7] == "http://" || imageSource[:8] == "https://") {
		response, err = s.ExtractReceiptDataFromURL(ctx, imageSource, hints)
	} else {
		response, err = s.ExtractReceiptDataFromBase64(ctx, imageSource, hints)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to extract receipt: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("extraction failed: %s", response.Error)
	}

	return response.Data, nil
}

// DownloadAndProcessReceipt downloads an image from a URL and processes it
func (s *Service) DownloadAndProcessReceipt(ctx context.Context, imageURL string) (*ReceiptData, error) {
	// Download the image
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	// Encode to base64 and process
	base64Image := EncodeImageToBase64(imageBytes)
	return s.ProcessReceiptFromBase64(ctx, base64Image)
}

// ToJSON converts ReceiptData to JSON string
func (r *ReceiptData) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// ToPrettyJSON converts ReceiptData to pretty-printed JSON string
func (r *ReceiptData) ToPrettyJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// FromJSON creates a ReceiptData from JSON string
func FromJSON(jsonStr string) (*ReceiptData, error) {
	var data ReceiptData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &data, nil
}

// Validate checks if the receipt data has required fields
func (r *ReceiptData) Validate() error {
	if r.StoreName == "" {
		return fmt.Errorf("store name is required")
	}
	if r.TotalAmount <= 0 {
		return fmt.Errorf("total amount must be positive")
	}
	if r.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if r.ReceiptDate.IsZero() {
		return fmt.Errorf("receipt date is required")
	}
	return nil
}

// GetTotalWithoutTax calculates the total amount without tax
func (r *ReceiptData) GetTotalWithoutTax() float64 {
	if r.SubtotalAmount > 0 {
		return r.SubtotalAmount
	}
	// If subtotal is not available, calculate from total - tax
	if r.TaxAmount > 0 {
		return r.TotalAmount - r.TaxAmount
	}
	return r.TotalAmount
}

// GetItemCount returns the total number of items
func (r *ReceiptData) GetItemCount() int {
	return len(r.Items)
}

// GetTotalQuantity returns the sum of all item quantities
func (r *ReceiptData) GetTotalQuantity() float64 {
	total := 0.0
	for _, item := range r.Items {
		total += item.Quantity
	}
	return total
}

// GetItemsByCategory groups items by category
func (r *ReceiptData) GetItemsByCategory() map[string][]ReceiptItem {
	categories := make(map[string][]ReceiptItem)
	for _, item := range r.Items {
		category := item.Category
		if category == "" {
			category = "Uncategorized"
		}
		categories[category] = append(categories[category], item)
	}
	return categories
}

// Summary returns a brief summary of the receipt
func (r *ReceiptData) Summary() string {
	return fmt.Sprintf("%s | %s | %d items | Total: %.2f %s",
		r.StoreName,
		r.ReceiptDate.Format("2006-01-02"),
		len(r.Items),
		r.TotalAmount,
		r.Currency,
	)
}
