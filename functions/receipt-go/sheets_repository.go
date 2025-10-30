package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsRepository interface for storing receipt data to Google Sheets
type SheetsRepository interface {
	SaveReceipt(ctx context.Context, data *ReceiptData, s3URL string) error
}

// GoogleSheetsRepository implements SheetsRepository using Google Sheets API
type GoogleSheetsRepository struct {
	service     *sheets.Service
	spreadsheet string
}

// SheetRow represents a row in the Google Sheets
// 날짜    카테고리    상점명    총금액    항목수    항목내역    결제방법    영수증링크    메모
type SheetRow struct {
	Date          string // 날짜
	Category      string // 카테고리
	MerchantName  string // 상점명
	Total         string // 총금액
	ItemCount     string // 항목수
	ItemDetails   string // 항목내역
	PaymentMethod string // 결제방법
	ReceiptLink   string // 영수증링크
	Memo          string // 메모
}

// NewGoogleSheetsRepository creates a new Google Sheets repository
// credentialsJSON: Google service account credentials JSON
// spreadsheetID: The ID of the Google Spreadsheet to write to
func NewGoogleSheetsRepository(ctx context.Context, credentialsJSON []byte, spreadsheetID string) (*GoogleSheetsRepository, error) {
	log.Printf("[INFO] Creating Google Sheets repository - spreadsheetID: %s", spreadsheetID)

	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		log.Printf("[ERROR] Failed to create sheets service: %v", err)
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &GoogleSheetsRepository{
		service:     srv,
		spreadsheet: spreadsheetID,
	}, nil
}

// SaveReceipt saves receipt data to Google Sheets
func (r *GoogleSheetsRepository) SaveReceipt(ctx context.Context, data *ReceiptData, s3URL string) error {
	log.Printf("[INFO] Saving receipt to Google Sheets - merchant: %s, total: %.2f", data.MerchantName, data.Total)

	row := r.convertToSheetRow(data, s3URL)
	values := []interface{}{
		row.Date,
		row.Category,
		row.MerchantName,
		row.Total,
		row.ItemCount,
		row.ItemDetails,
		row.PaymentMethod,
		row.ReceiptLink,
		row.Memo,
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	// Append to the spreadsheet
	// Assumes the first sheet is the target sheet
	sheetRange := "Sheet1!A:I" // A to I columns for 9 fields

	_, err := r.service.Spreadsheets.Values.Append(r.spreadsheet, sheetRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()

	if err != nil {
		log.Printf("[ERROR] Failed to append to Google Sheets: %v", err)
		return fmt.Errorf("failed to append to Google Sheets: %w", err)
	}

	log.Printf("[INFO] Successfully saved receipt to Google Sheets - merchant: %s", data.MerchantName)
	return nil
}

// convertToSheetRow converts ReceiptData to SheetRow format
func (r *GoogleSheetsRepository) convertToSheetRow(data *ReceiptData, s3URL string) *SheetRow {
	// Parse date from transaction date (format may vary)
	date := r.formatDate(data.TransactionDate)

	// Category - can be set to empty or implement category logic
	category := "" // TODO: Implement category detection

	// Merchant name
	merchantName := data.MerchantName

	// Total amount
	total := fmt.Sprintf("%.2f", data.Total)

	// Item count
	itemCount := fmt.Sprintf("%d", len(data.Items))

	// Item details - format as "item1 (qty) x price, item2 (qty) x price, ..."
	itemDetails := r.formatItemDetails(data.Items)

	// Payment method
	paymentMethod := r.formatPaymentMethod(data.PaymentMethod, data.CardLastFour)

	// Receipt link (S3 URL)
	receiptLink := s3URL

	// Memo - initially empty
	memo := ""

	return &SheetRow{
		Date:          date,
		Category:      category,
		MerchantName:  merchantName,
		Total:         total,
		ItemCount:     itemCount,
		ItemDetails:   itemDetails,
		PaymentMethod: paymentMethod,
		ReceiptLink:   receiptLink,
		Memo:          memo,
	}
}

// formatDate formats the transaction date
func (r *GoogleSheetsRepository) formatDate(dateStr string) string {
	if dateStr == "" {
		return time.Now().Format("2006-01-02")
	}

	// Try to parse common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"Jan 02, 2006",
		"January 02, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Format("2006-01-02")
		}
	}

	// If parsing fails, return as-is
	log.Printf("[WARN] Failed to parse date: %s, using as-is", dateStr)
	return dateStr
}

// formatItemDetails formats receipt items into a readable string
func (r *GoogleSheetsRepository) formatItemDetails(items []ReceiptItem) string {
	if len(items) == 0 {
		return ""
	}

	var parts []string
	for _, item := range items {
		if item.Quantity > 1 {
			parts = append(parts, fmt.Sprintf("%s (%d개) x $%.2f", item.Name, item.Quantity, item.Price))
		} else {
			parts = append(parts, fmt.Sprintf("%s x $%.2f", item.Name, item.Price))
		}
	}

	return strings.Join(parts, ", ")
}

// formatPaymentMethod formats payment method with card details if available
func (r *GoogleSheetsRepository) formatPaymentMethod(method, cardLastFour string) string {
	if method == "" {
		return "Unknown"
	}

	if cardLastFour != "" {
		return fmt.Sprintf("%s ****%s", method, cardLastFour)
	}

	return method
}
