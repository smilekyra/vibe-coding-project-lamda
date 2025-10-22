package service

import (
	"context"
	"fmt"
	"log"

	"vibe-coding-project-lambda/shared/openai"
	"vibe-coding-project-lambda/shared/repository"
)

// SheetsService handles business logic for Google Sheets operations
type SheetsService struct {
	sheetsRepo *repository.SheetsRepository
	sheetName  string // The name of the sheet tab (e.g., "가계부", "Sheet1")
}

// SheetsServiceConfig contains configuration for the sheets service
type SheetsServiceConfig struct {
	SheetsRepo *repository.SheetsRepository
	SheetName  string // Default sheet name to use
}

// NewSheetsService creates a new sheets service
func NewSheetsService(config SheetsServiceConfig) *SheetsService {
	sheetName := config.SheetName
	if sheetName == "" {
		sheetName = "Sheet1" // Default to Sheet1 if not specified
	}

	return &SheetsService{
		sheetsRepo: config.SheetsRepo,
		sheetName:  sheetName,
	}
}

// AddReceiptToSpreadsheet adds receipt data to the spreadsheet
// Expected columns: 날짜,상점명,총금액,항목수,결제방법,영수증링크,메모
func (s *SheetsService) AddReceiptToSpreadsheet(ctx context.Context, receiptData *openai.ReceiptData, receiptURL string, memo string) error {
	if s.sheetsRepo == nil {
		return fmt.Errorf("sheets repository not initialized")
	}

	row := s.formatReceiptRow(receiptData, receiptURL, memo)

	log.Printf("Adding receipt to spreadsheet: date=%s, store=%s, total=%s",
		row[0], row[1], row[2])

	err := s.sheetsRepo.AppendRow(ctx, s.sheetName, row)
	if err != nil {
		return fmt.Errorf("failed to add receipt to spreadsheet: %w", err)
	}

	log.Printf("Successfully added receipt to spreadsheet")
	return nil
}

// AddMultipleReceipts adds multiple receipts to the spreadsheet in one batch
func (s *SheetsService) AddMultipleReceipts(ctx context.Context, receipts []ReceiptEntry) error {
	if s.sheetsRepo == nil {
		return fmt.Errorf("sheets repository not initialized")
	}

	if len(receipts) == 0 {
		return nil
	}

	rows := make([][]interface{}, len(receipts))
	for i, receipt := range receipts {
		rows[i] = s.formatReceiptRow(receipt.Data, receipt.ReceiptURL, receipt.Memo)
	}

	log.Printf("Adding %d receipts to spreadsheet", len(receipts))

	err := s.sheetsRepo.AppendRows(ctx, s.sheetName, rows)
	if err != nil {
		return fmt.Errorf("failed to add multiple receipts: %w", err)
	}

	log.Printf("Successfully added %d receipts to spreadsheet", len(receipts))
	return nil
}

// ReceiptEntry represents a single receipt entry to be added to the spreadsheet
type ReceiptEntry struct {
	Data       *openai.ReceiptData
	ReceiptURL string
	Memo       string
}

// formatReceiptRow formats receipt data into a spreadsheet row
// Columns: 날짜,카테고리,상점명,총금액,항목수,항목내역,결제방법,영수증링크,메모
func (s *SheetsService) formatReceiptRow(data *openai.ReceiptData, receiptURL string, memo string) []interface{} {
	// Default values
	date := ""
	category := ""
	storeName := ""
	var totalAmount interface{} = "" // Can be string (empty) or float64 (number)
	itemCount := 0
	itemsSummary := ""
	paymentMethod := ""

	if data != nil {
		// Date (날짜)
		if !data.ReceiptDate.IsZero() {
			// Format date as YYYY-MM-DD
			date = data.ReceiptDate.Format("2006-01-02")
		}

		// Category (카테고리)
		if data.ExpenseCategory != "" {
			category = data.ExpenseCategory
		} else {
			category = "미분류" // Uncategorized
		}

		// Store name (상점명)
		if data.StoreName != "" {
			storeName = data.StoreName
		}

		// Total amount as number (총금액)
		// Store as pure number for calculations in Google Sheets
		if data.TotalAmount > 0 {
			totalAmount = data.TotalAmount // Store as number, not formatted string
		}

		// Item count and summary (항목수, 항목내역)
		itemCount = len(data.Items)
		if itemCount > 0 {
			itemNames := make([]string, 0, itemCount)
			for _, item := range data.Items {
				if item.Name != "" {
					itemNames = append(itemNames, item.Name)
				}
			}
			// Join with comma and space for readability
			if len(itemNames) > 0 {
				itemsSummary = fmt.Sprintf("%s", itemNames[0])
				for i := 1; i < len(itemNames); i++ {
					itemsSummary += ", " + itemNames[i]
				}
			}
		}

		// Payment method (결제방법)
		if data.PaymentMethod != "" {
			paymentMethod = data.PaymentMethod
		} else {
			paymentMethod = "알 수 없음" // Unknown
		}
	}

	// Build row: 날짜,카테고리,상점명,총금액,항목수,항목내역,결제방법,영수증링크,메모
	return []interface{}{
		date,          // 날짜
		category,      // 카테고리
		storeName,     // 상점명
		totalAmount,   // 총금액 (as number for calculations)
		itemCount,     // 항목수
		itemsSummary,  // 항목내역 (comma-separated item names)
		paymentMethod, // 결제방법
		receiptURL,    // 영수증링크
		memo,          // 메모
	}
}

// InitializeSpreadsheet sets up the spreadsheet with headers if needed
func (s *SheetsService) InitializeSpreadsheet(ctx context.Context) error {
	if s.sheetsRepo == nil {
		return fmt.Errorf("sheets repository not initialized")
	}

	// Check if sheet already has data
	rangeNotation := fmt.Sprintf("%s!A1:I1", s.sheetName)
	values, err := s.sheetsRepo.ReadRange(ctx, rangeNotation)
	if err != nil {
		// If error reading, assume sheet doesn't exist or is empty
		log.Printf("Sheet appears to be new or empty, will create headers: %v", err)
	}

	// If first row is empty, add headers
	if len(values) == 0 || len(values[0]) == 0 {
		headers := []interface{}{
			"날짜",
			"카테고리",
			"상점명",
			"총금액",
			"항목수",
			"항목내역",
			"결제방법",
			"영수증링크",
			"메모",
		}

		log.Printf("Adding headers to spreadsheet: %s", s.sheetName)
		err := s.sheetsRepo.AppendRow(ctx, s.sheetName, headers)
		if err != nil {
			return fmt.Errorf("failed to add headers: %w", err)
		}
		log.Printf("Successfully initialized spreadsheet with headers")
	} else {
		log.Printf("Spreadsheet already has headers, skipping initialization")
	}

	return nil
}

// GetRecentReceipts retrieves recent receipt entries from the spreadsheet
func (s *SheetsService) GetRecentReceipts(ctx context.Context, limit int) ([][]interface{}, error) {
	if s.sheetsRepo == nil {
		return nil, fmt.Errorf("sheets repository not initialized")
	}

	// Read recent rows (skip header row)
	rangeNotation := fmt.Sprintf("%s!A2:I%d", s.sheetName, limit+1)
	values, err := s.sheetsRepo.ReadRange(ctx, rangeNotation)
	if err != nil {
		return nil, fmt.Errorf("failed to read recent receipts: %w", err)
	}

	return values, nil
}
