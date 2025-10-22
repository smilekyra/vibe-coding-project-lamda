package service

import (
	"testing"
	"time"

	"vibe-coding-project-lambda/shared/openai"
)

func TestFormatReceiptRow(t *testing.T) {
	service := &SheetsService{
		sheetName: "테스트",
	}

	tests := []struct {
		name           string
		receiptData    *openai.ReceiptData
		receiptURL     string
		memo           string
		wantTotalType  string // "number" or "string"
		wantTotalValue interface{}
	}{
		{
			name: "Valid receipt with amount",
			receiptData: &openai.ReceiptData{
				StoreName:   "セブンイレブン",
				ReceiptDate: time.Date(2024, 10, 18, 0, 0, 0, 0, time.UTC),
				TotalAmount: 1250.50,
				Currency:    "JPY",
				Items: []openai.ReceiptItem{
					{Name: "Item 1", TotalPrice: 500},
					{Name: "Item 2", TotalPrice: 750.50},
				},
				PaymentMethod: "Credit Card",
			},
			receiptURL:     "https://s3.example.com/receipt.jpg",
			memo:           "Test memo",
			wantTotalType:  "number",
			wantTotalValue: 1250.50,
		},
		{
			name: "Receipt without amount",
			receiptData: &openai.ReceiptData{
				StoreName:   "Store",
				ReceiptDate: time.Date(2024, 10, 18, 0, 0, 0, 0, time.UTC),
				TotalAmount: 0, // No amount
			},
			receiptURL:     "https://s3.example.com/receipt.jpg",
			memo:           "",
			wantTotalType:  "string",
			wantTotalValue: "",
		},
		{
			name:           "Nil receipt data",
			receiptData:    nil,
			receiptURL:     "https://s3.example.com/receipt.jpg",
			memo:           "",
			wantTotalType:  "string",
			wantTotalValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := service.formatReceiptRow(tt.receiptData, tt.receiptURL, tt.memo)

			// Check row length
			expectedLength := 9 // 날짜,카테고리,상점명,총금액,항목수,항목내역,결제방법,영수증링크,메모
			if len(row) != expectedLength {
				t.Errorf("Row length = %d, want %d", len(row), expectedLength)
			}

			// Check total amount (index 3, after adding category)
			totalAmount := row[3]

			// Verify it's the correct type
			switch tt.wantTotalType {
			case "number":
				if _, ok := totalAmount.(float64); !ok {
					t.Errorf("Total amount type = %T, want float64 (number)", totalAmount)
				}
				if totalAmount != tt.wantTotalValue {
					t.Errorf("Total amount = %v, want %v", totalAmount, tt.wantTotalValue)
				}
			case "string":
				if _, ok := totalAmount.(string); !ok {
					t.Errorf("Total amount type = %T, want string", totalAmount)
				}
				if totalAmount != tt.wantTotalValue {
					t.Errorf("Total amount = %v, want %v", totalAmount, tt.wantTotalValue)
				}
			}
		})
	}
}

func TestFormatReceiptRow_NumberTypeForCalculations(t *testing.T) {
	service := &SheetsService{
		sheetName: "가계부",
	}

	// Create receipt with specific amount
	receiptData := &openai.ReceiptData{
		StoreName:     "Test Store",
		ReceiptDate:   time.Now(),
		TotalAmount:   1500.75,
		Currency:      "JPY",
		PaymentMethod: "Cash",
	}

	row := service.formatReceiptRow(receiptData, "https://example.com/receipt.jpg", "")

	// The total amount should be at index 3 (날짜, 카테고리, 상점명, 총금액...)
	totalAmount := row[3]

	// CRITICAL: It must be a number (float64) not a string
	// This allows Google Sheets to calculate SUM, AVERAGE, etc.
	amount, ok := totalAmount.(float64)
	if !ok {
		t.Fatalf("Total amount must be float64 for calculations, got %T", totalAmount)
	}

	// Verify the exact value
	if amount != 1500.75 {
		t.Errorf("Total amount = %f, want 1500.75", amount)
	}

	// This demonstrates that the value is a number, not a formatted string like "JPY 1500.75"
	// In Google Sheets, this will allow formulas like =SUM(D2:D10) to work correctly
	t.Logf("✅ Total amount stored as number: %f (ready for calculations)", amount)
}

func TestFormatReceiptRow_ItemsSummaryColumn(t *testing.T) {
	service := &SheetsService{
		sheetName: "가계부",
	}

	tests := []struct {
		name             string
		items            []openai.ReceiptItem
		wantItemsSummary string
		wantItemCount    int
	}{
		{
			name: "Multiple items",
			items: []openai.ReceiptItem{
				{Name: "Coffee", TotalPrice: 500},
				{Name: "Sandwich", TotalPrice: 750},
				{Name: "Water", TotalPrice: 200},
			},
			wantItemsSummary: "Coffee, Sandwich, Water",
			wantItemCount:    3,
		},
		{
			name: "Single item",
			items: []openai.ReceiptItem{
				{Name: "Gasoline", TotalPrice: 5500},
			},
			wantItemsSummary: "Gasoline",
			wantItemCount:    1,
		},
		{
			name:             "No items",
			items:            []openai.ReceiptItem{},
			wantItemsSummary: "",
			wantItemCount:    0,
		},
		{
			name: "Items with empty names",
			items: []openai.ReceiptItem{
				{Name: "Apple", TotalPrice: 100},
				{Name: "", TotalPrice: 200},
				{Name: "Banana", TotalPrice: 150},
			},
			wantItemsSummary: "Apple, Banana",
			wantItemCount:    3, // Count includes all items
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiptData := &openai.ReceiptData{
				StoreName:   "Test Store",
				ReceiptDate: time.Now(),
				TotalAmount: 1000,
				Items:       tt.items,
			}

			row := service.formatReceiptRow(receiptData, "https://example.com/receipt.jpg", "")

			// Items summary should be at index 5 (날짜, 카테고리, 상점명, 총금액, 항목수, 항목내역, ...)
			itemsSummary := row[5]
			itemsSummaryStr, ok := itemsSummary.(string)
			if !ok {
				t.Fatalf("Items summary must be string, got %T", itemsSummary)
			}

			if itemsSummaryStr != tt.wantItemsSummary {
				t.Errorf("Items summary = %s, want %s", itemsSummaryStr, tt.wantItemsSummary)
			}

			// Item count should be at index 4
			itemCount := row[4]
			itemCountInt, ok := itemCount.(int)
			if !ok {
				t.Fatalf("Item count must be int, got %T", itemCount)
			}

			if itemCountInt != tt.wantItemCount {
				t.Errorf("Item count = %d, want %d", itemCountInt, tt.wantItemCount)
			}

			t.Logf("✅ Items: %d items - %s", itemCountInt, itemsSummaryStr)
		})
	}
}

func TestFormatReceiptRow_CategoryColumn(t *testing.T) {
	service := &SheetsService{
		sheetName: "가계부",
	}

	tests := []struct {
		name            string
		expenseCategory string
		wantCategory    string
	}{
		{
			name:            "Food category",
			expenseCategory: "식비",
			wantCategory:    "식비",
		},
		{
			name:            "Transportation category",
			expenseCategory: "교통비",
			wantCategory:    "교통비",
		},
		{
			name:            "No category - should default",
			expenseCategory: "",
			wantCategory:    "미분류",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiptData := &openai.ReceiptData{
				StoreName:       "Test Store",
				ReceiptDate:     time.Now(),
				TotalAmount:     1000,
				ExpenseCategory: tt.expenseCategory,
			}

			row := service.formatReceiptRow(receiptData, "https://example.com/receipt.jpg", "")

			// Category should be at index 1 (날짜, 카테고리, ...)
			category := row[1]
			categoryStr, ok := category.(string)
			if !ok {
				t.Fatalf("Category must be string, got %T", category)
			}

			if categoryStr != tt.wantCategory {
				t.Errorf("Category = %s, want %s", categoryStr, tt.wantCategory)
			}

			t.Logf("✅ Category: %s", categoryStr)
		})
	}
}
