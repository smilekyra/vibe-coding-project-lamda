# 환경 변수 설정 가이드

## 필수 환경 변수

Lambda 함수에 다음 환경 변수를 설정해야 합니다:

### 1. OPENAI_API_KEY
OpenAI API 키 (영수증 데이터 추출용)

### 2. GOOGLE_CREDENTIALS_JSON
Google 서비스 계정 JSON 키 (전체 내용)

### 3. GOOGLE_SPREADSHEET_ID
Google Sheets 스프레드시트 ID
- URL: `https://docs.google.com/spreadsheets/d/[이 부분]/edit`

### 4. GOOGLE_SHEET_NAME (선택사항)
Google Sheets의 시트 이름 (탭 이름)
- **귀하의 경우: `가계부`**
- 설정하지 않으면 첫 번째 시트를 자동으로 사용

## AWS Lambda Console에서 환경 변수 설정

1. [AWS Lambda Console](https://console.aws.amazon.com/lambda/) 로그인
2. 함수 목록에서 **receipt-go** 함수 선택
3. **Configuration** 탭 클릭
4. 왼쪽 메뉴에서 **Environment variables** 선택
5. **Edit** 버튼 클릭
6. **Add environment variable** 클릭하여 다음 4개 추가:

```
Key: OPENAI_API_KEY
Value: sk-proj-... (OpenAI API 키)

Key: GOOGLE_CREDENTIALS_JSON
Value: {"type":"service_account",...} (전체 JSON 내용)

Key: GOOGLE_SPREADSHEET_ID
Value: 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms (실제 ID)

Key: GOOGLE_SHEET_NAME
Value: 가계부
```

7. **Save** 클릭

## 변경 사항

이제 코드가 다음과 같이 동작합니다:

1. `GOOGLE_SHEET_NAME` 환경 변수가 설정되어 있으면 → 해당 시트 사용
2. `GOOGLE_SHEET_NAME` 환경 변수가 없으면 → 첫 번째 시트 자동 사용
3. 시트 이름이 잘못되었거나 존재하지 않으면 → 에러 발생 (CloudWatch Logs에서 확인 가능)

## 설정 후 확인

환경 변수 설정 후 테스트:

```bash
cd /Users/kyra/Documents/vibe-coding-project/251023/vibe-coding-project-lambda/functions/receipt-go
bash test-real-receipt.sh
```

CloudWatch Logs에서 다음 메시지 확인:
```
[INFO] Creating Google Sheets repository - spreadsheetID: xxx, sheetName: 가계부
[INFO] Using first sheet: 가계부
[INFO] Appending to sheet range: 가계부!A:I
[INFO] Successfully saved receipt to Google Sheets
```

## 에러 해결

이전 에러:
```
Error 400: Unable to parse range: Sheet1!A:I
```

이유: 시트 이름이 "Sheet1"이 아니라 "가계부"였기 때문

해결: `GOOGLE_SHEET_NAME=가계부` 환경 변수 추가
