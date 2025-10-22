package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
)

// parseMultipartRequest parses a multipart/form-data request
func parseMultipartRequest(body string, contentType string) (fileName string, fileContent []byte, fileContentType string, err error) {
	// Parse the content type to get the boundary
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to parse content type: %w", err)
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "", nil, "", fmt.Errorf("boundary not found in content type")
	}

	// Lambda Function URLs with base64 encoding enabled
	var bodyBytes []byte
	if strings.Contains(body, "base64") || len(body) > 0 {
		// Try to decode if it's base64
		decoded, decodeErr := base64.StdEncoding.DecodeString(body)
		if decodeErr == nil {
			bodyBytes = decoded
		} else {
			// Not base64, use as is
			bodyBytes = []byte(body)
		}
	} else {
		bodyBytes = []byte(body)
	}

	// Create a multipart reader
	reader := multipart.NewReader(bytes.NewReader(bodyBytes), boundary)

	// Read the file part
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", nil, "", fmt.Errorf("failed to read part: %w", err)
		}

		// Check if this is the file field
		if part.FormName() == "file" {
			fileName = part.FileName()
			fileContentType = part.Header.Get("Content-Type")

			// Read the file content
			fileContent, err = io.ReadAll(part)
			if err != nil {
				return "", nil, "", fmt.Errorf("failed to read file content: %w", err)
			}

			part.Close()
			break
		}

		part.Close()
	}

	if fileName == "" {
		return "", nil, "", fmt.Errorf("no file found in request (looking for 'file' field)")
	}

	return fileName, fileContent, fileContentType, nil
}
