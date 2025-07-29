package minio_client

import (
	"context"
	"log"

	"github.com/sulibao/knowledge/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinio(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MinIO!")

	// Check if the bucket exists
	ctx := context.Background()
	found, err := minioClient.BucketExists(ctx, cfg.Minio.BucketName)
	if err != nil {
		return nil, err
	}

	if !found {
		log.Printf("Bucket '%s' not found, creating it...", cfg.Minio.BucketName)
		err = minioClient.MakeBucket(ctx, cfg.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Printf("Bucket '%s' created successfully.", cfg.Minio.BucketName)
	} else {
		log.Printf("Bucket '%s' already exists.", cfg.Minio.BucketName)
	}

	return minioClient, nil
}
