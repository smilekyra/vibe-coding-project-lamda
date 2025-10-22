package repository

import (
	"testing"
	"time"
)

func TestGetJSTDateFolder(t *testing.T) {
	result := getJSTDateFolder()

	// Should be in format YYYY-MM-DD
	if len(result) != 10 {
		t.Errorf("Expected date format YYYY-MM-DD, got %s", result)
	}

	// Should contain hyphens at correct positions
	if result[4] != '-' || result[7] != '-' {
		t.Errorf("Expected date format YYYY-MM-DD, got %s", result)
	}
}

func TestGenerateUniqueFileName(t *testing.T) {
	tests := []struct {
		name         string
		originalName string
		wantContains []string
	}{
		{
			name:         "Simple filename with extension",
			originalName: "receipt.jpg",
			wantContains: []string{"receipt_", ".jpg"},
		},
		{
			name:         "Filename without extension",
			originalName: "receipt",
			wantContains: []string{"receipt_"},
		},
		{
			name:         "Filename with multiple dots",
			originalName: "my.receipt.image.png",
			wantContains: []string{"my.receipt.image_", ".png"},
		},
		{
			name:         "Filename with spaces",
			originalName: "my receipt.jpg",
			wantContains: []string{"my receipt_", ".jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateUniqueFileName(tt.originalName)

			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf("generateUniqueFileName() = %v, should contain %v", result, want)
				}
			}

			// Should have timestamp (8 digits + underscore + 6 digits)
			// Format: basename_YYYYMMDD_HHMMSS_randomhex.ext
			if len(result) < 20 {
				t.Errorf("Generated filename too short: %s", result)
			}
		})
	}
}

func TestGenerateUniqueFileName_Uniqueness(t *testing.T) {
	// Generate multiple filenames in quick succession
	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateUniqueFileName("test.jpg")
		if names[name] {
			t.Errorf("Generated duplicate filename: %s", name)
		}
		names[name] = true
		time.Sleep(time.Millisecond)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
