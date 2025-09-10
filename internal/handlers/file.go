// package handlers 提供了处理HTTP请求的处理器
package handlers

// 导入所需的包
import (
	"context"       // 用于控制请求的上下文
	"encoding/json" // 用于JSON编解码
	"fmt"           // 用于格式化输出
	"io"            // 提供I/O原语
	"log"           // 用于日志记录
	"net/http"      // 提供HTTP客户端和服务器实现
	"path/filepath" // 用于处理文件路径
	"time"          // 用于时间相关操作

	"github.com/minio/minio-go/v7"                 // MinIO客户端
	"github.com/sulibao/knowledge/internal/config" // 配置管理
)

// FileHandler 结构体用于处理文件相关的HTTP请求
// 包含MinIO客户端和应用配置
type FileHandler struct {
	MinioClient *minio.Client  // MinIO客户端，用于对象存储操作
	Config      *config.Config // 应用配置信息
}

// NewFileHandler 创建并返回一个新的FileHandler实例
//   - minioClient: MinIO客户端实例，用于与对象存储交互
//   - cfg: 应用配置信息
//   - 初始化后的FileHandler指针
func NewFileHandler(minioClient *minio.Client, cfg *config.Config) *FileHandler {
	return &FileHandler{MinioClient: minioClient, Config: cfg}
}

// UploadFile 处理文件上传请求
// 接收客户端上传的文件，验证后存储到MinIO
// - w: HTTP响应写入器
// - r: HTTP请求
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")

	// 从表单中获取上传的文件
	// FormFile返回指定名称的第一个文件、文件头和错误信息
	file, header, err := r.FormFile("file")
	if err != nil {
		// 记录错误日志
		log.Printf("Error retrieving file from form: %v", err)
		// 设置HTTP状态码为400 Bad Request
		w.WriteHeader(http.StatusBadRequest)
		// 返回错误信息给客户端
		json.NewEncoder(w).Encode(map[string]string{"error": "无法获取上传的文件", "message": err.Error()})
		return
	}
	// 确保在函数结束时关闭文件
	defer file.Close()

	// 记录文件上传开始的信息
	log.Printf("File upload started: %s, size: %d bytes, content-type: %s",
		header.Filename, header.Size, header.Header.Get("Content-Type"))

	// 检查文件大小是否超过限制(1GB)
	if header.Size > 1024*1024*1024 { // 1GB = 1024MB = 1024*1024KB = 1024*1024*1024B
		// 记录文件过大的错误
		log.Printf("文件过大: %s (%d bytes)", header.Filename, header.Size)
		// 设置HTTP状态码为400 Bad Request
		w.WriteHeader(http.StatusBadRequest)
		// 返回错误信息给客户端
		json.NewEncoder(w).Encode(map[string]string{"error": "上传的文件过大", "message": "上传的文件大小不能超过1GB"})
		return
	}

	// 获取文件名（不包含路径）
	objectName := filepath.Base(header.Filename)
	// 获取文件的内容类型
	contentType := header.Header.Get("Content-Type")

	// 记录开始上传到MinIO的时间，用于计算上传耗时
	startTime := time.Now()
	log.Printf("Starting MinIO upload for %s", objectName)

	// 创建一个空的上下文
	ctx := context.Background()
	// 调用MinIO客户端的PutObject方法上传文件
	// 参数：上下文、存储桶名称、对象名称、文件内容、文件大小、上传选项
	info, err := h.MinioClient.PutObject(ctx, h.Config.Minio.BucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType, // 设置内容类型
	})
	if err != nil {
		// 记录上传失败的错误
		log.Printf("Error uploading file to MinIO: %v", err)
		// 设置HTTP状态码为500 Internal Server Error
		w.WriteHeader(http.StatusInternalServerError)
		// 返回错误信息给客户端
		json.NewEncoder(w).Encode(map[string]string{"error": "上传文件到存储服务失败", "message": err.Error()})
		return
	}

	// 计算上传耗时（从开始上传到上传完成的时间差）
	uploadDuration := time.Since(startTime)
	// 记录上传成功的信息
	log.Printf("Successfully uploaded %s of size %d bytes in %v",
		objectName, info.Size, uploadDuration)

	// 返回成功响应给客户端
	// 设置HTTP状态码为200 OK
	w.WriteHeader(http.StatusOK)
	// 返回JSON格式的成功信息
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "文件上传成功",                     // 成功消息
		"filename": objectName,                   // 文件名
		"size":     fmt.Sprintf("%d", info.Size), // 文件大小
		"duration": uploadDuration.String(),      // 上传耗时
	})
}

// ListFiles 处理获取文件列表的请求
// 从MinIO存储桶中获取所有文件的列表并返回给客户端
// 参数:
//   - w: HTTP响应写入器
//   - r: HTTP请求
func (h *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	// 设置响应头为JSON格式
	w.Header().Set("Content-Type", "application/json")

	// 记录开始获取文件列表的时间，用于计算耗时
	startTime := time.Now()
	// 记录开始获取文件列表的日志
	log.Printf("Starting to list files from bucket: %s", h.Config.Minio.BucketName)

	// 创建一个空的上下文
	ctx := context.Background()
	// 调用MinIO客户端的ListObjects方法获取存储桶中的所有对象
	// 返回一个对象信息通道，可以通过遍历该通道获取所有对象
	// Recursive: true表示递归列出所有对象，包括子目录中的对象
	objectsCh := h.MinioClient.ListObjects(ctx, h.Config.Minio.BucketName, minio.ListObjectsOptions{Recursive: true})

	// 用于存储文件信息的切片
	var files []map[string]interface{}
	// 标记是否发生错误的标志
	var errorOccurred bool

	// 遍历对象通道，获取每个对象的信息
	for object := range objectsCh {
		// 检查对象获取过程中是否有错误
		if object.Err != nil {
			// 记录获取对象列表失败的错误
			log.Printf("Error listing objects: %v", object.Err)
			// 设置错误标志
			errorOccurred = true
			// 设置HTTP状态码为500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			// 返回错误信息给客户端
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "获取文件列表失败",
				"message": object.Err.Error(),
			})
			return
		}

		// 将对象的信息添加到文件列表中
		files = append(files, map[string]interface{}{
			"name":         object.Key,          // 文件名/键
			"size":         object.Size,         // 文件大小
			"lastModified": object.LastModified, // 最后修改时间
		})
	}

	// 计算获取文件列表的耗时（从开始获取到获取完成的时间差）
	listDuration := time.Since(startTime)

	// 如果没有发生错误，返回文件列表给客户端
	if !errorOccurred {
		// 记录获取文件列表成功的信息
		log.Printf("Successfully listed %d files in %v", len(files), listDuration)
		// 设置HTTP状态码为200 OK
		w.WriteHeader(http.StatusOK)
		// 将文件列表编码为JSON并返回给客户端
		json.NewEncoder(w).Encode(files)
	}
}

// DownloadFile 处理文件下载请求
// 从MinIO获取指定文件并发送给客户端
// 参数:
//   - w: HTTP响应写入器
//   - r: HTTP请求
func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	// 从URL查询参数中获取文件名
	filename := r.URL.Query().Get("filename")
	// 检查文件名是否为空
	if filename == "" {
		// 如果文件名为空，返回400 Bad Request错误
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 创建一个空的上下文
	ctx := context.Background()
	// 调用MinIO客户端的GetObject方法获取指定的对象
	// 参数：上下文、存储桶名称、对象名称、获取选项
	object, err := h.MinioClient.GetObject(ctx, h.Config.Minio.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		// 记录获取对象失败的错误
		log.Printf("Error getting object from MinIO: %v", err)
		// 返回404 Not Found错误给客户端
		http.Error(w, "文件不存在或获取文件时出错", http.StatusNotFound)
		return
	}
	// 确保在函数结束时关闭对象
	defer object.Close()

	// 获取对象的元数据信息
	stat, err := object.Stat()
	if err != nil {
		// 记录获取对象元数据失败的错误
		log.Printf("Error stating object: %v", err)
		// 返回500 Internal Server Error错误给客户端
		http.Error(w, "检索文件信息时出错", http.StatusInternalServerError)
		return
	}

	// 设置HTTP响应头
	// Content-Disposition: 指示浏览器将响应作为附件下载，并指定文件名
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	// Content-Type: 指定文件的MIME类型
	w.Header().Set("Content-Type", stat.ContentType)
	// Content-Length: 指定文件的大小
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	// 将对象内容复制到HTTP响应
	if _, err := io.Copy(w, object); err != nil {
		// 记录复制文件内容失败的错误
		log.Printf("Error copying file to response: %v", err)
		// 返回500 Internal Server Error错误给客户端
		http.Error(w, "下载文件时出错", http.StatusInternalServerError)
		return
	}
	// 如果没有错误，文件内容已成功发送给客户端
}

// DeleteFile 处理文件删除请求
// 从MinIO存储桶中删除指定的文件
// 参数:
//   - w: HTTP响应写入器
//   - r: HTTP请求
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	// 从URL查询参数中获取文件名
	filename := r.URL.Query().Get("filename")
	// 检查文件名是否为空
	if filename == "" {
		// 如果文件名为空，返回400 Bad Request错误
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 创建一个空的上下文
	ctx := context.Background()
	// 调用MinIO客户端的RemoveObject方法删除指定的对象
	// 参数：上下文、存储桶名称、对象名称、删除选项
	err := h.MinioClient.RemoveObject(ctx, h.Config.Minio.BucketName, filename, minio.RemoveObjectOptions{})
	if err != nil {
		// 记录删除对象失败的错误
		log.Printf("Error deleting object from MinIO: %v", err)
		// 返回500 Internal Server Error错误给客户端
		http.Error(w, "删除文件时出错", http.StatusInternalServerError)
		return
	}

	// 设置HTTP状态码为200 OK
	w.WriteHeader(http.StatusOK)
	// 返回JSON格式的成功信息
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "文件删除成功", // 成功消息
		"filename": filename, // 被删除的文件名
	})
}
