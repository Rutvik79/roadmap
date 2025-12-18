package s3helper

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucketName string
}

// NewS3Client creates a new S3 client
func NewS3Client(bucketName string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	return &S3Client{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

// DownloadFile downloads a file from S3
func (s *S3Client) DownloadFile(ctx context.Context, key string, writer io.WriterAt) error {
	_, err := s.downloader.Download(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// ListFiles lists all files in the bucket
func (s *S3Client) ListFiles(ctx context.Context) ([]string, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var files []string
	for _, obj := range result.Contents {
		files = append(files, *obj.Key)
	}
	return files, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GeneratePresignedURL generates a pre-signed URL for downloading
func (s *S3Client) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// GetFileMetadata retrieves file metadata
func (s *S3Client) GetFileMetadata(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	return result, nil
}
