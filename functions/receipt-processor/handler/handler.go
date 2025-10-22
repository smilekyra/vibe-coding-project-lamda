package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"vibe-coding-project-lambda/functions/receipt-processor/service"
	"vibe-coding-project-lambda/shared/repository"

	"github.com/aws/aws-lambda-go/events"
)

// ReceiptHandler handles Lambda requests for receipt processing
type ReceiptHandler struct {
	receiptService *service.ReceiptService
	sheetsService  *service.SheetsService
}

// NewReceiptHandler creates a new receipt handler
func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// SetSheetsService sets the Google Sheets service (optional)
func (h *ReceiptHandler) SetSheetsService(sheetsService *service.SheetsService) {
	h.sheetsService = sheetsService
}

// Handle handles the Lambda function invocation
func (h *ReceiptHandler) Handle(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	timestamp := time.Now().Unix()

	// Handle CORS preflight requests
	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		}, nil
	}

	// Only accept POST method
	if request.RequestContext.HTTP.Method != "POST" {
		return h.errorResponse(405, "Method not allowed. Only POST is supported.", "Invalid HTTP method", timestamp)
	}

	// Parse request and extract file data
	fileName, fileContent, contentType, err := h.parseRequest(request)
	if err != nil {
		return h.errorResponse(400, err.Error(), "Failed to parse request", timestamp)
	}

	// Validate we have file content
	if len(fileContent) == 0 {
		return h.errorResponse(400, "File content is empty", "Validation error", timestamp)
	}

	// Process receipt (upload + OCR)
	result, err := h.receiptService.ProcessReceipt(ctx, fileName, fileContent, contentType)
	if err != nil {
		return h.errorResponse(500, "Failed to process receipt", err.Error(), timestamp)
	}

	// Add to Google Sheets if available and receipt was processed
	if h.sheetsService != nil && result.ReceiptData != nil {
		memo := "" // Optional memo field - could be extracted from request if needed
		if err := h.sheetsService.AddReceiptToSpreadsheet(ctx, result.ReceiptData, result.FileInfo.URL, memo); err != nil {
			// Log error but don't fail the request
			// The receipt has already been uploaded to S3 and processed
			// Sheets sync is a nice-to-have feature
			_ = err // Ignore error for now
		}
	}

	// Build success response
	message := "File uploaded successfully"
	if result.ReceiptData != nil {
		message = "File uploaded and receipt processed successfully"
	}

	response := UploadResponse{
		Success: true,
		Message: message,
		FileInfo: &FileInfo{
			OriginalName: result.FileInfo.OriginalName,
			FileName:     result.FileInfo.FileName,
			BucketName:   result.FileInfo.BucketName,
			Key:          result.FileInfo.Key,
			Size:         result.FileInfo.Size,
			ContentType:  result.FileInfo.ContentType,
			URL:          result.FileInfo.URL,
			UploadDate:   result.FileInfo.UploadDate,
		},
		ReceiptData: result.ReceiptData,
		Timestamp:   timestamp,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return h.errorResponse(500, "Failed to generate response", err.Error(), timestamp)
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(responseBody),
	}, nil
}

// parseRequest parses the request body (multipart or JSON)
func (h *ReceiptHandler) parseRequest(request events.LambdaFunctionURLRequest) (fileName string, fileContent []byte, contentType string, err error) {
	// Determine content type
	requestContentType := request.Headers["content-type"]
	if requestContentType == "" {
		requestContentType = request.Headers["Content-Type"]
	}

	// Check if it's multipart/form-data
	if strings.HasPrefix(requestContentType, "multipart/form-data") {
		return parseMultipartRequest(request.Body, requestContentType)
	}

	// Parse as JSON (backward compatibility)
	var uploadReq UploadRequest
	if err := json.Unmarshal([]byte(request.Body), &uploadReq); err != nil {
		return "", nil, "", err
	}

	// Validate required fields
	if uploadReq.FileName == "" || uploadReq.FileContent == "" {
		return "", nil, "", err
	}

	// Set default content type if not provided
	if uploadReq.ContentType == "" {
		uploadReq.ContentType = "application/octet-stream"
	}

	// Decode base64 file content
	fileContent, err = base64.StdEncoding.DecodeString(uploadReq.FileContent)
	if err != nil {
		return "", nil, "", err
	}

	return uploadReq.FileName, fileContent, uploadReq.ContentType, nil
}

// errorResponse creates an error response
func (h *ReceiptHandler) errorResponse(statusCode int, message, errorDetail string, timestamp int64) (events.LambdaFunctionURLResponse, error) {
	response := UploadResponse{
		Success:   false,
		Message:   message,
		Error:     errorDetail,
		Timestamp: timestamp,
	}

	responseBody, _ := json.Marshal(response)

	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}
	if statusCode == 405 {
		headers["Allow"] = "POST"
	}

	return events.LambdaFunctionURLResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

// Helper to convert repository.FileInfo to handler.FileInfo
func toHandlerFileInfo(repoInfo *repository.FileInfo) *FileInfo {
	if repoInfo == nil {
		return nil
	}
	return &FileInfo{
		OriginalName: repoInfo.OriginalName,
		FileName:     repoInfo.FileName,
		BucketName:   repoInfo.BucketName,
		Key:          repoInfo.Key,
		Size:         repoInfo.Size,
		ContentType:  repoInfo.ContentType,
		URL:          repoInfo.URL,
		UploadDate:   repoInfo.UploadDate,
	}
}
