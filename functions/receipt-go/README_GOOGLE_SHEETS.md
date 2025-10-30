# Google Sheets Integration

이 문서는 영수증 데이터를 Google Sheets에 자동으로 저장하는 기능에 대한 설명입니다.

## 기능 개요

영수증 이미지를 업로드하면:
1. S3에 이미지가 저장됩니다
2. OpenAI Vision API를 사용하여 영수증 데이터를 추출합니다
3. 추출된 데이터를 Google Sheets에 자동으로 저장합니다

## Google Sheets 형식

저장되는 데이터는 다음과 같은 컬럼으로 구성됩니다:

| 날짜 | 카테고리 | 상점명 | 총금액 | 항목수 | 항목내역 | 결제방법 | 영수증링크 | 메모 |
|------|----------|--------|--------|--------|----------|----------|------------|------|
| 2025-10-30 | | Starbucks | 15.50 | 2 | Americano (1개) x $4.50, Latte (1개) x $5.00 | Card ****1234 | https://... | |

## 설정 방법

### 1. Google Cloud Console 설정

1. [Google Cloud Console](https://console.cloud.google.com/)에 접속합니다
2. 새 프로젝트를 생성하거나 기존 프로젝트를 선택합니다
3. Google Sheets API를 활성화합니다:
   - "API 및 서비스" > "라이브러리"로 이동
   - "Google Sheets API"를 검색하여 활성화

### 2. 서비스 계정 생성

1. "API 및 서비스" > "사용자 인증 정보"로 이동
2. "사용자 인증 정보 만들기" > "서비스 계정" 선택
3. 서비스 계정 이름을 입력하고 생성
4. 생성된 서비스 계정을 클릭하여 "키" 탭으로 이동
5. "키 추가" > "새 키 만들기" > "JSON" 선택
6. JSON 키 파일이 다운로드됩니다

### 3. Google Sheets 준비

1. [Google Sheets](https://sheets.google.com/)에서 새 스프레드시트를 생성합니다
2. 첫 번째 시트의 첫 번째 행에 다음 헤더를 입력합니다:
   ```
   날짜    카테고리    상점명    총금액    항목수    항목내역    결제방법    영수증링크    메모
   ```
3. 스프레드시트를 서비스 계정 이메일과 공유합니다:
   - 스프레드시트 우측 상단의 "공유" 버튼 클릭
   - 서비스 계정 이메일 주소 입력 (예: `your-service-account@your-project.iam.gserviceaccount.com`)
   - "편집자" 권한 부여
4. 스프레드시트 URL에서 ID를 복사합니다:
   ```
   https://docs.google.com/spreadsheets/d/[SPREADSHEET_ID]/edit
   ```

### 4. Lambda 환경 변수 설정

AWS Lambda 함수에 다음 환경 변수를 추가합니다:

1. `GOOGLE_CREDENTIALS_JSON`: 다운로드한 JSON 키 파일의 내용 전체를 문자열로 입력
   ```json
   {
     "type": "service_account",
     "project_id": "your-project",
     "private_key_id": "...",
     "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
     "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
     ...
   }
   ```

2. `GOOGLE_SPREADSHEET_ID`: Google Sheets 스프레드시트 ID

3. `OPENAI_API_KEY`: OpenAI API 키 (영수증 데이터 추출용)

## 코드 구조

### sheets_repository.go

Google Sheets와 통신하는 Repository 패턴 구현:

- `SheetsRepository` interface: 영수증 데이터 저장 인터페이스
- `GoogleSheetsRepository`: Google Sheets API를 사용한 구현체
- `SaveReceipt()`: 영수증 데이터를 Google Sheets에 저장

### 데이터 변환

`ReceiptData` (영수증 추출 데이터) → `SheetRow` (Google Sheets 행) 변환:

- **날짜**: `TransactionDate`를 `YYYY-MM-DD` 형식으로 변환
- **카테고리**: 현재는 비어있음 (향후 구현 가능)
- **상점명**: `MerchantName`
- **총금액**: `Total` (소수점 2자리)
- **항목수**: `Items` 배열의 길이
- **항목내역**: 항목들을 "상품명 (수량) x 가격" 형식으로 조합
- **결제방법**: `PaymentMethod`와 `CardLastFour`를 조합
- **영수증링크**: S3 URL
- **메모**: 비어있음 (나중에 수동으로 입력 가능)

## 사용 예시

영수증 이미지를 업로드하면 자동으로:

1. S3에 이미지 저장
2. OpenAI로 데이터 추출
3. Google Sheets에 행 추가

Lambda 로그에서 확인 가능:
```
[INFO] Saving receipt to Google Sheets - merchant: Starbucks, total: 15.50
[INFO] Successfully saved receipt to Google Sheets - merchant: Starbucks
```

## 에러 처리

- Google Sheets 저장 실패 시에도 Lambda 함수는 성공을 반환합니다
- 에러는 CloudWatch Logs에 기록됩니다
- 영수증 추출 실패 시에도 S3 업로드는 완료됩니다

## 테스트

로컬 테스트를 위해 환경 변수를 설정하고 실행:

```bash
export GOOGLE_CREDENTIALS_JSON='{"type":"service_account",...}'
export GOOGLE_SPREADSHEET_ID='your-spreadsheet-id'
export OPENAI_API_KEY='your-openai-key'

go run main.go types.go openai_client.go service.go sheets_repository.go
```

## 주의사항

1. Google Sheets API 할당량 제한이 있습니다 (일반적으로 충분함)
2. 서비스 계정 JSON 키는 안전하게 보관해야 합니다
3. 스프레드시트를 서비스 계정과 공유하는 것을 잊지 마세요
4. 환경 변수가 설정되지 않으면 Google Sheets 통합이 비활성화되지만 S3 업로드는 계속 작동합니다
