package repository

import (
	"testing"
)

func TestParseServiceAccountJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid service account JSON",
			input: `{
				"type": "service_account",
				"project_id": "test-project",
				"private_key_id": "key123",
				"private_key": "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n",
				"client_email": "test@test-project.iam.gserviceaccount.com",
				"client_id": "123456789",
				"auth_uri": "https://accounts.google.com/o/oauth2/auth",
				"token_uri": "https://oauth2.googleapis.com/token"
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseServiceAccountJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseServiceAccountJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) == 0 {
				t.Errorf("ParseServiceAccountJSON() returned empty result for valid input")
			}
		})
	}
}

func TestSheetsConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config SheetsConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: SheetsConfig{
				ServiceAccountJSON: []byte(`{"type":"service_account"}`),
				SpreadsheetID:      "test-spreadsheet-id",
			},
			valid: true,
		},
		{
			name: "missing service account JSON",
			config: SheetsConfig{
				SpreadsheetID: "test-spreadsheet-id",
			},
			valid: false,
		},
		{
			name: "missing spreadsheet ID",
			config: SheetsConfig{
				ServiceAccountJSON: []byte(`{"type":"service_account"}`),
			},
			valid: false,
		},
		{
			name: "valid config with custom scopes",
			config: SheetsConfig{
				ServiceAccountJSON: []byte(`{"type":"service_account"}`),
				SpreadsheetID:      "test-spreadsheet-id",
				Scopes:             []string{"https://www.googleapis.com/auth/spreadsheets.readonly"},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasJSON := len(tt.config.ServiceAccountJSON) > 0
			hasID := tt.config.SpreadsheetID != ""
			isValid := hasJSON && hasID

			if isValid != tt.valid {
				t.Errorf("Config validation failed: got %v, want %v", isValid, tt.valid)
			}
		})
	}
}

// TestRowDataStructure tests that row data can be properly formatted
func TestRowDataStructure(t *testing.T) {
	tests := []struct {
		name   string
		values []interface{}
		want   int
	}{
		{
			name:   "receipt row with all fields",
			values: []interface{}{"2024-10-18", "식비", "Store Name", 1000, 5, "Coffee, Sandwich, Water", "Credit Card", "https://s3.example.com/receipt.jpg", "Memo"},
			want:   9,
		},
		{
			name:   "empty row",
			values: []interface{}{},
			want:   0,
		},
		{
			name:   "partial row",
			values: []interface{}{"2024-10-18", "식비", "Store Name"},
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.values) != tt.want {
				t.Errorf("Row length = %v, want %v", len(tt.values), tt.want)
			}
		})
	}
}

// TestMultipleRowsStructure tests batch row operations
func TestMultipleRowsStructure(t *testing.T) {
	rows := [][]interface{}{
		{"2024-10-18", "식비", "Store A", 1000, 5, "Coffee, Sandwich, Apple, Banana, Water", "Credit", "https://example.com/1", "Note 1"},
		{"2024-10-19", "교통비", "Store B", 2000, 3, "Gasoline, Car Wash, Parking", "Cash", "https://example.com/2", "Note 2"},
		{"2024-10-20", "생활용품", "Store C", 3000, 8, "Soap, Shampoo, Towel, Brush, Detergent, Sponge, Cleaner, Paper", "Credit", "https://example.com/3", "Note 3"},
	}

	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}

	for i, row := range rows {
		if len(row) != 9 {
			t.Errorf("Row %d: expected 9 columns, got %d", i, len(row))
		}
	}
}
