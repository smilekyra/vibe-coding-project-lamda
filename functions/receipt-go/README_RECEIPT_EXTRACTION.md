# Receipt Extraction Service

영수증 이미지에서 구조화된 데이터를 추출하는 서비스입니다. OpenAI Vision API와 Structured Outputs를 사용하여 영수증 정보를 JSON 형태로 정형화합니다.

## 주요 기능

- 영수증 이미지에서 텍스트 정보 자동 추출
- OpenAI GPT-4o Vision API를 사용한 이미지 분석
- Structured Outputs을 통한 JSON 스키마 보장
- 서비스 레이어 패턴으로 구현된 확장 가능한 구조

## 파일 구조

```
receipt-go/
├── types.go              # 영수증 데이터 구조 정의
├── openai_client.go      # OpenAI API 클라이언트
├── service.go            # 영수증 추출 서비스 레이어
├── example_usage.go      # 사용 예제
├── main.go              # Lambda 핸들러 (기존)
└── README_RECEIPT_EXTRACTION.md
```

## 데이터 구조

### ReceiptData
추출되는 영수증 정보:

```go
type ReceiptData struct {
    MerchantName     string        // 상점 이름
    MerchantAddress  string        // 상점 주소
    PhoneNumber      string        // 전화번호
    TransactionDate  string        // 거래 날짜 (YYYY-MM-DD)
    TransactionTime  string        // 거래 시간 (HH:MM:SS)
    Items            []ReceiptItem // 구매 항목 목록
    Subtotal         float64       // 소계
    Tax              float64       // 세금
    Total            float64       // 총액
    PaymentMethod    string        // 결제 수단
    CardLastFour     string        // 카드 마지막 4자리
    ReceiptNumber    string        // 영수증 번호
    CashierName      string        // 계산원 이름
}
```

### ReceiptItem
개별 구매 항목:

```go
type ReceiptItem struct {
    Name     string  // 상품명
    Quantity int     // 수량
    Price    float64 // 단가
    Total    float64 // 총액
}
```

## 사용 방법

### 1. 환경 변수 설정

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### 2. 기본 사용 예제

```go
package main

import (
    "context"
    "fmt"
    "os"
)

func main() {
    // 1. 이미지 파일 읽기
    imageData, err := os.ReadFile("receipt.jpg")
    if err != nil {
        panic(err)
    }

    // 2. OpenAI 클라이언트 생성
    openAIClient := NewOpenAIClient("")

    // 3. 영수증 추출 서비스 생성
    service := NewReceiptExtractionService(openAIClient)

    // 4. 영수증 데이터 추출
    ctx := context.Background()
    response, err := service.ExtractFromImage(ctx, imageData)
    if err != nil {
        panic(err)
    }

    // 5. 결과 확인
    if response.Success {
        fmt.Printf("상점: %s\n", response.Data.MerchantName)
        fmt.Printf("날짜: %s\n", response.Data.TransactionDate)
        fmt.Printf("총액: $%.2f\n", response.Data.Total)

        for _, item := range response.Data.Items {
            fmt.Printf("- %s: $%.2f\n", item.Name, item.Total)
        }
    } else {
        fmt.Printf("추출 실패: %s\n", response.Error)
    }
}
```

### 3. 기존 Lambda Handler와 통합

기존 Lambda 핸들러에 영수증 추출 기능을 추가하려면:

```go
// Handler 함수 내부에서 S3 업로드 후:

// OpenAI 클라이언트와 서비스 생성
openAIClient := NewOpenAIClient("")
extractionService := NewReceiptExtractionService(openAIClient)

// 영수증 데이터 추출
extractionResponse, err := extractionService.ExtractFromImage(ctx, decodedFile)
if err != nil {
    // 에러 로깅 (추출 실패해도 S3 업로드는 성공)
    fmt.Printf("Failed to extract receipt data: %v\n", err)
}

// 응답에 추출된 데이터 포함
type EnhancedResponse struct {
    // 기존 필드들
    FileName  string `json:"fileName"`
    FileSize  int64  `json:"fileSize"`
    S3Key     string `json:"s3Key"`
    S3Bucket  string `json:"s3Bucket"`
    Timestamp int64  `json:"timestamp"`

    // 새로운 필드들
    ExtractedData   *ReceiptData `json:"extractedData,omitempty"`
    ExtractionError string       `json:"extractionError,omitempty"`
}
```

## API 응답 예제

### 성공적인 추출

```json
{
  "success": true,
  "data": {
    "merchant_name": "ABC Supermarket",
    "merchant_address": "123 Main St, New York, NY 10001",
    "phone_number": "555-1234",
    "transaction_date": "2024-10-30",
    "transaction_time": "14:35:22",
    "items": [
      {
        "name": "Apple",
        "quantity": 3,
        "price": 1.99,
        "total": 5.97
      },
      {
        "name": "Bread",
        "quantity": 1,
        "price": 3.49,
        "total": 3.49
      }
    ],
    "subtotal": 9.46,
    "tax": 0.85,
    "total": 10.31,
    "payment_method": "CREDIT",
    "card_last_four": "4242",
    "receipt_number": "R-2024-001234",
    "cashier_name": "John Doe"
  }
}
```

### 추출 실패

```json
{
  "success": false,
  "error": "failed to extract receipt data: API request failed with status 401"
}
```

## 데이터 검증

서비스는 추출된 데이터의 유효성을 검증하는 기능을 제공합니다:

```go
validationErrors := ValidateReceiptData(response.Data)
if len(validationErrors) > 0 {
    for _, err := range validationErrors {
        fmt.Printf("Warning: %s\n", err)
    }
}
```

검증 항목:
- 상점명 존재 여부
- 총액 유효성 (0보다 큰 값)
- 구매 항목 존재 여부
- 소계 + 세금 = 총액 계산 검증 (±5센트 허용)

## 테스트

테스트 코드 작성 예제:

```go
// service_test.go
package main

import (
    "context"
    "testing"
)

type MockOpenAIClient struct{}

func (m *MockOpenAIClient) ExtractReceiptData(ctx context.Context, imageData []byte) (*ReceiptData, error) {
    return &ReceiptData{
        MerchantName: "Test Store",
        Total: 10.00,
        Items: []ReceiptItem{
            {Name: "Item 1", Quantity: 1, Price: 10.00, Total: 10.00},
        },
    }, nil
}

func TestExtractFromImage(t *testing.T) {
    mockClient := &MockOpenAIClient{}
    service := NewReceiptExtractionService(mockClient)

    response, err := service.ExtractFromImage(context.Background(), []byte("test"))
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    if !response.Success {
        t.Fatalf("Expected success, got failure")
    }

    if response.Data.MerchantName != "Test Store" {
        t.Errorf("Expected 'Test Store', got '%s'", response.Data.MerchantName)
    }
}
```

## 비용 고려사항

- GPT-4o Vision API는 사용량에 따라 과금됩니다
- 이미지 토큰 수는 이미지 크기와 해상도에 따라 달라집니다
- 고해상도 이미지를 사용하면 더 정확한 추출이 가능하지만 비용이 증가합니다
- 프로덕션 환경에서는 에러 처리와 재시도 로직 구현을 권장합니다

## 성능 최적화

1. **이미지 크기 최적화**: 영수증이 명확하게 보이는 최소 해상도 사용
2. **병렬 처리**: 여러 영수증을 동시에 처리할 때 고루틴 활용
3. **캐싱**: 같은 영수증의 중복 요청을 피하기 위한 캐싱 구현
4. **타임아웃 설정**: HTTP 클라이언트 타임아웃 적절히 설정 (현재 60초)

## 제한사항

- OpenAI API 호출 제한(rate limit)에 영향을 받을 수 있습니다
- 이미지 품질이 낮으면 추출 정확도가 떨어질 수 있습니다
- 손글씨 영수증은 인식률이 낮을 수 있습니다
- 매우 긴 영수증의 경우 토큰 제한으로 일부 정보가 누락될 수 있습니다

## 문제 해결

### API 키 오류
```
Error: failed to send request: 401 Unauthorized
```
해결: OPENAI_API_KEY 환경 변수가 올바르게 설정되었는지 확인

### 타임아웃 오류
```
Error: context deadline exceeded
```
해결: HTTP 클라이언트의 타임아웃을 늘리거나 이미지 크기를 줄이기

### 추출 데이터 불완전
해결:
- 이미지 품질 향상
- 프롬프트 개선
- JSON 스키마의 필수 필드를 조정 (일부 필드를 선택적으로 변경)

## 라이센스

이 프로젝트는 MIT 라이센스 하에 배포됩니다.
