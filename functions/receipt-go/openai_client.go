package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// OpenAIClient interface for OpenAI API operations
type OpenAIClient interface {
	ExtractReceiptData(ctx context.Context, imageData []byte) (*ReceiptData, error)
}

// RealOpenAIClient implements OpenAIClient using OpenAI API
type RealOpenAIClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string) *RealOpenAIClient {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	if apiKey == "" {
		log.Printf("[WARN] OpenAI API key not provided")
	} else {
		log.Printf("[INFO] OpenAI client initialized")
	}

	return &RealOpenAIClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// OpenAI API request/response structures
type openAIRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIMessage       `json:"messages"`
	ResponseFormat *openAIResponseFormat `json:"response_format"`
	MaxTokens      int                   `json:"max_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string                   `json:"role"`
	Content []openAIMessageContent   `json:"content"`
}

type openAIMessageContent struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *openAIImageURL   `json:"image_url,omitempty"`
}

type openAIImageURL struct {
	URL string `json:"url"`
}

type openAIResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *openAIJSONSchema `json:"json_schema"`
}

type openAIJSONSchema struct {
	Name   string                 `json:"name"`
	Strict bool                   `json:"strict"`
	Schema map[string]interface{} `json:"schema"`
}

type openAIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []openAIChoice   `json:"choices"`
	Usage   openAIUsage      `json:"usage"`
}

type openAIChoice struct {
	Index        int              `json:"index"`
	Message      openAIRespMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

type openAIRespMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Refusal string `json:"refusal,omitempty"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ExtractReceiptData extracts structured data from a receipt image using OpenAI Vision API
func (c *RealOpenAIClient) ExtractReceiptData(ctx context.Context, imageData []byte) (*ReceiptData, error) {
	log.Printf("[INFO] Starting receipt data extraction - imageSize: %d bytes", len(imageData))

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Image)

	// Define the JSON schema for structured output
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"merchant_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the merchant or store",
			},
			"merchant_address": map[string]interface{}{
				"type":        "string",
				"description": "The address of the merchant",
			},
			"phone_number": map[string]interface{}{
				"type":        "string",
				"description": "The phone number of the merchant",
			},
			"transaction_date": map[string]interface{}{
				"type":        "string",
				"description": "The date of the transaction in YYYY-MM-DD format",
			},
			"transaction_time": map[string]interface{}{
				"type":        "string",
				"description": "The time of the transaction in HH:MM:SS format",
			},
			"items": map[string]interface{}{
				"type":        "array",
				"description": "List of items purchased",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the item",
						},
						"quantity": map[string]interface{}{
							"type":        "integer",
							"description": "The quantity of the item",
						},
						"price": map[string]interface{}{
							"type":        "number",
							"description": "The unit price of the item",
						},
						"total": map[string]interface{}{
							"type":        "number",
							"description": "The total price for this item (quantity * price)",
						},
					},
					"required":             []string{"name", "quantity", "price", "total"},
					"additionalProperties": false,
				},
			},
			"subtotal": map[string]interface{}{
				"type":        "number",
				"description": "The subtotal amount before tax",
			},
			"tax": map[string]interface{}{
				"type":        "number",
				"description": "The tax amount",
			},
			"total": map[string]interface{}{
				"type":        "number",
				"description": "The total amount including tax",
			},
			"payment_method": map[string]interface{}{
				"type":        "string",
				"description": "The payment method used (e.g., CASH, CREDIT, DEBIT)",
			},
			"card_last_four": map[string]interface{}{
				"type":        "string",
				"description": "The last four digits of the card if applicable",
			},
			"receipt_number": map[string]interface{}{
				"type":        "string",
				"description": "The receipt or transaction number",
			},
			"cashier_name": map[string]interface{}{
				"type":        "string",
				"description": "The name of the cashier",
			},
		},
		"required": []string{
			"merchant_name",
			"merchant_address",
			"phone_number",
			"transaction_date",
			"transaction_time",
			"items",
			"subtotal",
			"tax",
			"total",
			"payment_method",
			"card_last_four",
			"receipt_number",
			"cashier_name",
		},
		"additionalProperties": false,
	}

	// Create the API request
	reqBody := openAIRequest{
		Model: "gpt-4o-2024-08-06",
		Messages: []openAIMessage{
			{
				Role: "system",
				Content: []openAIMessageContent{
					{
						Type: "text",
						Text: "You are an expert at extracting structured data from receipt images. Extract all relevant information from the receipt and return it in the specified JSON format. If any field is not found, use empty string for strings, 0 for numbers, and empty array for items.",
					},
				},
			},
			{
				Role: "user",
				Content: []openAIMessageContent{
					{
						Type: "text",
						Text: "Extract all the information from this receipt image and structure it according to the schema.",
					},
					{
						Type: "image_url",
						ImageURL: &openAIImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
		ResponseFormat: &openAIResponseFormat{
			Type: "json_schema",
			JSONSchema: &openAIJSONSchema{
				Name:   "receipt_extraction",
				Strict: true,
				Schema: schema,
			},
		},
		MaxTokens: 2000,
	}

	// Marshal request to JSON
	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal OpenAI request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[INFO] Calling OpenAI API - model: %s", reqBody.Model)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Printf("[ERROR] Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIKey))

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] OpenAI API request failed: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read OpenAI response body: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] OpenAI API returned non-OK status - statusCode: %d, response: %s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		log.Printf("[ERROR] Failed to parse OpenAI response JSON: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Log token usage
	log.Printf("[INFO] OpenAI API call successful - promptTokens: %d, completionTokens: %d, totalTokens: %d",
		openAIResp.Usage.PromptTokens, openAIResp.Usage.CompletionTokens, openAIResp.Usage.TotalTokens)

	// Check if there are choices
	if len(openAIResp.Choices) == 0 {
		log.Printf("[ERROR] OpenAI response contains no choices")
		return nil, fmt.Errorf("no choices in response")
	}

	// Check for refusal
	if openAIResp.Choices[0].Message.Refusal != "" {
		log.Printf("[ERROR] OpenAI request refused: %s", openAIResp.Choices[0].Message.Refusal)
		return nil, fmt.Errorf("request refused: %s", openAIResp.Choices[0].Message.Refusal)
	}

	// Parse the structured data
	var receiptData ReceiptData
	if err := json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &receiptData); err != nil {
		log.Printf("[ERROR] Failed to parse receipt data from OpenAI response: %v", err)
		return nil, fmt.Errorf("failed to parse receipt data: %w", err)
	}

	log.Printf("[INFO] Receipt data extracted successfully - merchant: %s, total: %.2f, items: %d",
		receiptData.MerchantName, receiptData.Total, len(receiptData.Items))

	return &receiptData, nil
}
