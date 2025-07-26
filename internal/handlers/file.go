package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"go-pro/0602/internal/config"

	"github.com/minio/minio-go/v7"
)

type FileHandler struct {
	MinioClient *minio.Client
	Config      *config.Config
}

func NewFileHandler(minioClient *minio.Client, cfg *config.Config) *FileHandler {
	return &FileHandler{MinioClient: minioClient, Config: cfg}
}

func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	objectName := filepath.Base(header.Filename)
	contentType := header.Header.Get("Content-Type")

	ctx := context.Background()
	info, err := h.MinioClient.PutObject(ctx, h.Config.Minio.BucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		log.Printf("Error uploading file to MinIO: %v", err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File uploaded successfully", "filename": objectName})
}

func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	objectsCh := h.MinioClient.ListObjects(ctx, h.Config.Minio.BucketName, minio.ListObjectsOptions{Recursive: true})

	var files []string
	for object := range objectsCh {
		if object.Err != nil {
			log.Printf("Error listing object: %v", object.Err)
			continue
		}
		files = append(files, object.Key)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
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
