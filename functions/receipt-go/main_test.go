package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        events.LambdaFunctionURLRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name: "Valid file upload",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Body: `{"file":"` + base64.StdEncoding.EncodeToString([]byte("Hello World")) + `","fileName":"test.txt"}`,
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				var response ReceiptResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.FileName != "test.txt" {
					t.Errorf("Expected fileName 'test.txt', got '%s'", response.FileName)
				}
				if response.FileSize != 11 {
					t.Errorf("Expected fileSize 11, got %d", response.FileSize)
				}
			},
		},
		{
			name: "GET method not allowed",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
					},
				},
			},
			expectedStatus: 405,
			checkResponse: func(t *testing.T, body string) {
				var response ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
			},
		},
		{
			name: "Invalid JSON body",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Body: `invalid json`,
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body string) {
				var response ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
			},
		},
		{
			name: "Missing file field",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Body: `{"fileName":"test.txt"}`,
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body string) {
				var response ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
			},
		},
		{
			name: "Invalid base64 encoding",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Body: `{"file":"invalid-base64!!!","fileName":"test.txt"}`,
			},
			expectedStatus: 400,
			checkResponse: func(t *testing.T, body string) {
				var response ErrorResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse error response: %v", err)
				}
			},
		},
		{
			name: "File upload without fileName",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
				Body: `{"file":"` + base64.StdEncoding.EncodeToString([]byte("Test content")) + `"}`,
			},
			expectedStatus: 200,
			checkResponse: func(t *testing.T, body string) {
				var response ReceiptResponse
				if err := json.Unmarshal([]byte(body), &response); err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.FileName != "unknown" {
					t.Errorf("Expected fileName 'unknown', got '%s'", response.FileName)
				}
				if response.FileSize != 12 {
					t.Errorf("Expected fileSize 12, got %d", response.FileSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(context.Background(), tt.request)
			if err != nil {
				t.Errorf("Handler returned error: %v", err)
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response.Body)
			}
		})
	}
}
