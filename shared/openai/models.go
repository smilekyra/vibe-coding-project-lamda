package openai

import "time"

// ReceiptData represents the structured data extracted from a receipt
type ReceiptData struct {
	// Core fields
	StoreName   string    `json:"store_name"`
	ReceiptDate time.Time `json:"receipt_date"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`

	// Items
	Items []ReceiptItem `json:"items"`

	// Additional fields
	StoreAddress   string  `json:"store_address,omitempty"`
	StorePhone     string  `json:"store_phone,omitempty"`
	TaxAmount      float64 `json:"tax_amount,omitempty"`
	SubtotalAmount float64 `json:"subtotal_amount,omitempty"`
	DiscountAmount float64 `json:"discount_amount,omitempty"`
	TipAmount      float64 `json:"tip_amount,omitempty"`

	// Payment details
	PaymentMethod  string `json:"payment_method,omitempty"`
	CardLastDigits string `json:"card_last_digits,omitempty"`

	// Receipt metadata
	ReceiptNumber  string `json:"receipt_number,omitempty"`
	TransactionID  string `json:"transaction_id,omitempty"`
	CashierName    string `json:"cashier_name,omitempty"`
	RegisterNumber string `json:"register_number,omitempty"`

	// Expense tracking for household budget
	ExpenseCategory string `json:"expense_category,omitempty"` // 식비, 교통비, 생활용품, 의료, 문화/여가, 교육, 통신, 기타

	// Additional information
	Notes           string            `json:"notes,omitempty"`
	CustomFields    map[string]string `json:"custom_fields,omitempty"`
	RawText         string            `json:"raw_text,omitempty"`         // Original OCR text
	ConfidenceLevel float64           `json:"confidence_level,omitempty"` // 0-1 scale
}

// ReceiptItem represents a single item from a receipt
type ReceiptItem struct {
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
	Category    string  `json:"category,omitempty"`
	SKU         string  `json:"sku,omitempty"`
	Discount    float64 `json:"discount,omitempty"`
	TaxAmount   float64 `json:"tax_amount,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ReceiptExtractionRequest represents the request to extract receipt data
type ReceiptExtractionRequest struct {
	// Image data (base64 encoded or URL)
	ImageData string `json:"image_data"`
	ImageURL  string `json:"image_url,omitempty"`

	// Optional hints for better extraction
	ExpectedCurrency string `json:"expected_currency,omitempty"`
	ExpectedLanguage string `json:"expected_language,omitempty"`
	StoreHint        string `json:"store_hint,omitempty"`
}

// ReceiptExtractionResponse represents the response from receipt extraction
type ReceiptExtractionResponse struct {
	Success bool         `json:"success"`
	Data    *ReceiptData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
	RawText string       `json:"raw_text,omitempty"`
}
