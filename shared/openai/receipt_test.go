package openai

import (
	"context"
	"os"
	"testing"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name    string
		config  ServiceConfig
		wantErr bool
	}{
		{
			name: "with API key in config",
			config: ServiceConfig{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "with custom settings",
			config: ServiceConfig{
				APIKey:          "test-api-key",
				DefaultCurrency: "JPY",
				DefaultLanguage: "ja",
				VisionModel:     "gpt-4o",
				MaxTokens:       2000,
				Temperature:     0.2,
			},
			wantErr: false,
		},
		{
			name:    "without API key",
			config:  ServiceConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable for testing
			oldKey := os.Getenv("OPENAI_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")
			defer os.Setenv("OPENAI_API_KEY", oldKey)

			service, err := NewService(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if service == nil {
					t.Error("Expected service to be created, got nil")
				}

				// Verify defaults are set
				config := service.GetConfig()
				if config.VisionModel == "" {
					t.Error("Expected VisionModel to have default value")
				}
				if config.MaxTokens == 0 {
					t.Error("Expected MaxTokens to have default value")
				}
			}
		})
	}
}

func TestServiceConfig(t *testing.T) {
	service, err := NewService(ServiceConfig{
		APIKey:          "test-key",
		DefaultCurrency: "EUR",
		DefaultLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test GetConfig
	config := service.GetConfig()
	if config.DefaultCurrency != "EUR" {
		t.Errorf("Expected currency EUR, got %s", config.DefaultCurrency)
	}

	// Test UpdateConfig
	service.UpdateConfig(ServiceConfig{
		DefaultCurrency: "GBP",
		MaxTokens:       3000,
	})

	updatedConfig := service.GetConfig()
	if updatedConfig.DefaultCurrency != "GBP" {
		t.Errorf("Expected currency GBP after update, got %s", updatedConfig.DefaultCurrency)
	}
	if updatedConfig.MaxTokens != 3000 {
		t.Errorf("Expected MaxTokens 3000 after update, got %d", updatedConfig.MaxTokens)
	}
}

func TestBuildReceiptExtractionPrompt(t *testing.T) {
	service, err := NewService(ServiceConfig{
		APIKey:          "test-key",
		DefaultCurrency: "USD",
		DefaultLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	tests := []struct {
		name string
		req  ReceiptExtractionRequest
		want []string // Strings that should be in the prompt
	}{
		{
			name: "basic request",
			req: ReceiptExtractionRequest{
				ImageData: "base64data",
			},
			want: []string{"JSON", "receipt", "USD", "en"},
		},
		{
			name: "with custom currency",
			req: ReceiptExtractionRequest{
				ImageData:        "base64data",
				ExpectedCurrency: "JPY",
			},
			want: []string{"JPY"},
		},
		{
			name: "with store hint",
			req: ReceiptExtractionRequest{
				ImageData: "base64data",
				StoreHint: "Walmart",
			},
			want: []string{"Walmart", "Additional context"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := service.buildReceiptExtractionPrompt(tt.req)
			for _, want := range tt.want {
				if len(prompt) > 0 && !contains(prompt, want) {
					t.Errorf("Prompt does not contain expected string: %s", want)
				}
			}
		})
	}
}

func TestEncodeImageToBase64(t *testing.T) {
	testData := []byte("test image data")
	encoded := EncodeImageToBase64(testData)

	if encoded == "" {
		t.Error("Expected non-empty base64 string")
	}
}

func TestPrepareImageDataURI(t *testing.T) {
	tests := []struct {
		name     string
		base64   string
		mimeType string
		want     string
	}{
		{
			name:     "with jpeg mime type",
			base64:   "abc123",
			mimeType: "image/jpeg",
			want:     "data:image/jpeg;base64,abc123",
		},
		{
			name:     "with default mime type",
			base64:   "abc123",
			mimeType: "",
			want:     "data:image/jpeg;base64,abc123",
		},
		{
			name:     "with png mime type",
			base64:   "xyz789",
			mimeType: "image/png",
			want:     "data:image/png;base64,xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareImageDataURI(tt.base64, tt.mimeType)
			if got != tt.want {
				t.Errorf("PrepareImageDataURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractReceiptDataValidation(t *testing.T) {
	service, err := NewService(ServiceConfig{
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()

	// Test with no image data
	resp, err := service.ExtractReceiptData(ctx, ReceiptExtractionRequest{})
	if err == nil {
		t.Error("Expected error when no image data provided")
	}
	if resp == nil || resp.Success {
		t.Error("Expected unsuccessful response")
	}
}

// Helper function
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
