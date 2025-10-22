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
		checkBody      bool
	}{
		{
			name: "Valid GET request",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "GET",
					},
				},
			},
			expectedStatus: 200,
			checkBody:      true,
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
			checkBody:      false,
		},
		{
			name: "Invalid PUT request",
			request: events.LambdaFunctionURLRequest{
				RequestContext: events.LambdaFunctionURLRequestContext{
					HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
						Method: "PUT",
					},
				},
			},
			expectedStatus: 405,
			checkBody:      false,
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

			if tt.checkBody {
				var resp TimeResponse
				err := json.Unmarshal([]byte(response.Body), &resp)
				if err != nil {
					t.Fatalf("Failed to unmarshal response body: %v", err)
				}

				if resp.Message != "Current time in Japan" {
					t.Errorf("Expected message 'Current time in Japan', got '%s'", resp.Message)
				}

				if resp.Timezone != "JST (Asia/Tokyo)" {
					t.Errorf("Expected timezone 'JST (Asia/Tokyo)', got '%s'", resp.Timezone)
				}

				if resp.CurrentTime == "" {
					t.Error("Expected current_time to be set")
				}
			}
		})
	}
}
