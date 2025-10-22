package openai

import (
	"strings"
	"testing"
)

func TestValidateImageSize(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty image",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "small image",
			data:    make([]byte, 1024), // 1 KB
			wantErr: false,
		},
		{
			name:    "medium image",
			data:    make([]byte, 10*1024*1024), // 10 MB
			wantErr: false,
		},
		{
			name:    "large image at limit",
			data:    make([]byte, MaxImageSizeBytes), // 50 MB
			wantErr: false,
		},
		{
			name:    "image exceeds limit",
			data:    make([]byte, MaxImageSizeBytes+1), // 50 MB + 1 byte
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageSize(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateImageFormat(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte{0x00, 0x01},
			wantErr: true,
		},
		{
			name:    "PNG format",
			data:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantErr: false,
		},
		{
			name:    "JPEG format",
			data:    []byte{0xFF, 0xD8, 0xFF, 0xE0},
			wantErr: false,
		},
		{
			name:    "GIF format",
			data:    []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61},
			wantErr: false,
		},
		{
			name:    "WEBP format",
			data:    []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			wantErr: false,
		},
		{
			name:    "unsupported format (BMP)",
			data:    []byte{0x42, 0x4D, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "random data",
			data:    []byte{0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageFormat(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateImageForOpenAI(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid PNG",
			data:    []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			wantErr: false,
		},
		{
			name:    "valid JPEG",
			data:    []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			wantErr: false,
		},
		{
			name:    "empty image",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "wrong format",
			data:    []byte{0x42, 0x4D, 0x00, 0x00}, // BMP
			wantErr: true,
		},
		{
			name:    "too large PNG",
			data:    append([]byte{0x89, 0x50, 0x4E, 0x47}, make([]byte, MaxImageSizeBytes)...),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageForOpenAI(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageForOpenAI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateImageSizeFromBase64(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "empty string",
			data:    "",
			wantErr: true,
		},
		{
			name:    "small base64",
			data:    "iVBORw0KGgo=", // Small PNG header
			wantErr: false,
		},
		{
			name:    "with data URI prefix",
			data:    "data:image/png;base64,iVBORw0KGgo=",
			wantErr: false,
		},
		{
			name:    "very large base64",
			data:    strings.Repeat("A", MaxImageSizeBytes*2), // Much larger than limit
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageSizeFromBase64(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageSizeFromBase64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetImageSizeInfo(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "1 MB image",
			data: make([]byte, 1024*1024),
			want: "1.00 MB",
		},
		{
			name: "10 MB image",
			data: make([]byte, 10*1024*1024),
			want: "10.00 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetImageSizeInfo(tt.data)
			if !strings.Contains(got, tt.want) {
				t.Errorf("GetImageSizeInfo() = %v, want to contain %v", got, tt.want)
			}
		})
	}
}

func TestGetImageFormatInfo(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "PNG",
			data: []byte{0x89, 0x50, 0x4E, 0x47},
			want: "PNG",
		},
		{
			name: "JPEG",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0},
			want: "JPEG",
		},
		{
			name: "GIF",
			data: []byte{0x47, 0x49, 0x46, 0x38},
			want: "GIF",
		},
		{
			name: "WEBP",
			data: []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50},
			want: "WEBP",
		},
		{
			name: "Unknown",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			want: "Unknown/Unsupported",
		},
		{
			name: "Too short",
			data: []byte{0x00},
			want: "Unknown format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetImageFormatInfo(tt.data)
			if got != tt.want {
				t.Errorf("GetImageFormatInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageValidationError(t *testing.T) {
	err := &ImageValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "test_field: test message"
	if err.Error() != expected {
		t.Errorf("ImageValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}

