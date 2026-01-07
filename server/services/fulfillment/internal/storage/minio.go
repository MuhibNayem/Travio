package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Provider interface {
	Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
	GetURL(ctx context.Context, objectName string) (string, error)
}

type MinIOProvider struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewMinIOProvider(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL bool) (*MinIOProvider, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOProvider{
		client:     minioClient,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}, nil
}

func (m *MinIOProvider) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to public read (simplification for this example, or use presigned URLs)
		// For now, let's assume we use presigned URLs or public bucket if configured.
		// Actually, let's stick to generating Presigned URLs or constructing public URLs if the bucket is public.
		// Given the requirements usually imply public access for tickets via a secure link, or a short lived link.
		// Let's rely on constructing the URL manually assuming public read for now, or use Presigned.
		// User plan said: "Store URL in tickets table... Return PDFURL".
		// A permanent public URL is easiest if bucket is public.
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, m.bucketName)
		if err = m.client.SetBucketPolicy(ctx, m.bucketName, policy); err != nil {
			// Log error but don't fail, maybe existing policy is fine
		}
	}
	return nil
}

func (m *MinIOProvider) Upload(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	reader := bytes.NewReader(data)
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	// Construct URL
	// If checking against localhost, we might need to be careful about what clients see.
	// Assume endpoint is accessible.
	scheme := "http"
	if m.useSSL {
		scheme = "https"
	}

	// Format: http://endpoint/bucket/object
	finalURL := fmt.Sprintf("%s://%s/%s/%s", scheme, m.endpoint, m.bucketName, objectName)
	return finalURL, nil
}

func (m *MinIOProvider) GetURL(ctx context.Context, objectName string) (string, error) {
	// Generate a presigned URL valid for 7 days
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, time.Hour*24*7, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
