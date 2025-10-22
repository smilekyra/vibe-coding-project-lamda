package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAI API structures
type openAIChatRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIChatMessage   `json:"messages"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
	Temperature    float32               `json:"temperature,omitempty"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIChatMessage struct {
	Role    string                 `json:"role"`
	Content []openAIMessageContent `json:"content"`
}

type openAIMessageContent struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openAIImageURL `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ExtractReceiptData extracts structured data from a receipt image
func (s *Service) ExtractReceiptData(ctx context.Context, req ReceiptExtractionRequest) (*ReceiptExtractionResponse, error) {
	// Validate request
	if req.ImageData == "" && req.ImageURL == "" {
		return &ReceiptExtractionResponse{
			Success: false,
			Error:   "either image_data or image_url must be provided",
		}, fmt.Errorf("no image data provided")
	}

	// Prepare image URL for API
	imageURL := req.ImageURL
	if imageURL == "" && req.ImageData != "" {
		// Check if it's already base64 encoded with data URI
		if len(req.ImageData) > 5 && req.ImageData[:5] == "data:" {
			imageURL = req.ImageData
		} else {
			// Detect image format from base64 data or use default
			mimeType := detectImageMimeType(req.ImageData)
			imageURL = fmt.Sprintf("data:%s;base64,%s", mimeType, req.ImageData)
		}
	}

	// Build the extraction prompt
	prompt := s.buildReceiptExtractionPrompt(req)

	// Call OpenAI Vision API
	receiptData, rawText, err := s.callVisionAPI(ctx, imageURL, prompt)
	if err != nil {
		return &ReceiptExtractionResponse{
			Success: false,
			Error:   err.Error(),
			RawText: rawText,
		}, err
	}

	receiptData.RawText = rawText

	return &ReceiptExtractionResponse{
		Success: true,
		Data:    receiptData,
		RawText: rawText,
	}, nil
}

// ExtractReceiptDataFromBase64 is a convenience method for base64 encoded images
func (s *Service) ExtractReceiptDataFromBase64(ctx context.Context, base64Image string, hints map[string]string) (*ReceiptExtractionResponse, error) {
	req := ReceiptExtractionRequest{
		ImageData: base64Image,
	}

	if hints != nil {
		if currency, ok := hints["currency"]; ok {
			req.ExpectedCurrency = currency
		}
		if language, ok := hints["language"]; ok {
			req.ExpectedLanguage = language
		}
		if store, ok := hints["store"]; ok {
			req.StoreHint = store
		}
	}

	return s.ExtractReceiptData(ctx, req)
}

// ExtractReceiptDataFromURL is a convenience method for image URLs
func (s *Service) ExtractReceiptDataFromURL(ctx context.Context, imageURL string, hints map[string]string) (*ReceiptExtractionResponse, error) {
	req := ReceiptExtractionRequest{
		ImageURL: imageURL,
	}

	if hints != nil {
		if currency, ok := hints["currency"]; ok {
			req.ExpectedCurrency = currency
		}
		if language, ok := hints["language"]; ok {
			req.ExpectedLanguage = language
		}
		if store, ok := hints["store"]; ok {
			req.StoreHint = store
		}
	}

	return s.ExtractReceiptData(ctx, req)
}

// buildReceiptExtractionPrompt creates a comprehensive prompt for receipt extraction
func (s *Service) buildReceiptExtractionPrompt(req ReceiptExtractionRequest) string {
	// Use config defaults if not specified in request
	currency := req.ExpectedCurrency
	if currency == "" {
		currency = s.config.DefaultCurrency
	}

	language := req.ExpectedLanguage
	if language == "" {
		language = s.config.DefaultLanguage
	}

	prompt := fmt.Sprintf(`You are an expert at extracting structured data from receipt images. Analyze this receipt image and extract all available information in JSON format.

Instructions:
1. Extract the receipt date and convert it to ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
2. Extract all items with their names, quantities, unit prices, and total prices
3. Identify the store name, address, and phone number if visible
4. Extract the total amount, currency, tax, subtotal, discounts, and tips
5. Look for payment method, receipt number, transaction ID, cashier name, register number
6. Extract any other relevant information you can find
7. If the currency is not visible, assume: %s
8. The receipt may be in: %s (or other languages - detect automatically)
9. Be precise with numbers and dates
10. If information is unclear or not visible, omit that field or set it to null

11. Classify the receipt into ONE expense category for household budget tracking:
   - "식비" (Food & Groceries) - restaurants, supermarkets, convenience stores, cafes
   - "교통비" (Transportation) - gas stations, tolls, parking, public transport
   - "생활용품" (Household Items) - home supplies, cleaning products, furniture
   - "의료" (Medical) - pharmacies, hospitals, clinics
   - "문화/여가" (Culture/Leisure) - movies, books, entertainment, sports
   - "교육" (Education) - books, courses, supplies
   - "통신" (Communication) - phone bills, internet
   - "기타" (Other) - anything else

Return ONLY a valid JSON object matching this structure:
{
  "store_name": "string",
  "receipt_date": "2024-01-01T12:00:00Z",
  "total_amount": 0.0,
  "currency": "USD",
  "items": [
    {
      "name": "string",
      "quantity": 1.0,
      "unit_price": 0.0,
      "total_price": 0.0,
      "category": "string",
      "sku": "string",
      "discount": 0.0,
      "tax_amount": 0.0,
      "description": "string"
    }
  ],
  "store_address": "string",
  "store_phone": "string",
  "tax_amount": 0.0,
  "subtotal_amount": 0.0,
  "discount_amount": 0.0,
  "tip_amount": 0.0,
  "payment_method": "string",
  "card_last_digits": "string",
  "receipt_number": "string",
  "transaction_id": "string",
  "cashier_name": "string",
  "register_number": "string",
  "expense_category": "식비",
  "notes": "string",
  "confidence_level": 0.95
}

Do not include any markdown formatting, explanations, or text outside the JSON object.`, currency, language)

	if req.StoreHint != "" {
		prompt += fmt.Sprintf("\n\nAdditional context: This receipt is likely from %s", req.StoreHint)
	}

	return prompt
}

// callVisionAPI makes the actual API call to OpenAI
func (s *Service) callVisionAPI(ctx context.Context, imageURL string, prompt string) (*ReceiptData, string, error) {
	// Prepare the API request
	apiReq := openAIChatRequest{
		Model:       s.config.VisionModel,
		MaxTokens:   s.config.MaxTokens,
		Temperature: s.config.Temperature,
		Messages: []openAIChatMessage{
			{
				Role: "user",
				Content: []openAIMessageContent{
					{
						Type: "text",
						Text: prompt,
					},
					{
						Type: "image_url",
						ImageURL: &openAIImageURL{
							URL:    imageURL,
							Detail: "high", // Use high detail for better accuracy
						},
					},
				},
			},
		},
		ResponseFormat: &openAIResponseFormat{
			Type: "json_object",
		},
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(apiReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// Make the request
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		var apiError openAIChatResponse
		if err := json.Unmarshal(responseBody, &apiError); err == nil && apiError.Error != nil {
			return nil, "", fmt.Errorf("OpenAI API error: %s", apiError.Error.Message)
		}
		return nil, "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse response
	var apiResp openAIChatResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the content
	if len(apiResp.Choices) == 0 {
		return nil, "", fmt.Errorf("no choices returned from API")
	}

	content := apiResp.Choices[0].Message.Content

	// Parse the JSON response into ReceiptData
	var receiptData ReceiptData
	if err := json.Unmarshal([]byte(content), &receiptData); err != nil {
		return nil, content, fmt.Errorf("failed to parse receipt data: %w", err)
	}

	return &receiptData, content, nil
}

// detectImageMimeType detects the image MIME type from base64 encoded data
func detectImageMimeType(base64Data string) string {
	// Decode first few bytes to check magic numbers
	if len(base64Data) < 20 {
		return "image/jpeg" // default
	}

	// Decode first 12 bytes (enough to detect most formats)
	decoded, err := base64.StdEncoding.DecodeString(base64Data[:min(len(base64Data), 16)])
	if err != nil || len(decoded) < 4 {
		return "image/jpeg" // default
	}

	// Check magic bytes
	switch {
	case decoded[0] == 0x89 && decoded[1] == 0x50 && decoded[2] == 0x4E && decoded[3] == 0x47:
		return "image/png"
	case decoded[0] == 0xFF && decoded[1] == 0xD8 && decoded[2] == 0xFF:
		return "image/jpeg"
	case decoded[0] == 0x47 && decoded[1] == 0x49 && decoded[2] == 0x46:
		return "image/gif"
	case decoded[0] == 0x52 && decoded[1] == 0x49 && decoded[2] == 0x46 && decoded[3] == 0x46:
		// RIFF format, check if it's WebP
		if len(decoded) >= 12 && decoded[8] == 0x57 && decoded[9] == 0x45 && decoded[10] == 0x42 && decoded[11] == 0x50 {
			return "image/webp"
		}
		return "image/jpeg" // default
	case decoded[0] == 0x42 && decoded[1] == 0x4D:
		return "image/bmp"
	default:
		return "image/jpeg" // default
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// EncodeImageToBase64 is a helper function to encode image bytes to base64
func EncodeImageToBase64(imageBytes []byte) string {
	return base64.StdEncoding.EncodeToString(imageBytes)
}

// PrepareImageDataURI creates a data URI from base64 image data
func PrepareImageDataURI(base64Data string, mimeType string) string {
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
}
