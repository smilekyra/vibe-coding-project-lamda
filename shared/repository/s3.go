package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// FileInfo contains information about an uploaded file
type FileInfo struct {
	OriginalName string `json:"original_name"`
	FileName     string `json:"file_name"`
	BucketName   string `json:"bucket_name"`
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	URL          string `json:"url"`
	UploadDate   string `json:"upload_date"`
}

// S3Repository handles S3 operations
type S3Repository struct {
	client     *s3.Client
	bucketName string
	region     string
}

// NewS3Repository creates a new S3 repository
func NewS3Repository(client *s3.Client, bucketName, region string) *S3Repository {
	return &S3Repository{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}
}

// EnsureBucketExists creates the S3 bucket if it doesn't exist
func (r *S3Repository) EnsureBucketExists(ctx context.Context) error {
	// Check if bucket exists
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	})

	if err == nil {
		return nil
	}

	// Create bucket
	_, err = r.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(r.bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(r.region),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Wait for bucket to be created
	waiter := s3.NewBucketExistsWaiter(r.client)
	err = waiter.Wait(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucketName),
	}, 30*time.Second)

	if err != nil {
		return fmt.Errorf("failed to wait for bucket creation: %w", err)
	}

	return nil
}

// Upload uploads a file to S3 with JST date folder structure
func (r *S3Repository) Upload(ctx context.Context, originalFileName string, fileContent []byte, contentType string) (*FileInfo, error) {
	// Ensure bucket exists
	if err := r.EnsureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Generate unique filename
	uniqueFileName := generateUniqueFileName(originalFileName)

	// Create key with JST date folder
	dateFolder := getJSTDateFolder()
	key := fmt.Sprintf("%s/%s", dateFolder, uniqueFileName)

	// Upload file to S3
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(fileContent)),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct file URL
	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", r.bucketName, r.region, key)

	fileInfo := &FileInfo{
		OriginalName: originalFileName,
		FileName:     uniqueFileName,
		BucketName:   r.bucketName,
		Key:          key,
		Size:         int64(len(fileContent)),
		ContentType:  contentType,
		URL:          fileURL,
		UploadDate:   dateFolder,
	}

	return fileInfo, nil
}

// getJSTDateFolder returns the current date in JST as YYYY-MM-DD format
func getJSTDateFolder() string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst)
	return now.Format("2006-01-02")
}

// generateUniqueFileName creates a unique filename with JST timestamp and random suffix
func generateUniqueFileName(originalName string) string {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	now := time.Now().In(jst)

	// Extract file extension and base name
	ext := filepath.Ext(originalName)
	baseName := strings.TrimSuffix(originalName, ext)

	// Generate random suffix (4 bytes = 8 hex characters)
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomSuffix := hex.EncodeToString(randomBytes)

	// Create timestamp in JST: YYYYMMDD_HHMMSS
	timestamp := now.Format("20060102_150405")

	// Combine: basename_timestamp_random.ext
	uniqueName := fmt.Sprintf("%s_%s_%s%s", baseName, timestamp, randomSuffix, ext)

	return uniqueName
}
