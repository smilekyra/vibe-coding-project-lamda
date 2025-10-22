package openai

import (
	"context"
	"fmt"
	"os"
)

// ServiceConfig holds configuration for the OpenAI service
type ServiceConfig struct {
	APIKey string

	// Context-specific settings for receipt processing
	DefaultCurrency string
	DefaultLanguage string
	DefaultTimezone string

	// Model configuration
	VisionModel     string
	CompletionModel string
	MaxTokens       int
	Temperature     float32
}

// Service provides methods to interact with OpenAI API
type Service struct {
	config ServiceConfig
	apiKey string
}

// NewService creates a new OpenAI service instance
func NewService(config ServiceConfig) (*Service, error) {
	// Use provided API key or fall back to environment variable
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required. Set OPENAI_API_KEY environment variable or provide it in config")
	}

	// Set defaults if not provided
	if config.VisionModel == "" {
		config.VisionModel = "gpt-4o" // Latest vision model
	}
	if config.CompletionModel == "" {
		config.CompletionModel = "gpt-4o"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Temperature == 0 {
		config.Temperature = 0.1 // Low temperature for consistent extraction
	}
	if config.DefaultCurrency == "" {
		config.DefaultCurrency = "USD"
	}
	if config.DefaultLanguage == "" {
		config.DefaultLanguage = "en"
	}
	if config.DefaultTimezone == "" {
		config.DefaultTimezone = "UTC"
	}

	return &Service{
		config: config,
		apiKey: apiKey,
	}, nil
}

// GetConfig returns the service configuration
func (s *Service) GetConfig() ServiceConfig {
	return s.config
}

// UpdateConfig updates the service configuration
func (s *Service) UpdateConfig(config ServiceConfig) {
	if config.APIKey != "" {
		s.apiKey = config.APIKey
	}
	if config.VisionModel != "" {
		s.config.VisionModel = config.VisionModel
	}
	if config.CompletionModel != "" {
		s.config.CompletionModel = config.CompletionModel
	}
	if config.MaxTokens > 0 {
		s.config.MaxTokens = config.MaxTokens
	}
	if config.Temperature >= 0 {
		s.config.Temperature = config.Temperature
	}
	if config.DefaultCurrency != "" {
		s.config.DefaultCurrency = config.DefaultCurrency
	}
	if config.DefaultLanguage != "" {
		s.config.DefaultLanguage = config.DefaultLanguage
	}
	if config.DefaultTimezone != "" {
		s.config.DefaultTimezone = config.DefaultTimezone
	}
}

// ValidateConnection checks if the API key is valid by making a simple API call
func (s *Service) ValidateConnection(ctx context.Context) error {
	// This is a placeholder - we'll implement actual validation when we add the API calls
	if s.apiKey == "" {
		return fmt.Errorf("API key is not set")
	}
	return nil
}
