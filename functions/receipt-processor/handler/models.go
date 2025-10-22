package handler

import "vibe-coding-project-lambda/shared/openai"

// UploadRequest represents the file upload request structure
type UploadRequest struct {
	FileName    string `json:"filename"`
	FileContent string `json:"file_content"` // Base64 encoded file content
	ContentType string `json:"content_type"`
}

// UploadResponse represents the API response structure
type UploadResponse struct {
	Success     bool                `json:"success"`
	Message     string              `json:"message"`
	FileInfo    *FileInfo           `json:"file_info,omitempty"`
	ReceiptData *openai.ReceiptData `json:"receipt_data,omitempty"`
	Error       string              `json:"error,omitempty"`
	Timestamp   int64               `json:"timestamp"`
}

// FileInfo contains information about the uploaded file
type FileInfo struct {
	OriginalName string `json:"original_name"`
	FileName     string `json:"file_name"`
	BucketName   string `json:"bucket_name"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	URL          string `json:"url"`
	UploadDate   string `json:"upload_date"`
}
