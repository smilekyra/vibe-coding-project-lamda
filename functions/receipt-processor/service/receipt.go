package service

import (
	"context"
	"log"

	"vibe-coding-project-lambda/shared/openai"
	"vibe-coding-project-lambda/shared/repository"
)

// ReceiptService handles receipt processing business logic
type ReceiptService struct {
	s3Repo        *repository.S3Repository
	openaiService *openai.Service
}

// NewReceiptService creates a new receipt service
func NewReceiptService(s3Repo *repository.S3Repository, openaiService *openai.Service) *ReceiptService {
	return &ReceiptService{
		s3Repo:        s3Repo,
		openaiService: openaiService,
	}
}

// ProcessResult contains the result of receipt processing
type ProcessResult struct {
	FileInfo    *repository.FileInfo
	ReceiptData *openai.ReceiptData
}

// ProcessReceipt processes a receipt: uploads to S3 and extracts data with OpenAI
func (s *ReceiptService) ProcessReceipt(ctx context.Context, fileName string, fileContent []byte, contentType string) (*ProcessResult, error) {
	// Upload to S3 first (always succeeds or fails hard)
	fileInfo, err := s.s3Repo.Upload(ctx, fileName, fileContent, contentType)
	if err != nil {
		return nil, err
	}

	result := &ProcessResult{
		FileInfo: fileInfo,
	}

	// Process with OpenAI if it's an image and service is available
	if s.openaiService != nil && isImageFile(contentType) {
		log.Printf("Processing receipt image with OpenAI")

		// Validate image first
		if err := openai.ValidateImageForOpenAI(fileContent); err != nil {
			log.Printf("Warning: Image validation failed: %v", err)
			log.Printf("Image info: Format=%s, %s",
				openai.GetImageFormatInfo(fileContent),
				openai.GetImageSizeInfo(fileContent))
		} else {
			log.Printf("Image validation passed: Format=%s, %s",
				openai.GetImageFormatInfo(fileContent),
				openai.GetImageSizeInfo(fileContent))

			// Process with OpenAI
			base64Image := openai.EncodeImageToBase64(fileContent)
			receiptData, err := s.openaiService.ProcessReceiptFromBase64(ctx, base64Image)
			if err != nil {
				log.Printf("Warning: Failed to process receipt with OpenAI: %v", err)
			} else {
				log.Printf("Successfully processed receipt: %s", receiptData.Summary())
				result.ReceiptData = receiptData
			}
		}
	}

	return result, nil
}

// isImageFile checks if the content type is an image
func isImageFile(contentType string) bool {
	imageTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/bmp",
	}
	for _, imgType := range imageTypes {
		if contentType == imgType {
			return true
		}
	}
	return false
}
