// package minio_client 提供MinIO对象存储服务的客户端功能
package minio_client

import (
	"context" // 提供上下文控制
	"log"     // 提供日志记录

	"github.com/sulibao/knowledge/internal/config" // 导入配置包

	"github.com/minio/minio-go/v7"                 // MinIO Go客户端
	"github.com/minio/minio-go/v7/pkg/credentials" // MinIO凭证管理
)

// InitMinio 初始化MinIO客户端并确保存储桶存在
//   - 创建并配置MinIO客户端连接
//   - 检查指定的存储桶是否存在
//   - 如果存储桶不存在，则创建它
//   - cfg: 包含MinIO配置信息的配置对象
//   - *minio.Client: MinIO客户端对象
//   - error: 如果初始化过程中发生错误，返回相应的错误信息
func InitMinio(cfg *config.Config) (*minio.Client, error) {
	// 创建MinIO客户端实例
	minioClient, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		// 设置访问凭证
		Creds: credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		// 是否使用SSL连接
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MinIO!")

	// 创建上下文对象
	ctx := context.Background()
	// 检查存储桶是否存在
	found, err := minioClient.BucketExists(ctx, cfg.Minio.BucketName)
	if err != nil {
		return nil, err
	}

	if !found {
		// 存储桶不存在，创建新的存储桶
		log.Printf("Bucket '%s' not found, creating it...", cfg.Minio.BucketName)
		err = minioClient.MakeBucket(ctx, cfg.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Printf("Bucket '%s' created successfully.", cfg.Minio.BucketName)
	} else {
		// 存储桶已存在
		log.Printf("Bucket '%s' already exists.", cfg.Minio.BucketName)
	}

	return minioClient, nil // 返回初始化好的MinIO客户端
}
