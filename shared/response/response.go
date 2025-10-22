package response

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// Success creates a successful API Gateway response
func Success(statusCode int, body interface{}) (events.APIGatewayProxyResponse, error) {
	responseBody, err := json.Marshal(body)
	if err != nil {
		return Error(500, fmt.Sprintf("Failed to marshal response: %s", err.Error()))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(responseBody),
	}, nil
}

// Error creates an error API Gateway response
func Error(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	errorResponse := map[string]string{
		"error": message,
	}

	responseBody, _ := json.Marshal(errorResponse)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(responseBody),
	}, nil
}

// MethodNotAllowed creates a 405 Method Not Allowed response
func MethodNotAllowed(allowedMethods string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 405,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"Allow": allowedMethods,
		},
		Body: `{"error": "Method not allowed"}`,
	}, nil
}

