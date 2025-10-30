package main

// ReceiptData represents the structured data extracted from a receipt
type ReceiptData struct {
	MerchantName     string        `json:"merchant_name"`
	MerchantAddress  string        `json:"merchant_address"`
	PhoneNumber      string        `json:"phone_number"`
	TransactionDate  string        `json:"transaction_date"`
	TransactionTime  string        `json:"transaction_time"`
	Items            []ReceiptItem `json:"items"`
	Subtotal         float64       `json:"subtotal"`
	Tax              float64       `json:"tax"`
	Total            float64       `json:"total"`
	PaymentMethod    string        `json:"payment_method"`
	CardLastFour     string        `json:"card_last_four"`
	ReceiptNumber    string        `json:"receipt_number"`
	CashierName      string        `json:"cashier_name"`
}

// ReceiptItem represents an individual item on the receipt
type ReceiptItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Total    float64 `json:"total"`
}

// ExtractionRequest represents the request to extract receipt data
type ExtractionRequest struct {
	ImageData []byte `json:"-"` // Image data in bytes
}

// ExtractionResponse represents the response after extracting receipt data
type ExtractionResponse struct {
	Success bool         `json:"success"`
	Data    *ReceiptData `json:"data,omitempty"`
	Error   string       `json:"error,omitempty"`
}
