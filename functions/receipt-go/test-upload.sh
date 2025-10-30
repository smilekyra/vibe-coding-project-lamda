#!/bin/bash

# Lambda Function URL
LAMBDA_URL="https://c27whwi7rosvpiyg2pxngptiye0eklkj.lambda-url.us-east-1.on.aws/"

echo "=========================================="
echo "Testing Receipt Upload Lambda Function"
echo "=========================================="
echo ""

# Test 1: Valid file upload
echo "Test 1: Valid file upload with test.txt"
echo "------------------------------------------"

# Create a test file content and encode to base64
TEST_CONTENT="Hello World! This is a test receipt file."
BASE64_CONTENT=$(echo -n "$TEST_CONTENT" | base64)

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$BASE64_CONTENT\",
    \"fileName\": \"test-receipt.txt\"
  }" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 2: Upload with image file (simulated)
echo "Test 2: Valid file upload with image"
echo "------------------------------------------"

# Create a small test image (1x1 pixel PNG in base64)
PNG_BASE64="iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$PNG_BASE64\",
    \"fileName\": \"test-image.png\"
  }" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 3: Missing file field (should fail)
echo "Test 3: Missing file field (should return 400)"
echo "------------------------------------------"

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"fileName\": \"test.txt\"
  }" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 4: Invalid JSON (should fail)
echo "Test 4: Invalid JSON (should return 400)"
echo "------------------------------------------"

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "invalid json" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 5: GET method (should fail)
echo "Test 5: GET method (should return 405)"
echo "------------------------------------------"

curl -X GET "$LAMBDA_URL" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 6: File without fileName (should use 'unknown')
echo "Test 6: File upload without fileName"
echo "------------------------------------------"

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"$BASE64_CONTENT\"
  }" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo ""

# Test 7: Invalid base64 (should fail)
echo "Test 7: Invalid base64 encoding (should return 400)"
echo "------------------------------------------"

curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"file\": \"invalid-base64-content!!!\",
    \"fileName\": \"test.txt\"
  }" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo "=========================================="
echo "All tests completed!"
echo "=========================================="
