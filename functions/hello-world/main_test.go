package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        events.LambdaFunctionURLRequest
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "Valid GET request without name",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
					},
				},
			},
			expectedStatus: 200,
			expectedMsg:    "Hello, World!",
		},
		{
			name: "Valid GET request with name",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
					},
				},
				QueryStringParameters: map[string]string{
					"name": "Alice",
				},
			},
			expectedStatus: 200,
			expectedMsg:    "Hello, Alice!",
		},
		{
			name: "Invalid POST request",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "POST",
					},
				},
			},
			expectedStatus: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(context.Background(), tt.request)
			if err != nil {
				t.Fatalf("Handler returned an error: %v", err)
			}

			if response.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, response.StatusCode)
			}

			if tt.expectedStatus == 200 {
				var resp HelloResponse
				err := json.Unmarshal([]byte(response.Body), &resp)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}

				if resp.Message != tt.expectedMsg {
					t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, resp.Message)
				}
			}
		})
	}
}
