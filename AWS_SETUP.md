# AWS Setup Guide

This guide walks you through setting up the AWS infrastructure needed for multiple Lambda functions.

## 🏗️ Multi-Function Architecture

This project supports multiple Lambda functions. You'll need to create each function in AWS with a consistent naming convention.

### Naming Convention

Functions are deployed with the format: `${LAMBDA_FUNCTION_PREFIX}${function-name}`

**Example:**
- Prefix: `vibe-`
- Functions: `time-api`, `hello-world`
- Deployed as: `vibe-time-api`, `vibe-hello-world`

## Step 1: Create Lambda Execution Role

First, create a role that your Lambda functions will use when they execute. This is different from the GitHub Actions role.

  **This role allows Lambda to:**
- ✅ Create CloudWatch Log Groups
- ✅ Create CloudWatch Log Streams  
- ✅ Write logs to CloudWatch (PutLogEvents)

### Using AWS Console

1. **Go to IAM Console** → https://console.aws.amazon.com/iam/
2. Click **Roles** → **Create role**
3. **Trusted entity type**: `AWS service`
4. **Use case**: Select `Lambda`
5. Click **Next**
6. **Add permissions**: Search and select `AWSLambdaBasicExecutionRole`
   - This policy includes CloudWatch Logs permissions
7. Click **Next**
8. **Role name**: `lambda-execution-role`
9. **Description**: `Execution role for Lambda functions with CloudWatch logging`
10. Click **Create role**
11. **Copy the Role ARN**: Click on the role and copy the ARN
    - Example: `arn:aws:iam::728643659807:role/lambda-execution-role`

### Using AWS CLI

```bash
# Create trust policy for Lambda
cat > lambda-trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# Create the role
aws iam create-role \
  --role-name lambda-execution-role \
  --assume-role-policy-document file://lambda-trust-policy.json

# Attach basic execution policy (includes CloudWatch Logs)
aws iam attach-role-policy \
  --role-name lambda-execution-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

# Get the role ARN
aws iam get-role --role-name lambda-execution-role --query 'Role.Arn' --output text
```

**What `AWSLambdaBasicExecutionRole` includes:**
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ],
    "Resource": "arn:aws:logs:*:*:*"
  }]
}
```

This allows your Lambda functions to:
- Create log groups in CloudWatch
- Create log streams
- Write log events

**Save this ARN!** You'll need it for GitHub secrets.

## Step 2: Lambda Functions (Auto-Created!)

**Good news!** The `aws-lambda-deploy` GitHub Action automatically creates Lambda functions if they don't exist. You just push your code and it handles the rest!

**Go Runtime Configuration:**
- **Runtime**: `provided.al2023` (Custom runtime for Go)
- **Handler**: `bootstrap` (the compiled Go binary)
- **Architecture**: `x86_64`

The workflows are already configured with these settings.

## Step 3: Create API Gateway (Optional)

To expose your Lambda function as an HTTP API:

### Using AWS Console

1. Go to API Gateway Console
2. Click **Create API**
3. Choose **HTTP API** → **Build**
4. Configure:
   - **API name**: `vibe-coding-api`
   - **Add integration**: Lambda
   - **Lambda function**: Select your function
   - **Method**: GET
5. Click **Create**
6. Note your API endpoint URL

### Using AWS CLI

```bash
# Create API
aws apigatewayv2 create-api \
  --name vibe-coding-api \
  --protocol-type HTTP \
  --target arn:aws:lambda:REGION:ACCOUNT_ID:function:vibe-coding-project-lambda

# Create integration
aws apigatewayv2 create-integration \
  --api-id YOUR_API_ID \
  --integration-type AWS_PROXY \
  --integration-uri arn:aws:lambda:REGION:ACCOUNT_ID:function:vibe-coding-project-lambda \
  --payload-format-version 2.0

# Create route
aws apigatewayv2 create-route \
  --api-id YOUR_API_ID \
  --route-key 'GET /' \
  --target integrations/YOUR_INTEGRATION_ID
```

## Step 4: Set Up GitHub OIDC Provider (Secure Authentication)

### Why OIDC?

Instead of storing AWS credentials as GitHub secrets (which can be leaked), OIDC allows GitHub Actions to get temporary credentials directly from AWS. **Much more secure!**

### Step 4.1: Create OIDC Identity Provider in AWS

#### Option A: Using AWS Console (Recommended for Beginners)

1. **Go to IAM Console**
   - Open: https://console.aws.amazon.com/iam/
   - Click **Identity providers** in the left sidebar
   - Click **Add provider**

2. **Configure the provider**
   - **Provider type**: Select `OpenID Connect`
   - **Provider URL**: `https://token.actions.githubusercontent.com`
   - Click **Get thumbprint** (it will auto-fill)
   - **Audience**: `sts.amazonaws.com`
   - Click **Add provider**

3. **Verify**
   - You should see the provider listed as: `token.actions.githubusercontent.com`

#### Option B: Using AWS CLI

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

**Expected output:**
```json
{
    "OpenIDConnectProviderArn": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
}
```

arn:aws:iam::728643659807:oidc-provider/token.actions.githubusercontent.com

**Note:** Save this ARN - you'll need it in the next step!

### Step 4.2: Create IAM Role for GitHub Actions

This role allows GitHub Actions from YOUR repository to deploy Lambda functions.

#### Option A: Using AWS Console (Recommended for Beginners)

1. **Go to IAM Roles**
   - Open: https://console.aws.amazon.com/iam/
   - Click **Roles** in the left sidebar
   - Click **Create role**

2. **Select Trusted Entity**
   - Select **Web identity**
   - **Identity provider**: Select `token.actions.githubusercontent.com`
   - **Audience**: Select `sts.amazonaws.com`
   - Click **Next**

3. **Add Permissions**
   - Don't select any policies yet, we'll add custom policy next
   - Click **Next**

4. **Name and Create**
   - **Role name**: `GitHubActionsLambdaRole`
   - **Description**: `Role for GitHub Actions to deploy Lambda functions`
   - Click **Create role**

5. **Edit Trust Policy**
   - Find your new role in the list and click on it
   - Click on **Trust relationships** tab
   - Click **Edit trust policy**
   - Replace with this (update YOUR_GITHUB_USERNAME and YOUR_REPO_NAME):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_USERNAME/YOUR_REPO_NAME:*"
        }
      }
    }
  ]
}
```

   - **Important**: Replace:
     - `YOUR_ACCOUNT_ID` with your AWS account ID (e.g., `123456789012`)
     - `YOUR_GITHUB_USERNAME` with your GitHub username (e.g., `johndoe`)
     - `YOUR_REPO_NAME` with your repository name (e.g., `lambda-project`)
   - Click **Update policy**

#### Option B: Using AWS CLI

1. **Create trust policy file:**

Create a file `trust-policy.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:YOUR_GITHUB_USERNAME/YOUR_REPO_NAME:*"
        }
      }
    }
  ]
}
```

2. **Create the role:**

```bash
aws iam create-role \
  --role-name GitHubActionsLambdaRole \
  --assume-role-policy-document file://trust-policy.json
```

**Expected output:**
```json
{
    "Role": {
        "RoleName": "GitHubActionsLambdaRole",
        "Arn": "arn:aws:iam::123456789012:role/GitHubActionsLambdaRole"
    }
}
```

**Note:** Save this ARN - you'll need it for GitHub secrets!

### Step 4.3: Attach Lambda Permissions to the Role

Now give the role permission to deploy Lambda functions.

#### Option A: Using AWS Console

1. **Go to your role**
   - Open: https://console.aws.amazon.com/iam/
   - Click **Roles** → Find `GitHubActionsLambdaRole`
   - Click on the role name

2. **Add inline policy**
   - Click **Permissions** tab
   - Click **Add permissions** → **Create inline policy**
   - Click **JSON** tab
   - Paste this policy (replace YOUR_ACCOUNT_ID):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:GetFunctionConfiguration",
        "lambda:CreateFunction",
        "lambda:UpdateFunctionCode",
        "lambda:UpdateFunctionConfiguration",
        "lambda:PublishVersion"
      ],
      "Resource": "arn:aws:lambda:*:YOUR_ACCOUNT_ID:function:vibe-*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole"
      ],
      "Resource": "arn:aws:iam::YOUR_ACCOUNT_ID:role/lambda-execution-role"
    }
  ]
}
```

3. **Name and create**
   - Click **Next**
   - **Policy name**: `LambdaDeployPolicy`
   - Click **Create policy**

#### Option B: Using AWS CLI

Create a file `lambda-policy.json`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:GetFunctionConfiguration",
        "lambda:CreateFunction",
        "lambda:UpdateFunctionCode",
        "lambda:UpdateFunctionConfiguration",
        "lambda:PublishVersion"
      ],
      "Resource": "arn:aws:lambda:*:YOUR_ACCOUNT_ID:function:vibe-*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole"
      ],
      "Resource": "arn:aws:iam::YOUR_ACCOUNT_ID:role/lambda-execution-role"
    }
  ]
}
```

Attach the policy:

```bash
aws iam put-role-policy \
  --role-name GitHubActionsLambdaRole \
  --policy-name LambdaDeployPolicy \
  --policy-document file://lambda-policy.json
```

**Expected output:** (No output means success)

#### Verify Your Setup

```bash
# Check if the role exists
aws iam get-role --role-name GitHubActionsLambdaRole

# Check if the policy is attached
aws iam get-role-policy \
  --role-name GitHubActionsLambdaRole \
  --policy-name LambdaDeployPolicy
```

> **Note**: The workflow uses the official [aws-actions/aws-lambda-deploy](https://github.com/aws-actions/aws-lambda-deploy) action, which automatically handles zip file creation, upload, and function updates with proper error handling and validation.

---

## ✅ Quick Checklist: OIDC Setup

After completing Step 4, verify you have:

- [ ] **OIDC Provider** created in AWS IAM
  - Name: `token.actions.githubusercontent.com`
- [ ] **IAM Role** created: `GitHubActionsLambdaRole`
  - Trust policy points to your GitHub repo
- [ ] **Lambda permissions** attached to the role
  - Can create/update functions with prefix `vibe-*`
- [ ] **Role ARN** saved (you'll need it for GitHub secrets)
  - Format: `arn:aws:iam::123456789012:role/GitHubActionsLambdaRole`

---

## Step 5: Configure GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions

Add the following secrets:

| Secret Name | Value | Example |
|-------------|-------|---------|
| `AWS_ROLE_ARN` | GitHub Actions IAM role ARN (from Step 4.2) | `arn:aws:iam::123456789012:role/GitHubActionsLambdaRole` |
| `AWS_REGION` | AWS region where Lambda is deployed | `us-east-1` or `ap-northeast-1` |
| `LAMBDA_FUNCTION_PREFIX` | Prefix for all Lambda function names | `vibe-` |
| `LAMBDA_EXECUTION_ROLE_ARN` | Lambda execution role ARN (from Step 1) | `arn:aws:iam::123456789012:role/lambda-execution-role` |

**Note:** You need TWO different role ARNs:
- `AWS_ROLE_ARN` - For GitHub Actions to deploy (Step 4)
- `LAMBDA_EXECUTION_ROLE_ARN` - For Lambda functions to execute (Step 1)

## Step 6: Grant API Gateway Permission to Lambda

If you're using API Gateway, grant permissions for each function:

```bash
PREFIX="vibe-"
REGION="us-east-1"
ACCOUNT_ID="123456789012"
API_ID="your-api-id"

FUNCTIONS=("time-api" "hello-world")

for func in "${FUNCTIONS[@]}"; do
  echo "Adding permission for ${PREFIX}${func}..."
  aws lambda add-permission \
    --function-name ${PREFIX}${func} \
    --statement-id apigateway-invoke \
    --action lambda:InvokeFunction \
    --principal apigateway.amazonaws.com \
    --source-arn "arn:aws:execute-api:${REGION}:${ACCOUNT_ID}:${API_ID}/*/*"
done
```

## Step 7: Test Your Setup

### Test Lambda Functions Directly

```bash
# Test time-api
aws lambda invoke \
  --function-name vibe-time-api \
  --payload '{"httpMethod":"GET"}' \
  response-time.json
cat response-time.json

# Test hello-world
aws lambda invoke \
  --function-name vibe-hello-world \
  --payload '{"httpMethod":"GET","queryStringParameters":{"name":"Alice"}}' \
  response-hello.json
cat response-hello.json
```

### Test via API Gateway

```bash
# Test time-api
curl https://YOUR_API_ID.execute-api.REGION.amazonaws.com/time

# Expected response:
# {
#   "message": "Current time in Japan",
#   "current_time": "2025-10-18 15:30:45",
#   "timezone": "JST (Asia/Tokyo)",
#   "timestamp": 1729238445
# }

# Test hello-world
curl "https://YOUR_API_ID.execute-api.REGION.amazonaws.com/hello?name=Alice"

# Expected response:
# {
#   "message": "Hello, Alice!",
#   "version": "1.0.0",
#   "timestamp": 1729238445
# }
```

## Troubleshooting

### Lambda Function Not Found

List all your Lambda functions:

```bash
aws lambda list-functions --query 'Functions[?starts_with(FunctionName, `vibe-`)].FunctionName'
```

Verify specific function exists:

```bash
aws lambda get-function --function-name vibe-time-api
```

### GitHub Actions Permission Denied

Verify the IAM role trust policy allows your GitHub repository:

```bash
aws iam get-role --role-name GitHubActionsLambdaRole
```

Check the IAM policy has wildcard permission for all your functions:

```bash
aws iam get-role-policy --role-name GitHubActionsLambdaRole --policy-name LambdaDeployPolicy
```

### Deployment Succeeds but Function Not Updated

Check if the correct function name is being used:

```bash
# GitHub Secret should be the prefix only
echo $LAMBDA_FUNCTION_PREFIX  # Should be: vibe-
# NOT the full function name like: vibe-time-api
```

### API Gateway 502 Error

Check logs for specific function:

```bash
# time-api logs
aws logs tail /aws/lambda/vibe-time-api --follow

# hello-world logs
aws logs tail /aws/lambda/vibe-hello-world --follow
```

### One Function Fails to Deploy

GitHub Actions uses `fail-fast: false`, so other functions continue deploying. Check the specific function logs in GitHub Actions.

### Can't See CloudWatch Logs

If Lambda functions run but you don't see logs:

```bash
# Check if execution role has CloudWatch permissions
aws iam get-attached-role-policies --role-name lambda-execution-role

# Should include: arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

If missing, attach it:
```bash
aws iam attach-role-policy \
  --role-name lambda-execution-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

## Cost Estimation

- **Lambda**: First 1M requests/month are free, then $0.20 per 1M requests
- **API Gateway**: First 1M requests are free for 12 months, then $1.00 per million requests
- **Data Transfer**: First 1GB/month free

For a low-traffic API, this should stay within AWS Free Tier limits.

## Security Best Practices

1. ✅ Use OIDC instead of long-lived AWS credentials
2. ✅ Follow principle of least privilege for IAM roles
3. ✅ Enable AWS CloudTrail for audit logging
4. ✅ Use API Gateway throttling to prevent abuse
5. ✅ Enable API Gateway request validation

## Next Steps

- [ ] Set up CloudWatch alarms for errors
- [ ] Add API Gateway custom domain
- [ ] Implement request logging
- [ ] Add rate limiting
- [ ] Set up staging environment

