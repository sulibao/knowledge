package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/sulibao/knowledge/internal/config"
)

type FileHandler struct {
	MinioClient *minio.Client
	Config      *config.Config
}

func NewFileHandler(minioClient *minio.Client, cfg *config.Config) *FileHandler {
	return &FileHandler{MinioClient: minioClient, Config: cfg}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为JSON
	w.Header().Set("Content-Type", "application/json")

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file from form: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "无法获取上传的文件", "message": err.Error()})
		return
	}
	defer file.Close()

	// 记录上传信息
	log.Printf("File upload started: %s, size: %d bytes, content-type: %s",
		header.Filename, header.Size, header.Header.Get("Content-Type"))

	// 检查文件大小
	if header.Size > 1024*1024*1024 { // 1GB
		log.Printf("File too large: %s (%d bytes)", header.Filename, header.Size)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "文件太大", "message": "上传的文件不能超过1GB"})
		return
	}

	objectName := filepath.Base(header.Filename)
	contentType := header.Header.Get("Content-Type")

	// 记录开始上传到MinIO的时间
	startTime := time.Now()
	log.Printf("Starting MinIO upload for %s", objectName)

	ctx := context.Background()
	info, err := h.MinioClient.PutObject(ctx, h.Config.Minio.BucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("Error uploading file to MinIO: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "上传文件到存储服务失败", "message": err.Error()})
		return
	}

	// 计算上传耗时
	uploadDuration := time.Since(startTime)
	log.Printf("Successfully uploaded %s of size %d bytes in %v",
		objectName, info.Size, uploadDuration)

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "文件上传成功",
		"filename": objectName,
		"size":     fmt.Sprintf("%d", info.Size),
		"duration": uploadDuration.String(),
	})
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为JSON
	w.Header().Set("Content-Type", "application/json")

	// 记录开始获取文件列表的时间
	startTime := time.Now()
	log.Printf("Starting to list files from bucket: %s", h.Config.Minio.BucketName)

	ctx := context.Background()
	objectsCh := h.MinioClient.ListObjects(ctx, h.Config.Minio.BucketName, minio.ListObjectsOptions{Recursive: true})

	var files []map[string]interface{}
	var errorOccurred bool

	for object := range objectsCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			errorOccurred = true
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "获取文件列表失败",
				"message": object.Err.Error(),
			})
			return
		}

		files = append(files, map[string]interface{}{
			"name":         object.Key,
			"size":         object.Size,
			"lastModified": object.LastModified,
		})
	}

	// 计算获取文件列表耗时
	listDuration := time.Since(startTime)

	// 如果没有发生错误，返回文件列表
	if !errorOccurred {
		log.Printf("Successfully listed %d files in %v", len(files), listDuration)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(files)
	}
}

func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	object, err := h.MinioClient.GetObject(ctx, h.Config.Minio.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error getting object from MinIO: %v", err)
		http.Error(w, "File not found or error retrieving file", http.StatusNotFound)
		return
	}
	defer object.Close()

	stat, err := object.Stat()
	if err != nil {
		log.Printf("Error stating object: %v", err)
		http.Error(w, "Error retrieving file info", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", stat.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	if _, err := io.Copy(w, object); err != nil {
		log.Printf("Error copying file to response: %v", err)
		http.Error(w, "Error downloading file", http.StatusInternalServerError)
		return
	}
}

func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := h.MinioClient.RemoveObject(ctx, h.Config.Minio.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Error deleting object from MinIO: %v", err)
		http.Error(w, "Error deleting file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted successfully", "filename": filename})
}
