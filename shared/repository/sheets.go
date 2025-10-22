package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsRepository handles Google Sheets operations
type SheetsRepository struct {
	service       *sheets.Service
	spreadsheetID string
}

// SheetsConfig contains configuration for Google Sheets
type SheetsConfig struct {
	// ServiceAccountJSON is the JSON content of the service account key file
	ServiceAccountJSON []byte
	// SpreadsheetID is the ID of the Google Spreadsheet
	SpreadsheetID string
	// Scopes defines the access level (default: spreadsheets scope)
	Scopes []string
}

// NewSheetsRepository creates a new Google Sheets repository
func NewSheetsRepository(ctx context.Context, config SheetsConfig) (*SheetsRepository, error) {
	if len(config.ServiceAccountJSON) == 0 {
		return nil, fmt.Errorf("service account JSON is required")
	}
	if config.SpreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet ID is required")
	}

	// Default to full spreadsheet access if no scopes provided
	scopes := config.Scopes
	if len(scopes) == 0 {
		scopes = []string{sheets.SpreadsheetsScope}
	}

	// Create credentials from service account JSON
	credentials, err := google.CredentialsFromJSON(ctx, config.ServiceAccountJSON, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	// Create Sheets service
	service, err := sheets.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &SheetsRepository{
		service:       service,
		spreadsheetID: config.SpreadsheetID,
	}, nil
}

// AppendRow appends a row of values to the specified sheet
// sheetName: the name of the sheet tab (e.g., "Sheet1")
// values: the row data to append
func (r *SheetsRepository) AppendRow(ctx context.Context, sheetName string, values []interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := r.service.Spreadsheets.Values.Append(
		r.spreadsheetID,
		sheetName,
		valueRange,
	).ValueInputOption("USER_ENTERED").Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("failed to append row: %w", err)
	}

	return nil
}

// AppendRows appends multiple rows to the specified sheet
func (r *SheetsRepository) AppendRows(ctx context.Context, sheetName string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	valueRange := &sheets.ValueRange{
		Values: rows,
	}

	_, err := r.service.Spreadsheets.Values.Append(
		r.spreadsheetID,
		sheetName,
		valueRange,
	).ValueInputOption("USER_ENTERED").Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("failed to append rows: %w", err)
	}

	return nil
}

// ReadRange reads values from a specified range
// rangeNotation: A1 notation (e.g., "Sheet1!A1:E10")
func (r *SheetsRepository) ReadRange(ctx context.Context, rangeNotation string) ([][]interface{}, error) {
	resp, err := r.service.Spreadsheets.Values.Get(
		r.spreadsheetID,
		rangeNotation,
	).Context(ctx).Do()

	if err != nil {
		return nil, fmt.Errorf("failed to read range: %w", err)
	}

	return resp.Values, nil
}

// UpdateRange updates values in a specified range
// rangeNotation: A1 notation (e.g., "Sheet1!A1:E10")
// values: 2D array of values to update
func (r *SheetsRepository) UpdateRange(ctx context.Context, rangeNotation string, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	_, err := r.service.Spreadsheets.Values.Update(
		r.spreadsheetID,
		rangeNotation,
		valueRange,
	).ValueInputOption("USER_ENTERED").Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("failed to update range: %w", err)
	}

	return nil
}

// GetSpreadsheetInfo retrieves basic information about the spreadsheet
func (r *SheetsRepository) GetSpreadsheetInfo(ctx context.Context) (*sheets.Spreadsheet, error) {
	spreadsheet, err := r.service.Spreadsheets.Get(r.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet info: %w", err)
	}
	return spreadsheet, nil
}

// CreateSheet creates a new sheet (tab) in the spreadsheet
func (r *SheetsRepository) CreateSheet(ctx context.Context, sheetName string) error {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: sheetName,
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err := r.service.Spreadsheets.BatchUpdate(
		r.spreadsheetID,
		batchUpdateRequest,
	).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}

	return nil
}

// ClearRange clears values in a specified range
func (r *SheetsRepository) ClearRange(ctx context.Context, rangeNotation string) error {
	_, err := r.service.Spreadsheets.Values.Clear(
		r.spreadsheetID,
		rangeNotation,
		&sheets.ClearValuesRequest{},
	).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("failed to clear range: %w", err)
	}

	return nil
}

// Helper function to parse service account JSON from string
func ParseServiceAccountJSON(jsonString string) ([]byte, error) {
	// Validate it's valid JSON
	var temp map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &temp); err != nil {
		return nil, fmt.Errorf("invalid service account JSON: %w", err)
	}
	return []byte(jsonString), nil
}
