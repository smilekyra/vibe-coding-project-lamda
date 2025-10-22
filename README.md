# AWS Lambda Multi-Function Project (Go)

A beginner-friendly AWS Lambda project in Go that makes it easy to manage multiple functions. Each function has its own deployment workflow - just copy and paste to add new functions!

## 🎯 Why This Structure?

- ✅ **Simple**: One workflow file per function - easy to understand
- ✅ **Independent**: Deploy functions separately or together
- ✅ **Smart**: Only deploys when that function changes
- ✅ **Scalable**: Add new functions by copying existing ones
- ✅ **Shared Code**: Common utilities in `shared/` package

## 📦 Project Structure

```
vibe-coding-project-lambda/
├── functions/                    # All your Lambda functions
│   ├── time-api/                 # Function 1: JST time API
│   │   ├── main.go
│   │   └── main_test.go
│   ├── hello-world/              # Function 2: Hello world API
│   │   ├── main.go
│   │   └── main_test.go
│   └── receipt-processor/        # Function 3: Receipt OCR processor
│       ├── main.go
│       └── main_test.go
├── shared/                       # Shared utilities
│   └── response/                 # API response helpers
│       └── response.go
├── .github/workflows/            # Deployment workflows
│   ├── deploy-time-api.yml       # Deploys time-api
│   ├── deploy-hello-world.yml    # Deploys hello-world
│   ├── deploy-receipt-processor.yml  # Deploys receipt-processor
│   └── _TEMPLATE.yml             # Copy this for new functions
└── scripts/
    ├── new-function.sh           # Create new function
    └── create-functions.sh       # Create functions in AWS
```

## 🚀 Available Functions

> **Runtime:** All functions use Go with `provided.al2023` custom runtime and `bootstrap` handler

### time-api
Returns current time in JST timezone.

**Test it:**
```bash
curl https://your-api.com/time
```

**Response:**
```json
{
  "message": "Current time in Japan",
  "current_time": "2025-10-18 15:30:45",
  "timezone": "JST (Asia/Tokyo)",
  "timestamp": 1729238445
}
```

### hello-world
Simple hello API with optional name parameter.

**Test it:**
```bash
curl "https://your-api.com/hello?name=Alice"
```

**Response:**
```json
{
  "message": "Hello, Alice!",
  "version": "1.0.0",
  "timestamp": 1729238445
}
```

### receipt-processor
Process receipt images with OCR and extract structured data. Stores images in S3 with JST date-based folder organization.

**Test it:**
```bash
# Prepare base64 encoded receipt image
FILE_CONTENT=$(base64 -i receipt.jpg)

# Upload receipt image
curl -X POST "https://your-api.com/receipt" \
  -H "Content-Type: application/json" \
  -d "{
    \"filename\": \"receipt.jpg\",
    \"file_content\": \"$FILE_CONTENT\",
    \"content_type\": \"image/jpeg\"
  }"
```

**Response:**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "file_info": {
    "file_name": "receipt.jpg",
    "bucket_name": "lambda-file-uploads",
    "key": "2025-10-18/receipt.jpg",
    "size": 45123,
    "content_type": "image/jpeg",
    "url": "https://lambda-file-uploads.s3.ap-northeast-1.amazonaws.com/2025-10-18/receipt.jpg",
    "upload_date": "2025-10-18"
  },
  "timestamp": 1729238445
}
```

**Features:**
- ✅ Upload receipt images to S3
- ✅ Automatic S3 bucket creation (if not exists)
- ✅ JST date-based folder structure (YYYY-MM-DD)
- ✅ Base64 encoded file content
- 🚧 OCR text extraction (coming soon)
- 🚧 Structured receipt data extraction (coming soon)

**Environment Variables:**
- `S3_BUCKET_NAME` (optional): Custom bucket name (default: `lambda-file-uploads`)

## 🛠️ Quick Start

### 1. Install Dependencies
```bash
make deps
```

### 2. Build All Functions
```bash
make build-all
```

### 3. Test Locally
```bash
make test
```

### 4. Deploy to AWS
Push to master branch and the functions will auto-deploy!

## 📋 How to Add a New Function

### Step 1: Create Function Code

Create manually:
```bash
mkdir -p functions/user-api
# Add main.go and main_test.go
```

### Step 2: Create Deployment Workflow

Copy the template:
```bash
cp .github/workflows/_TEMPLATE.yml .github/workflows/deploy-user-api.yml
```

Open `deploy-user-api.yml` and replace `YOUR-FUNCTION-NAME` with `user-api` (4 places).

**That's it!** 🎉

### Step 3: Deploy

The `aws-lambda-deploy` action will automatically create the Lambda function if it doesn't exist!

```bash
git add .
git commit -m "Add user-api function"
git push origin master
```

GitHub Actions will automatically:
1. Build your function
2. Create the Lambda function (if it doesn't exist)
3. Deploy your code

🎉 That's it!

## 🔧 Available Commands

```bash
# Build all functions
make build-all

# Build specific function
make build FUNCTION=time-api

# Test all functions
make test

# Test specific function
make test-function FUNCTION=hello-world

# Create deployment zips
make zip-all                    # All functions
make zip FUNCTION=time-api      # Specific function

# List available functions
make list-functions

# Clean build artifacts
make clean
```

## 🤖 GitHub Actions Setup

### Required Secrets

Go to GitHub → Settings → Secrets and variables → Actions

Add these 4 secrets:

| Secret | Example | Description |
|--------|---------|-------------|
| `AWS_ROLE_ARN` | `arn:aws:iam::123456789012:role/GitHubActionsLambdaRole` | IAM role for GitHub Actions deployment |
| `AWS_REGION` | `us-east-1` | AWS region |
| `LAMBDA_FUNCTION_PREFIX` | `vibe-` | Prefix for function names |
| `LAMBDA_EXECUTION_ROLE_ARN` | `arn:aws:iam::123456789012:role/lambda-execution-role` | IAM role for Lambda function execution |

**Note:** You need TWO different roles:
- GitHub Actions role (for deployment)
- Lambda execution role (for running functions)

### How Deployment Works

Each function has its own workflow file:
- **Triggers automatically** when you push changes to that function
- **Manual trigger** via GitHub Actions tab
- **Independent deployment** - one function failing doesn't affect others

Example: Change `functions/time-api/main.go` → Only `deploy-time-api.yml` runs!

## 🎓 Understanding the Workflow Files

Each workflow file is simple and self-contained:

```yaml
name: Deploy time-api              # 1. Name it clearly

on:
  push:
    paths:
      - 'functions/time-api/**'    # 2. Only run when this function changes
      - 'shared/**'                # Also run if shared code changes

jobs:
  deploy:
    steps:
      - Checkout code
      - Setup Go
      - Build time-api             # 3. Build this specific function
      - Configure AWS
      - Deploy to Lambda           # 4. Deploy to vibe-time-api
```

**Benefits:**
- ✅ Easy to read and understand
- ✅ Easy to customize per function
- ✅ No complex matrix or loops
- ✅ Perfect for beginners

## 📚 Using Shared Code

Put common utilities in `shared/`:

```go
// In your function
import "vibe-coding-project-lambda/shared/response"

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) {
    // Use shared helpers
    return response.Success(200, myData)
    return response.Error(500, "Something went wrong")
    return response.MethodNotAllowed("GET")
}
```

When you update `shared/`, all functions that use it will be redeployed automatically!

## 🔍 Troubleshooting

### Function not deploying?

Check GitHub Actions tab to see the workflow status.

### Wrong function name in AWS?

Make sure you set `LAMBDA_FUNCTION_PREFIX` secret correctly. Function names are: `PREFIX + function-name`

Example: `vibe-` + `time-api` = `vibe-time-api`

### Build fails?

```bash
make clean
make deps
make build-all
```

### Want to deploy manually?

Go to GitHub → Actions → Select workflow → Run workflow

## 📊 Monitoring & Logs

Your Lambda functions automatically write logs to CloudWatch. The `lambda-execution-role` includes permissions for:
- ✅ Creating log groups
- ✅ Creating log streams
- ✅ Writing log events

### View Logs via AWS CLI

```bash
# View real-time logs for time-api
aws logs tail /aws/lambda/vibe-time-api --follow

# View logs from last 10 minutes
aws logs tail /aws/lambda/vibe-hello-world --since 10m

# View logs from specific time
aws logs tail /aws/lambda/vibe-time-api --since 2025-10-18T14:00:00
```

### View Logs in AWS Console

1. Go to CloudWatch: https://console.aws.amazon.com/cloudwatch/
2. Click **Logs** → **Log groups**
3. Find `/aws/lambda/vibe-time-api` or `/aws/lambda/vibe-hello-world`
4. Click to view logs and filter by time

### CloudWatch Logs Not Appearing?

Check if the execution role has the correct permissions:

```bash
aws iam get-attached-role-policies --role-name lambda-execution-role
# Should show: AWSLambdaBasicExecutionRole
```

## 📖 Complete Setup Guide

See `AWS_SETUP.md` for detailed AWS infrastructure setup including:
- Creating IAM roles
- Setting up OIDC authentication
- Creating Lambda functions
- Configuring API Gateway

## 🎯 Best Practices

1. **One workflow per function** - Keep it simple
2. **Use the template** - Copy `_TEMPLATE.yml` for new functions
3. **Test locally first** - Run `make test` before pushing
4. **Shared IAM role** - All functions use the same role
5. **Name consistently** - Use kebab-case (like `user-api`)

## 💡 Tips

- **Deploy all functions?** Push changes to `shared/` - all workflows run!
- **Deploy one function?** Only change that function's directory
- **Customize per function?** Edit that function's workflow file
- **No manual AWS setup needed!** The deploy action creates functions automatically

## 🚀 Real-World Example

Let's add a "user-api" function from scratch:

```bash
# 1. Create function directory and code
mkdir -p functions/user-api
# Add main.go and main_test.go (copy from existing functions)

# 2. Copy workflow template
cp .github/workflows/_TEMPLATE.yml .github/workflows/deploy-user-api.yml

# 3. Replace YOUR-FUNCTION-NAME with user-api
# In your editor: Find "YOUR-FUNCTION-NAME" → Replace with "user-api"

# 4. Test locally
make build FUNCTION=user-api
make test-function FUNCTION=user-api

# 5. Commit and push
git add .
git commit -m "Add user-api function"
git push origin master
```

Done! GitHub Actions will automatically create the Lambda function and deploy it. 🎉

## 📄 Files Explained

| File | Purpose |
|------|---------|
| `functions/*/main.go` | Your Lambda function code |
| `functions/*/main_test.go` | Tests for your function |
| `shared/response/` | Helper functions used by all functions |
| `.github/workflows/deploy-*.yml` | Deployment workflow for each function |
| `.github/workflows/_TEMPLATE.yml` | Copy this to add new functions |
| `Makefile` | Build commands |
| `go.mod` | Go dependencies |

## 🤝 Contributing

1. Fork the repository
2. Add your function following the guide above
3. Test locally
4. Submit a Pull Request

## 📝 License

MIT License

## 🔗 Learn More

- [AWS Lambda Go Runtime](https://github.com/aws/aws-lambda-go)
- [GitHub Actions](https://docs.github.com/en/actions)
- [AWS Lambda Deploy Action](https://github.com/aws-actions/aws-lambda-deploy)

## ❓ FAQ

**Q: Why one workflow per function instead of a matrix?**  
A: Simpler for beginners! Each workflow is self-contained and easy to understand.

**Q: Can I still deploy multiple functions at once?**  
A: Yes! If you change `shared/` or push changes to multiple functions, all their workflows run.

**Q: Do I need separate IAM roles?**  
A: No! All functions share one IAM role with wildcard permissions (`vibe-*`).

**Q: How do I customize a function's deployment?**  
A: Just edit that function's workflow file. They're independent!

**Q: What if I want different Go versions per function?**  
A: Edit the `go-version` in that function's workflow file.
