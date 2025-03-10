package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/Zaprit/LBPSearch/pkg/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"path"
)

type S3Storage struct {
	client        *s3.Client
	preSignClient *s3.PresignClient
	storageBucket string
	cacheBucket   string
}

func (s *S3Storage) HasLevel(ctx context.Context, id string) (bool, error) {
	return s.hasResource(ctx, s.cacheBucket, path.Join("levels", id+".zip"))
}

func (s *S3Storage) PutLevel(ctx context.Context, id string, r io.Reader) error {
	return s.putResource(ctx, path.Join("levels", id+".zip"), r)
}

func (s *S3Storage) GetLevelURL(ctx context.Context, id string) (string, error) {
	return s.getObjectURL(ctx, path.Join("levels", id+".zip"))
}

func NewS3StorageBackend(cfg *config.Config) (*S3Storage, error) {
	var endpoint *string
	if cfg.S3Endpoint != "" {
		endpoint = aws.String(cfg.S3Endpoint)
	}

	client := s3.New(s3.Options{
		AppID:              "LBPSearch",
		BaseEndpoint:       endpoint,
		Credentials:        credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "LBPSearch"),
		DefaultsMode:       "",
		ExpressCredentials: nil,
		HTTPSignerV4:       nil,
		Region:             cfg.S3Region,
	})

	psClient := s3.NewPresignClient(client)

	return &S3Storage{
		client:        client,
		preSignClient: psClient,
		storageBucket: cfg.ArchiveBucket,
		cacheBucket:   cfg.CacheBucket,
	}, nil
}

func (s *S3Storage) hasResource(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3Storage) HasResource(ctx context.Context, key string) (bool, error) {
	return s.hasResource(ctx, s.storageBucket, key)
}

func (s *S3Storage) putResource(ctx context.Context, key string, src io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.cacheBucket),
		Key:    aws.String(key),
		Body:   src,
	})
	return err
}

// GetObjectURL gets a pre-signed object url so the user can download the file
func (s *S3Storage) getObjectURL(ctx context.Context, key string) (string, error) {
	obj, err := s.preSignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cacheBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	return obj.URL, nil
}

func (s *S3Storage) GetResource(ctx context.Context, path string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.storageBucket),
		Key:    aws.String(FormatURL(path)),
	})
	if err != nil {
		return nil, err
	}
	return obj.Body, nil
}

func (s *S3Storage) HasIcon(ctx context.Context, hash string) (bool, error) {
	return s.hasResource(ctx, s.cacheBucket, path.Join("icons/", hash))
}

func (s *S3Storage) PutIcon(ctx context.Context, hash string, r io.Reader) error {
	return s.putResource(ctx, path.Join("icons/", hash), r)
}

func (s *S3Storage) GetIconURL(ctx context.Context, hash string) (string, error) {
	return s.getObjectURL(ctx, path.Join("icons/", hash))
}

// FormatURL formats a URL in the same way that the LBP archive does
// e.g. the resource 1234567890abcdef1234 will be formatted as 12/34/1234567890abcdef1234
func FormatURL(path string) string {
	return fmt.Sprintf("%s/%s/%s", path[0:2], path[2:4], path)
}
