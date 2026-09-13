package storage

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type AWSS3StorageAdapter struct {
	Client *s3.Client
}

func NewAWSS3StorageAdapter(client *s3.Client) *AWSS3StorageAdapter {
	return &AWSS3StorageAdapter{
		Client: client,
	}
}

func (s *AWSS3StorageAdapter) SaveFile(bucket string, key string, file multipart.File) error {
	_, err := s.Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key),
		ChecksumAlgorithm: types.ChecksumAlgorithmCrc32,
		Body:              file,
	})

	return err
}
