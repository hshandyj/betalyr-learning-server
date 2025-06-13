package repository

import (
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/models"
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MediaRepository 定义媒体存储库接口
type MediaRepository interface {
	// 文件存储相关操作
	// 上传媒体文件，返回文件URL
	UploadMedia(file io.Reader, fileSize int64, fileName, contentType string) (string, error)
	// 删除媒体文件
	DeleteMedia(fileKey string) error

	// 数据库相关操作
	// 创建媒体记录
	CreateMedia(media *models.Media) error
	// 根据ID获取媒体信息
	GetMediaByID(id string) (*models.Media, error)
	// 完全删除媒体（包括文件和数据库记录）
	DeleteMediaCompletely(id string) error
	// 获取视频列表
	GetVideos(page, limit int) ([]models.Media, error)
	// 获取音频列表
	GetAudios(page, limit int) ([]models.Media, error)
}

// mediaRepository 实现媒体存储库接口
type mediaRepository struct {
	client    *s3.Client
	bucket    string
	publicURL string
	db        *gorm.DB
}

// NewMediaRepository 创建新的媒体存储库实例
func NewMediaRepository() MediaRepository {
	return &mediaRepository{
		client:    storage.R2Client,
		bucket:    storage.R2Bucket,
		publicURL: storage.R2PublicURL,
		db:        database.DB,
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

// 数据库相关操作方法

// CreateMedia 创建媒体记录
func (r *mediaRepository) CreateMedia(media *models.Media) error {
	return r.db.Create(media).Error
}

// GetMediaByID 根据ID获取媒体信息
func (r *mediaRepository) GetMediaByID(id string) (*models.Media, error) {
	var media models.Media
	result := r.db.Where("id = ?", id).First(&media)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到记录返回nil而不是错误
		}
		return nil, result.Error
	}
	return &media, nil
}

// GetVideos 获取视频列表
func (r *mediaRepository) GetVideos(page, limit int) ([]models.Media, error) {
	var videos []models.Media
	offset := (page - 1) * limit

	// 查询视频，按创建时间降序排序
	result := r.db.Where("media_type = ? AND status = ?", models.MediaTypeVideo, models.MediaStatusReady).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&videos)

	if result.Error != nil {
		return nil, result.Error
	}

	return videos, nil
}

// GetAudios 获取音频列表
func (r *mediaRepository) GetAudios(page, limit int) ([]models.Media, error) {
	var audios []models.Media
	offset := (page - 1) * limit

	// 查询音频，按创建时间降序排序
	result := r.db.Where("media_type = ? AND status = ?", models.MediaTypeAudio, models.MediaStatusReady).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&audios)

	if result.Error != nil {
		return nil, result.Error
	}

	return audios, nil
}

// DeleteMediaCompletely 完全删除媒体（包括文件和数据库记录）
func (r *mediaRepository) DeleteMediaCompletely(id string) error {
	// 先获取媒体信息
	media, err := r.GetMediaByID(id)
	if err != nil {
		return err
	}

	if media == nil {
		return fmt.Errorf("media not found")
	}

	// 开始事务
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 删除数据库记录
	if err := tx.Where("id = ?", id).Delete(&models.Media{}).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to delete media record", zap.Error(err), zap.String("id", id))
		return err
	}

	// 提交数据库事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to commit transaction", zap.Error(err), zap.String("id", id))
		return err
	}

	// 删除存储文件（即使失败也不回滚数据库，因为文件可以手动清理）
	if media.FileKey != "" {
		if err := r.DeleteMedia(media.FileKey); err != nil {
			logger.Error("Failed to delete media file, but database record was deleted",
				zap.Error(err),
				zap.String("id", id),
				zap.String("fileKey", media.FileKey))
			// 不返回错误，因为数据库记录已经删除成功
		}
	}

	logger.Info("Media deleted completely", zap.String("id", id), zap.String("fileKey", media.FileKey))
	return nil
}
