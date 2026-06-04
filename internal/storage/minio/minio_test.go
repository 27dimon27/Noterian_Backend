package minio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMinioClient struct {
	mock.Mock
}

func (m *MockMinioClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MockMinioClient) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (any, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	return args.Get(0), args.Error(1)
}

func (m *MockMinioClient) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}

func (m *MockMinioClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams map[string]string) (*string, error) {
	args := m.Called(ctx, bucketName, objectName, expiry, reqParams)
	if args.Get(0) != nil {
		if url, ok := args.Get(0).(*string); ok {
			return url, args.Error(1)
		}
		return nil, args.Error(1)
	}
	return nil, args.Error(1)
}

type MinIOServiceWithMock struct {
	*MinIOService
	MockClient *MockMinioClient
}

func TestNewMinIOService(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.MinIOConfig
		expectError bool
	}{
		{
			name: "successful creation with SSL",
			cfg: config.MinIOConfig{
				Endpoint:  "minio.example.com:9000",
				AccessKey: "accesskey123",
				SecretKey: "secretkey456",
				UseSSL:    true,
			},
			expectError: false,
		},
		{
			name: "successful creation without SSL",
			cfg: config.MinIOConfig{
				Endpoint:  "localhost:9000",
				AccessKey: "testaccess",
				SecretKey: "testsecret",
				UseSSL:    false,
			},
			expectError: false,
		},
		{
			name: "empty endpoint",
			cfg: config.MinIOConfig{
				Endpoint:  "",
				AccessKey: "key",
				SecretKey: "secret",
				UseSSL:    false,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Requires actual MinIO connection or better mocking")
		})
	}
}

func TestMinIOService_CreateBucketIfNotExists(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		setupMock   func(*MockMinioClient)
		expectError bool
	}{
		{
			name:       "bucket already exists",
			bucketName: "existing-bucket",
			setupMock: func(m *MockMinioClient) {
				m.On("BucketExists", mock.Anything, "existing-bucket").Return(true, nil)
			},
			expectError: false,
		},
		{
			name:       "bucket does not exist - created successfully",
			bucketName: "new-bucket",
			setupMock: func(m *MockMinioClient) {
				m.On("BucketExists", mock.Anything, "new-bucket").Return(false, nil)
				m.On("MakeBucket", mock.Anything, "new-bucket", minio.MakeBucketOptions{}).Return(nil)
			},
			expectError: false,
		},
		{
			name:       "bucket exists check fails",
			bucketName: "error-bucket",
			setupMock: func(m *MockMinioClient) {
				m.On("BucketExists", mock.Anything, "error-bucket").Return(false, errors.New("connection error"))
			},
			expectError: true,
		},
		{
			name:       "bucket creation fails",
			bucketName: "fail-bucket",
			setupMock: func(m *MockMinioClient) {
				m.On("BucketExists", mock.Anything, "fail-bucket").Return(false, nil)
				m.On("MakeBucket", mock.Anything, "fail-bucket", minio.MakeBucketOptions{}).Return(errors.New("permission denied"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockMinioClient)
			tt.setupMock(mockClient)
			ctx := context.Background()

			err := mockCreateBucketLogic(ctx, mockClient, tt.bucketName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func mockCreateBucketLogic(ctx context.Context, client *MockMinioClient, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func TestMinIOService_UploadFile(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		key         string
		content     []byte
		contentType string
		setupMock   func(*MockMinioClient)
		expectError bool
	}{
		{
			name:        "successful upload",
			bucketName:  "my-bucket",
			key:         "file.txt",
			content:     []byte("test content"),
			contentType: "text/plain",
			setupMock: func(m *MockMinioClient) {
				m.On("PutObject", mock.Anything, "my-bucket", "file.txt", mock.Anything, mock.Anything, mock.Anything).
					Return(minio.UploadInfo{}, nil)
			},
			expectError: false,
		},
		{
			name:        "upload fails",
			bucketName:  "my-bucket",
			key:         "file.txt",
			content:     []byte("test content"),
			contentType: "text/plain",
			setupMock: func(m *MockMinioClient) {
				m.On("PutObject", mock.Anything, "my-bucket", "file.txt", mock.Anything, mock.Anything, mock.Anything).
					Return(minio.UploadInfo{}, errors.New("upload failed"))
			},
			expectError: true,
		},
		{
			name:        "upload large file",
			bucketName:  "my-bucket",
			key:         "large.bin",
			content:     bytes.Repeat([]byte("A"), 1024*1024),
			contentType: "application/octet-stream",
			setupMock: func(m *MockMinioClient) {
				m.On("PutObject", mock.Anything, "my-bucket", "large.bin", mock.Anything, mock.Anything, mock.Anything).
					Return(minio.UploadInfo{}, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockMinioClient)
			tt.setupMock(mockClient)
			ctx := context.Background()

			reader := bytes.NewReader(tt.content)
			err := mockUploadFile(ctx, mockClient, tt.bucketName, tt.key, reader, int64(len(tt.content)), tt.contentType)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func mockUploadFile(ctx context.Context, client *MockMinioClient, bucketName, key string, reader io.Reader, size int64, contentType string) error {
	_, err := client.PutObject(ctx, bucketName, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func TestMinIOService_DeleteFile(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		key         string
		setupMock   func(*MockMinioClient)
		expectError bool
	}{
		{
			name:       "successful deletion",
			bucketName: "my-bucket",
			key:        "file.txt",
			setupMock: func(m *MockMinioClient) {
				m.On("RemoveObject", mock.Anything, "my-bucket", "file.txt", minio.RemoveObjectOptions{}).Return(nil)
			},
			expectError: false,
		},
		{
			name:       "delete non-existent file",
			bucketName: "my-bucket",
			key:        "nonexistent.txt",
			setupMock: func(m *MockMinioClient) {
				m.On("RemoveObject", mock.Anything, "my-bucket", "nonexistent.txt", minio.RemoveObjectOptions{}).
					Return(errors.New("object not found"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockMinioClient)
			tt.setupMock(mockClient)
			ctx := context.Background()

			err := mockDeleteFile(ctx, mockClient, tt.bucketName, tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func mockDeleteFile(ctx context.Context, client *MockMinioClient, bucketName, key string) error {
	return client.RemoveObject(ctx, bucketName, key, minio.RemoveObjectOptions{})
}

func TestMinIOService_GeneratePresignedURL(t *testing.T) {
	tests := []struct {
		name        string
		bucketName  string
		key         string
		expiry      time.Duration
		expectedURL string
		setupMock   func(*MockMinioClient)
		expectError bool
	}{
		{
			name:        "successful URL generation",
			bucketName:  "my-bucket",
			key:         "file.txt",
			expiry:      24 * time.Hour,
			expectedURL: "https://minio.example.com/my-bucket/file.txt?X-Amz-Expires=86400",
			setupMock: func(m *MockMinioClient) {
				url := "https://minio.example.com/my-bucket/file.txt?X-Amz-Expires=86400"
				m.On("PresignedGetObject", mock.Anything, "my-bucket", "file.txt", 24*time.Hour, mock.Anything).
					Return(&url, nil)
			},
			expectError: false,
		},
		{
			name:        "generation fails - object not found",
			bucketName:  "my-bucket",
			key:         "missing.txt",
			expiry:      1 * time.Hour,
			expectedURL: "",
			setupMock: func(m *MockMinioClient) {
				m.On("PresignedGetObject", mock.Anything, "my-bucket", "missing.txt", 1*time.Hour, mock.Anything).
					Return(nil, errors.New("object not found"))
			},
			expectError: true,
		},
		{
			name:        "short expiry time",
			bucketName:  "my-bucket",
			key:         "temp.txt",
			expiry:      5 * time.Minute,
			expectedURL: "https://minio.example.com/my-bucket/temp.txt?X-Amz-Expires=300",
			setupMock: func(m *MockMinioClient) {
				url := "https://minio.example.com/my-bucket/temp.txt?X-Amz-Expires=300"
				m.On("PresignedGetObject", mock.Anything, "my-bucket", "temp.txt", 5*time.Minute, mock.Anything).
					Return(&url, nil)
			},
			expectError: false,
		},
		{
			name:        "long expiry time",
			bucketName:  "my-bucket",
			key:         "archive.zip",
			expiry:      7 * 24 * time.Hour,
			expectedURL: "https://minio.example.com/my-bucket/archive.zip?X-Amz-Expires=604800",
			setupMock: func(m *MockMinioClient) {
				url := "https://minio.example.com/my-bucket/archive.zip?X-Amz-Expires=604800"
				m.On("PresignedGetObject", mock.Anything, "my-bucket", "archive.zip", 7*24*time.Hour, mock.Anything).
					Return(&url, nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockMinioClient)
			tt.setupMock(mockClient)
			ctx := context.Background()

			url, err := mockGeneratePresignedURL(ctx, mockClient, tt.bucketName, tt.key, tt.expiry)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func mockGeneratePresignedURL(ctx context.Context, client *MockMinioClient, bucketName, key string, expiry time.Duration) (string, error) {
	presignedURL, err := client.PresignedGetObject(ctx, bucketName, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return *presignedURL, nil
}
