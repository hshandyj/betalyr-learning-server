package handler

import (
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MediaHandler 定义媒体处理器接口
type MediaHandler interface {
	// 上传媒体文件
	UploadMedia(c *gin.Context)
	// 获取媒体文件URL
	GetMediaURL(c *gin.Context)
	// 删除媒体文件
	DeleteMedia(c *gin.Context)
}

// mediaHandler 实现媒体处理器接口
type mediaHandler struct {
	repo repository.MediaRepository
}

// NewMediaHandler 创建新的媒体处理器实例
func NewMediaHandler(repo repository.MediaRepository) MediaHandler {
	return &mediaHandler{
		repo: repo,
	}
}

// UploadMedia 处理媒体文件上传请求
func (h *mediaHandler) UploadMedia(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("User ID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 获取上传的文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("Failed to get uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get uploaded file"})
		return
	}
	defer file.Close()

	// 获取文件信息
	fileName := fileHeader.Filename
	fileSize := fileHeader.Size
	contentType := fileHeader.Header.Get("Content-Type")

	// 如果未提供Content-Type，尝试根据文件名推断
	if contentType == "" {
		contentType = inferContentType(fileName)
	}

	// 记录上传信息
	logger.Info("Starting to upload media file",
		zap.String("userID", userID),
		zap.String("fileName", fileName),
		zap.Int64("fileSize", fileSize),
		zap.String("contentType", contentType))

	// 上传文件
	fileURL, err := h.repo.UploadMedia(file, fileSize, fileName, contentType)
	if err != nil {
		logger.Error("Failed to upload media file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// 返回上传成功结果
	c.JSON(http.StatusOK, gin.H{
		"url":         fileURL,
		"fileName":    fileName,
		"fileSize":    fileSize,
		"contentType": contentType,
	})
}

// GetMediaURL 获取媒体文件URL
func (h *mediaHandler) GetMediaURL(c *gin.Context) {
	fileKey := c.Param("key")
	if fileKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File key cannot be empty"})
		return
	}

	// 检查文件是否存在
	exists, err := h.repo.MediaExists(fileKey)
	if err != nil {
		logger.Error("Failed to check media file existence", zap.Error(err), zap.String("fileKey", fileKey))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 获取URL
	url, err := h.repo.GetMediaURL(fileKey)
	if err != nil {
		logger.Error("Failed to get media URL", zap.Error(err), zap.String("fileKey", fileKey))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// DeleteMedia 删除媒体文件
func (h *mediaHandler) DeleteMedia(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("User ID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileKey := c.Param("key")
	if fileKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File key cannot be empty"})
		return
	}

	// 检查文件是否存在
	exists, err := h.repo.MediaExists(fileKey)
	if err != nil {
		logger.Error("Failed to check media file existence", zap.Error(err), zap.String("fileKey", fileKey))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 记录删除操作
	logger.Info("Deleting media file", zap.String("userID", userID), zap.String("fileKey", fileKey))

	// 删除文件
	if err := h.repo.DeleteMedia(fileKey); err != nil {
		logger.Error("Failed to delete media file", zap.Error(err), zap.String("fileKey", fileKey))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 根据文件名推断内容类型
func inferContentType(fileName string) string {
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	switch ext {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "mp4":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "mov":
		return "video/quicktime"
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "ogg":
		return "audio/ogg"
	case "pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
