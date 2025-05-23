package repository

import (
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/storage"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MediaRepository 定义媒体存储库接口
type MediaRepository interface {
	// 上传媒体文件，返回文件URL
	UploadMedia(file io.Reader, fileSize int64, fileName, contentType string) (string, error)
	// 获取媒体直接下载URL
	GetMediaURL(fileKey string) (string, error)
	// 删除媒体文件
	DeleteMedia(fileKey string) error
	// 判断媒体文件是否存在
	MediaExists(fileKey string) (bool, error)
}

// mediaRepository 实现媒体存储库接口
type mediaRepository struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewMediaRepository 创建新的媒体存储库实例
func NewMediaRepository() MediaRepository {
	return &mediaRepository{
		client:    storage.R2Client,
		bucket:    storage.R2Bucket,
		publicURL: storage.R2PublicURL,
	}
}

// 生成唯一文件名
func generateUniqueFileName(originalFileName string) string {
	ext := filepath.Ext(originalFileName)
	baseName := strings.TrimSuffix(originalFileName, ext)
	// 使用base name + uuid + 原始扩展名
	return fmt.Sprintf("%s-%s%s", baseName, uuid.New().String(), ext)
}

// 生成仅用于内部存储的文件键
func generateInternalFileKey(fileType, originalFileName string) string {
	uniqueName := generateUniqueFileName(originalFileName)
	// 按文件类型分目录存储
	return fmt.Sprintf("%s/%s", fileType, uniqueName)
}

// UploadMedia 上传媒体文件到R2
func (r *mediaRepository) UploadMedia(file io.Reader, fileSize int64, fileName, contentType string) (string, error) {
	ctx := context.Background()

	// 根据MIME类型确定文件类型目录
	fileType := "other"
	if strings.Contains(contentType, "audio") {
		fileType = "audio"
	} else if strings.Contains(contentType, "video") {
		fileType = "video"
	} else if strings.Contains(contentType, "image") {
		fileType = "image"
	}

	fileKey := generateInternalFileKey(fileType, fileName)

	// 创建PutObject请求
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(fileKey),
		Body:          file,
		ContentLength: fileSize,
		ContentType:   aws.String(contentType),
		CacheControl:  aws.String("public, max-age=31536000"), // 缓存1年
	})

	if err != nil {
		logger.Error("Failed to upload media to R2", zap.Error(err), zap.String("fileKey", fileKey))
		return "", err
	}

	// 构建公共访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), fileKey)
	logger.Info("Media uploaded successfully", zap.String("URL", fileURL), zap.String("contentType", contentType))

	return fileURL, nil
}

// GetMediaURL 获取媒体文件的URL
func (r *mediaRepository) GetMediaURL(fileKey string) (string, error) {
	// 对于公开访问的bucket，直接返回公共URL
	if r.publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(r.publicURL, "/"), fileKey), nil
	}

	// 如果是私有bucket，生成预签名URL
	ctx := context.Background()
	presignClient := s3.NewPresignClient(r.client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(fileKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 24 * time.Hour // URL有效期24小时
	})

	if err != nil {
		logger.Error("Failed to generate presigned URL", zap.Error(err), zap.String("fileKey", fileKey))
		return "", err
	}

	return presignedReq.URL, nil
}

// DeleteMedia 从R2删除媒体文件
func (r *mediaRepository) DeleteMedia(fileKey string) error {
	if fileKey == "" {
		return fmt.Errorf("invalid file key")
	}

	ctx := context.Background()
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		logger.Error("Failed to delete media from R2", zap.Error(err), zap.String("fileKey", fileKey))
		return err
	}

	logger.Info("Media deleted successfully", zap.String("fileKey", fileKey))
	return nil
}

// MediaExists 检查媒体文件是否存在
func (r *mediaRepository) MediaExists(fileKey string) (bool, error) {
	ctx := context.Background()
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(fileKey),
	})

	if err != nil {
		// 解析错误类型，确定是否是NotFound错误
		var notFound bool
		if strings.Contains(err.Error(), "NotFound") {
			notFound = true
		}

		if notFound {
			return false, nil // 文件不存在，但不是错误
		}

		logger.Error("Failed to check media existence", zap.Error(err), zap.String("fileKey", fileKey))
		return false, err // 其他类型的错误
	}

	return true, nil // 文件存在
}
