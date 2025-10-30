#!/bin/bash

# Lambda Function URL
LAMBDA_URL="https://c27whwi7rosvpiyg2pxngptiye0eklkj.lambda-url.us-east-1.on.aws/"

# Receipt image path
RECEIPT_IMAGE="./test.png"

echo "=========================================="
echo "영수증 업로드 테스트"
echo "=========================================="
echo ""

# Check if file exists
if [ ! -f "$RECEIPT_IMAGE" ]; then
  echo "❌ 오류: $RECEIPT_IMAGE 파일을 찾을 수 없습니다"
  exit 1
fi

echo "📄 파일: $RECEIPT_IMAGE"
echo "📊 파일 크기: $(ls -lh "$RECEIPT_IMAGE" | awk '{print $5}')"
echo "🌐 업로드 URL: $LAMBDA_URL"
echo ""
echo "업로드 중..."
echo ""

# Create temp file for JSON payload
TEMP_FILE=$(mktemp)

# Convert image to base64 (remove newlines)
BASE64_CONTENT=$(base64 -i "$RECEIPT_IMAGE" | tr -d '\n')

# Create JSON payload
cat > "$TEMP_FILE" <<EOF
{
  "file": "$BASE64_CONTENT",
  "fileName": "muji-receipt.png"
}
EOF

# Upload with curl
RESPONSE=$(curl -X POST "$LAMBDA_URL" \
  -H "Content-Type: application/json" \
  -d @"$TEMP_FILE" \
  -w "\n%{http_code}" \
  -s)

# Extract HTTP status code (last line)
HTTP_CODE=$(echo "$RESPONSE" | tail -n 1)
# Extract response body (all except last line)
BODY=$(echo "$RESPONSE" | sed '$d')

# Cleanup temp file
rm "$TEMP_FILE"

echo "=========================================="
echo "📥 응답 결과"
echo "=========================================="
echo "HTTP 상태 코드: $HTTP_CODE"
echo ""
echo "응답 본문:"
echo "$BODY" | jq . 2>/dev/null || echo "$BODY"
echo ""

# Check status
if [ "$HTTP_CODE" -eq 200 ]; then
  echo "✅ 업로드 성공!"
elif [ "$HTTP_CODE" -eq 502 ]; then
  echo "⚠️  502 Bad Gateway - Lambda 타임아웃 또는 메모리 부족"
  echo ""
  echo "해결 방법:"
  echo "1. Lambda 타임아웃 설정을 30초 이상으로 증가"
  echo "2. Lambda 메모리를 512MB 이상으로 증가"
  echo "3. Lambda IAM 역할에 S3 PutObject 권한 확인"
else
  echo "❌ 업로드 실패 (HTTP $HTTP_CODE)"
fi

echo ""
echo "=========================================="
