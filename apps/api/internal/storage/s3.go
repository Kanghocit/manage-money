package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/manage-money/api/internal/config"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3(cfg config.Config) (*S3Store, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")
	client := s3.New(s3.Options{
		Region:       cfg.S3Region,
		Credentials:  creds,
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		UsePathStyle: true,
	})
	return &S3Store{client: client, bucket: cfg.S3Bucket}, nil
}

func (s *S3Store) Put(key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          r,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	return err
}

func (s *S3Store) Get(key string) (io.ReadCloser, string, error) {
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}
	ct := "application/octet-stream"
	if out.ContentType != nil {
		ct = *out.ContentType
	}
	return out.Body, ct, nil
}
