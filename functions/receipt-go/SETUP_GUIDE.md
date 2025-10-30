# Lambda 환경 변수 설정 가이드

구글 시트 통합 기능을 사용하려면 Lambda 함수에 환경 변수를 설정해야 합니다.

## 문제 진단

현재 구글 시트에 저장이 안 되는 이유는 **환경 변수가 설정되지 않았기 때문**입니다.

Lambda 함수는 다음 순서로 동작합니다:
1. 환경 변수 확인 (`initServices` 함수)
2. 환경 변수가 없으면 → 통합 기능 비활성화
3. 환경 변수가 있으면 → Google Sheets 연동 시도

## CloudWatch Logs 확인 방법

### 방법 1: AWS Console
1. [AWS Console](https://console.aws.amazon.com/) 로그인
2. **CloudWatch** 서비스로 이동
3. 왼쪽 메뉴에서 **Logs > Log groups** 클릭
4. `/aws/lambda/receipt-go` 또는 유사한 이름의 로그 그룹 찾기
5. 최신 로그 스트림 클릭
6. 다음과 같은 로그 메시지 찾기:
   ```
   [WARN] Google Sheets credentials not found, sheets integration disabled
   [WARN] OpenAI API key not found, extraction service disabled
   ```

### 방법 2: AWS CLI (설치 필요)
```bash
# 최근 10분간의 로그 확인
aws logs tail /aws/lambda/receipt-go --follow --since 10m
```

## 필요한 환경 변수

### 1. OPENAI_API_KEY
OpenAI API 키 (영수증 데이터 추출용)

**얻는 방법:**
1. https://platform.openai.com/ 방문
2. 로그인 후 **API keys** 메뉴
3. **Create new secret key** 클릭
4. 생성된 키 복사 (예: `sk-proj-...`)

### 2. GOOGLE_CREDENTIALS_JSON
Google 서비스 계정 JSON 키 (전체 내용)

**얻는 방법:**
1. https://console.cloud.google.com/ 방문
2. 프로젝트 선택 또는 생성
3. **APIs & Services > Library** → "Google Sheets API" 검색 → 활성화
4. **APIs & Services > Credentials** 이동
5. **Create Credentials > Service Account** 선택
6. 서비스 계정 이름 입력 후 생성
7. 생성된 서비스 계정 클릭
8. **Keys** 탭 → **Add Key > Create new key** → **JSON** 선택
9. 다운로드된 JSON 파일 내용 전체를 복사

JSON 파일 예시:
```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "your-service-account@your-project.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "..."
}
```

### 3. GOOGLE_SPREADSHEET_ID
Google Sheets 스프레드시트 ID

**얻는 방법:**
1. https://sheets.google.com/ 에서 새 스프레드시트 생성
2. 첫 번째 행에 헤더 입력:
   ```
   날짜    카테고리    상점명    총금액    항목수    항목내역    결제방법    영수증링크    메모
   ```
3. URL에서 ID 복사:
   ```
   https://docs.google.com/spreadsheets/d/[이 부분이 SPREADSHEET_ID]/edit
   ```
   예: `1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms`

4. **중요**: 스프레드시트를 서비스 계정 이메일과 공유
   - 우측 상단 **공유** 버튼 클릭
   - 서비스 계정 이메일 입력 (예: `your-service-account@your-project.iam.gserviceaccount.com`)
   - 권한: **편집자**
   - 공유

## Lambda 환경 변수 설정 방법

### AWS Console에서 설정

1. [AWS Lambda Console](https://console.aws.amazon.com/lambda/) 로그인
2. 함수 목록에서 **receipt-go** 함수 선택
3. **Configuration** 탭 클릭
4. 왼쪽 메뉴에서 **Environment variables** 선택
5. **Edit** 버튼 클릭
6. **Add environment variable** 클릭하여 다음 3개 추가:

| Key | Value |
|-----|-------|
| `OPENAI_API_KEY` | `sk-proj-...` (OpenAI API 키) |
| `GOOGLE_CREDENTIALS_JSON` | `{"type":"service_account",...}` (전체 JSON) |
| `GOOGLE_SPREADSHEET_ID` | `1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms` (스프레드시트 ID) |

7. **Save** 클릭

### AWS CLI로 설정 (옵션)

```bash
# OpenAI API Key 설정
aws lambda update-function-configuration \
  --function-name receipt-go \
  --environment Variables="{OPENAI_API_KEY=sk-proj-...}"

# Google Sheets 설정 (JSON을 한 줄로 변환 필요)
aws lambda update-function-configuration \
  --function-name receipt-go \
  --environment Variables="{
    OPENAI_API_KEY=sk-proj-...,
    GOOGLE_CREDENTIALS_JSON=$(cat credentials.json | jq -c .),
    GOOGLE_SPREADSHEET_ID=1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms
  }"
```

## GitHub Actions Secrets 설정 (자동 배포용)

자동 배포 시 환경 변수를 설정하려면:

1. GitHub 저장소 → **Settings** → **Secrets and variables** → **Actions**
2. **New repository secret** 클릭
3. 다음 secret들 추가:
   - `OPENAI_API_KEY`
   - `GOOGLE_CREDENTIALS_JSON`
   - `GOOGLE_SPREADSHEET_ID`

4. `.github/workflows/deploy-receipt-go.yml` 파일에 환경 변수 설정 추가:

```yaml
- name: Deploy receipt-go to Lambda
  uses: aws-actions/aws-lambda-deploy@v1
  with:
    function-name: ${{ secrets.LAMBDA_FUNCTION_PREFIX }}receipt-go
    code-artifacts-dir: ./dist/receipt-go
    runtime: provided.al2023
    handler: bootstrap
    role: ${{ secrets.LAMBDA_EXECUTION_ROLE_ARN }}
    timeout: 30
    memory-size: 1024
    environment-variables: |
      OPENAI_API_KEY=${{ secrets.OPENAI_API_KEY }}
      GOOGLE_CREDENTIALS_JSON=${{ secrets.GOOGLE_CREDENTIALS_JSON }}
      GOOGLE_SPREADSHEET_ID=${{ secrets.GOOGLE_SPREADSHEET_ID }}
```

## 테스트

환경 변수 설정 후 다시 테스트:

```bash
cd /Users/kyra/Documents/vibe-coding-project/251023/vibe-coding-project-lambda/functions/receipt-go
bash test-real-receipt.sh
```

CloudWatch Logs에서 다음 메시지를 확인:
```
[INFO] Initializing Google Sheets repository
[INFO] Google Sheets repository initialized successfully
[INFO] Initializing OpenAI extraction service
[INFO] OpenAI extraction service initialized successfully
[INFO] Extracting receipt data from image
[INFO] Receipt extraction successful
[INFO] Saving receipt data to Google Sheets
[INFO] Successfully saved receipt to Google Sheets
```

## 문제 해결

### "Google Sheets credentials not found"
→ `GOOGLE_CREDENTIALS_JSON` 환경 변수가 설정되지 않았거나 비어있음

### "OpenAI API key not found"
→ `OPENAI_API_KEY` 환경 변수가 설정되지 않았거나 비어있음

### "Failed to initialize Google Sheets repository"
→ JSON 형식이 잘못되었거나 서비스 계정 권한 문제

### "Failed to save to Google Sheets"
→ 스프레드시트가 서비스 계정과 공유되지 않았거나 스프레드시트 ID가 잘못됨

### "Failed to extract receipt data"
→ OpenAI API 키가 잘못되었거나 할당량 초과

## 확인 체크리스트

- [ ] OpenAI API 키 발급
- [ ] Google Cloud 프로젝트 생성
- [ ] Google Sheets API 활성화
- [ ] 서비스 계정 생성 및 JSON 키 다운로드
- [ ] Google Sheets 스프레드시트 생성
- [ ] 스프레드시트에 헤더 행 추가
- [ ] 스프레드시트를 서비스 계정과 공유 (편집자 권한)
- [ ] Lambda 함수에 3개 환경 변수 설정
- [ ] 테스트 실행
- [ ] CloudWatch Logs 확인
