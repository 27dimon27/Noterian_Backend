package minio

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOService struct {
	client     *minio.Client
	bucketName string
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
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *MinIOService) CreateBucketIfNotExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MinIOService) UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		key,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	return err
}

func (s *MinIOService) DeleteFile(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
}

func (s *MinIOService) GetFileInfo(ctx context.Context, key string) (minio.ObjectInfo, error) {
	return s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
}

func (s *MinIOService) ListFiles(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var files []minio.ObjectInfo

	ch := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range ch {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object)
	}

	return files, nil
}
