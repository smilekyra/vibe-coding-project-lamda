package openai

import (
	"encoding/base64"
	"fmt"
)

const (
	// OpenAI Vision API limits
	MaxImageSizeBytes       = 50 * 1024 * 1024 // 50 MB per image
	MaxImagesPerRequest     = 500              // Maximum images per request
	MaxTotalPayloadSize     = 50 * 1024 * 1024 // 50 MB total payload
)

// ImageValidationError represents an image validation error
type ImageValidationError struct {
	Field   string
	Message string
}

func (e *ImageValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateImageSize validates image size against OpenAI limits
// Returns error if image exceeds size limits
func ValidateImageSize(imageData []byte) error {
	size := len(imageData)
	
	if size == 0 {
		return &ImageValidationError{
			Field:   "image_data",
			Message: "image data is empty",
		}
	}
	
	if size > MaxImageSizeBytes {
		return &ImageValidationError{
			Field:   "image_size",
			Message: fmt.Sprintf("image size %d bytes (%.2f MB) exceeds OpenAI limit of %d bytes (%.2f MB)",
				size, float64(size)/(1024*1024),
				MaxImageSizeBytes, float64(MaxImageSizeBytes)/(1024*1024)),
		}
	}
	
	return nil
}

// ValidateImageSizeFromBase64 validates base64-encoded image size
func ValidateImageSizeFromBase64(base64Data string) error {
	if base64Data == "" {
		return &ImageValidationError{
			Field:   "image_data",
			Message: "base64 image data is empty",
		}
	}
	
	// Calculate decoded size without actually decoding
	// Base64 encoding increases size by ~33%, so divide by 1.33 to get original size
	// More accurate: use the formula (n * 3) / 4 where n is base64 length
	base64Len := len(base64Data)
	
	// Remove data URI prefix if present
	if len(base64Data) > 5 && base64Data[:5] == "data:" {
		// Find the comma that separates header from data
		for i, ch := range base64Data {
			if ch == ',' {
				base64Len = len(base64Data) - i - 1
				break
			}
		}
	}
	
	// Calculate approximate decoded size
	// Account for padding characters
	padding := 0
	if base64Len > 0 {
		if base64Data[len(base64Data)-1] == '=' {
			padding++
		}
		if base64Len > 1 && base64Data[len(base64Data)-2] == '=' {
			padding++
		}
	}
	
	decodedSize := (base64Len * 3 / 4) - padding
	
	if decodedSize > MaxImageSizeBytes {
		return &ImageValidationError{
			Field:   "image_size",
			Message: fmt.Sprintf("image size ~%d bytes (%.2f MB) exceeds OpenAI limit of %d bytes (%.2f MB)",
				decodedSize, float64(decodedSize)/(1024*1024),
				MaxImageSizeBytes, float64(MaxImageSizeBytes)/(1024*1024)),
		}
	}
	
	return nil
}

// ValidateImageFormat validates image format against OpenAI supported formats
func ValidateImageFormat(imageData []byte) error {
	if len(imageData) < 4 {
		return &ImageValidationError{
			Field:   "image_format",
			Message: "image data too short to determine format",
		}
	}
	
	// Check magic bytes for supported formats
	isPNG := imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47
	isJPEG := imageData[0] == 0xFF && imageData[1] == 0xD8 && imageData[2] == 0xFF
	isGIF := imageData[0] == 0x47 && imageData[1] == 0x49 && imageData[2] == 0x46
	isWEBP := len(imageData) >= 12 && 
		imageData[0] == 0x52 && imageData[1] == 0x49 && imageData[2] == 0x46 && imageData[3] == 0x46 &&
		imageData[8] == 0x57 && imageData[9] == 0x45 && imageData[10] == 0x42 && imageData[11] == 0x50
	
	if !isPNG && !isJPEG && !isGIF && !isWEBP {
		return &ImageValidationError{
			Field:   "image_format",
			Message: "unsupported image format. OpenAI supports: PNG, JPEG, WEBP, non-animated GIF",
		}
	}
	
	return nil
}

// ValidateImageForOpenAI performs complete validation for OpenAI Vision API
// Checks both size and format
func ValidateImageForOpenAI(imageData []byte) error {
	// Validate size
	if err := ValidateImageSize(imageData); err != nil {
		return err
	}
	
	// Validate format
	if err := ValidateImageFormat(imageData); err != nil {
		return err
	}
	
	return nil
}

// ValidateBase64ImageForOpenAI validates base64-encoded image for OpenAI
func ValidateBase64ImageForOpenAI(base64Data string) error {
	// Remove data URI prefix if present
	cleanBase64 := base64Data
	if len(base64Data) > 5 && base64Data[:5] == "data:" {
		// Find the comma that separates header from data
		for i, ch := range base64Data {
			if ch == ',' {
				cleanBase64 = base64Data[i+1:]
				break
			}
		}
	}
	
	// Validate size without full decoding
	if err := ValidateImageSizeFromBase64(base64Data); err != nil {
		return err
	}
	
	// Decode first few bytes to check format
	// Only decode what we need for magic bytes check
	decoded, err := base64.StdEncoding.DecodeString(cleanBase64[:min(len(cleanBase64), 20)])
	if err != nil {
		return &ImageValidationError{
			Field:   "image_data",
			Message: fmt.Sprintf("invalid base64 encoding: %v", err),
		}
	}
	
	// Validate format
	if err := ValidateImageFormat(decoded); err != nil {
		return err
	}
	
	return nil
}

// GetImageSizeInfo returns human-readable size information
func GetImageSizeInfo(imageData []byte) string {
	size := len(imageData)
	sizeMB := float64(size) / (1024 * 1024)
	maxMB := float64(MaxImageSizeBytes) / (1024 * 1024)
	percentage := (float64(size) / float64(MaxImageSizeBytes)) * 100
	
	return fmt.Sprintf("Size: %.2f MB / %.0f MB (%.1f%% of limit)", sizeMB, maxMB, percentage)
}

// GetImageFormatInfo returns human-readable format information
func GetImageFormatInfo(imageData []byte) string {
	if len(imageData) < 4 {
		return "Unknown format"
	}
	
	switch {
	case imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47:
		return "PNG"
	case imageData[0] == 0xFF && imageData[1] == 0xD8 && imageData[2] == 0xFF:
		return "JPEG"
	case imageData[0] == 0x47 && imageData[1] == 0x49 && imageData[2] == 0x46:
		return "GIF"
	case len(imageData) >= 12 && 
		imageData[0] == 0x52 && imageData[1] == 0x49 && imageData[2] == 0x46 && imageData[3] == 0x46 &&
		imageData[8] == 0x57 && imageData[9] == 0x45 && imageData[10] == 0x42 && imageData[11] == 0x50:
		return "WEBP"
	default:
		return "Unknown/Unsupported"
	}
}

