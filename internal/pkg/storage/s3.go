package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *s3.Client
}

func NewS3(client *s3.Client) *S3 {
	return &S3{client: client}
}

func (s *S3) SaveFile(bucket string, key string, file multipart.File) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	return err
}

func (s *S3) GetFile(bucket string, key string) (io.ReadCloser, *int64, string, error) {
	return s.GetFileContext(context.Background(), bucket, key)
}

func (s *S3) GetFileContext(ctx context.Context, bucket string, key string) (io.ReadCloser, *int64, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, nil, "", err
	}
	return result.Body, result.ContentLength, aws.ToString(result.ContentType), nil
}

func (s *S3) Download(ctx context.Context, bucket string, key string, destination string) error {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	return err
}

func (s *S3) Upload(ctx context.Context, bucket string, key string, source string, contentType string) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	return nil
}

func (s *S3) List(ctx context.Context, bucket string, prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	objects := make([]string, 0, len(result.Contents))
	for _, item := range result.Contents {
		if item.Key == nil || aws.ToString(item.Key) == prefix {
			continue
		}
		objects = append(objects, aws.ToString(item.Key))
	}
	return objects, nil
}
