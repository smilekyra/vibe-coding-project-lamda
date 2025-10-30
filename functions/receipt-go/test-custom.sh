#!/bin/bash

# Lambda Function URL
LAMBDA_URL="https://c27whwi7rosvpiyg2pxngptiye0eklkj.lambda-url.us-east-1.on.aws/"

# Path to the receipt image
RECEIPT_IMAGE="/Users/kyra/Documents/vibe-coding-project/test/test.png"

echo "=========================================="
echo "Testing Receipt Upload"
echo "=========================================="
echo ""

if [ ! -f "$RECEIPT_IMAGE" ]; then
  echo "Error: Receipt image not found at $RECEIPT_IMAGE"
  exit 1
fi

echo "Uploading receipt: $RECEIPT_IMAGE"
echo "File size: $(ls -lh "$RECEIPT_IMAGE" | awk '{print $5}')"
echo "------------------------------------------"

# Convert image to base64 and create JSON payload in a temp file
TEMP_FILE=$(mktemp)
BASE64_CONTENT=$(base64 -i "$RECEIPT_IMAGE" | tr -d '\n')

# Create JSON payload
cat > "$TEMP_FILE" <<PAYLOAD
{
  "file": "$BASE64_CONTENT",
  "fileName": "test-receipt.png"
}
PAYLOAD

# Upload the receipt
echo "Response:"
curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d @"$TEMP_FILE" \
  -w "\n\nHTTP Status: %{http_code}\n" \
  2>&1 | grep -E '(HTTP|fileName|error|Error)'

# Cleanup
rm "$TEMP_FILE"

echo ""
echo "=========================================="
echo "Upload test completed!"
echo "=========================================="
