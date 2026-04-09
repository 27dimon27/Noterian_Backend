package minio

import (
	"context"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	client *minio.Client
}

func NewMinIOService(cfg config.MinIOConfig) (*MinIOService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOService{
		client: client,
	}, nil
}

func (s *MinIOService) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MinIOService) UploadFile(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(
		ctx,
		bucketName,
		key,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	return err
}

func (s *MinIOService) DeleteFile(ctx context.Context, bucketName, key string) error {
	return s.client.RemoveObject(ctx, bucketName, key, minio.RemoveObjectOptions{})
}

func (s *MinIOService) GeneratePresignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
