package handler

import (
	"betalyr-learning-server/internal/models"
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
	"go.uber.org/zap"
)

// MediaHandler 定义媒体处理器接口
type MediaHandler interface {
	// 上传视频文件
	UploadVideo(c *gin.Context)
	// 上传音频文件
	UploadAudio(c *gin.Context)
	// 删除媒体文件
	DeleteMedia(c *gin.Context)
	// 获取视频列表
	GetVideos(c *gin.Context)
	// 获取视频详情
	GetVideoDetail(c *gin.Context)
	// 获取音频列表
	GetAudios(c *gin.Context)
	// 获取音频详情
	GetAudioDetail(c *gin.Context)
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

// DeleteMedia 删除媒体文件
func (h *mediaHandler) DeleteMedia(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("User ID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	mediaID := c.Param("id")
	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media ID cannot be empty"})
		return
	}

	// 根据ID获取媒体信息
	media, err := h.repo.GetMediaByID(mediaID)
	if err != nil {
		logger.Error("Failed to get media by ID", zap.Error(err), zap.String("mediaID", mediaID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if media == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// 检查权限：只有上传者可以删除自己的媒体文件
	if media.UploaderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 记录删除操作
	logger.Info("Deleting media",
		zap.String("userID", userID),
		zap.String("mediaID", mediaID),
		zap.String("fileKey", media.FileKey))

	// 完全删除媒体（包括数据库记录和文件）
	if err := h.repo.DeleteMediaCompletely(mediaID); err != nil {
		logger.Error("Failed to delete media completely", zap.Error(err), zap.String("mediaID", mediaID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Media deleted successfully",
	})
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

// GetVideos 获取公开视频列表 - 前端调用 /public/media/video
func (h *mediaHandler) GetVideos(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	// 获取视频列表
	videos, err := h.repo.GetVideos(page, limit)
	if err != nil {
		logger.Error("Failed to get videos", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get videos"})
		return
	}

	// 转换为列表格式
	videoList := make([]models.PublicVideoList, len(videos))
	for i, video := range videos {
		videoList[i] = video.ToPublicVideoList()
	}

	// 按前端期望的格式返回
	c.JSON(http.StatusOK, videoList)
}

// GetVideoDetail 获取视频详情
func (h *mediaHandler) GetVideoDetail(c *gin.Context) {
	videoID := c.Param("id")
	if videoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video ID cannot be empty"})
		return
	}

	// 根据ID获取媒体信息
	media, err := h.repo.GetMediaByID(videoID)
	if err != nil {
		logger.Error("Failed to get video by ID", zap.Error(err), zap.String("videoID", videoID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if media == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	// 检查是否为视频类型
	if media.MediaType != models.MediaTypeVideo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media is not a video"})
		return
	}

	// 返回视频详情
	videoDetail := media.ToVideoDetail()
	c.JSON(http.StatusOK, videoDetail)
}

// UploadVideo 处理视频文件上传请求
func (h *mediaHandler) UploadVideo(c *gin.Context) {
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

	// 验证文件类型
	if !strings.Contains(contentType, "video") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not a video"})
		return
	}

	// 获取标题、描述和分类（可选）
	title := c.PostForm("title")
	if title == "" {
		title = fileName
	}

	description := c.PostForm("description")
	category := c.PostForm("category")
	if category == "" {
		category = "其他"
	}

	logger.Info("Starting to upload video file",
		zap.String("userID", userID),
		zap.String("fileName", fileName),
		zap.Int64("fileSize", fileSize),
		zap.String("contentType", contentType))

	// 保存临时文件用于处理
	tempDir := os.TempDir()
	tempVideoPath := filepath.Join(tempDir, fmt.Sprintf("video_%s_%s", uuid.New().String(), fileName))

	tempFile, err := os.Create(tempVideoPath)
	if err != nil {
		logger.Error("Failed to create temp file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer os.Remove(tempVideoPath)
	defer tempFile.Close()

	// 复制文件内容到临时文件
	_, err = io.Copy(tempFile, file)
	if err != nil {
		logger.Error("Failed to copy file to temp", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	tempFile.Close()

	// 重新打开文件用于上传
	file.Seek(0, 0)

	// 上传原视频文件到存储
	fileURL, err := h.repo.UploadMedia(file, fileSize, fileName, contentType)
	if err != nil {
		logger.Error("Failed to upload video file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// 从URL中提取文件键
	fileKey := strings.TrimPrefix(fileURL, strings.Split(fileURL, "/")[0]+"//"+strings.Split(fileURL, "/")[2]+"/")

	// 提取视频帧
	previewURL, thumbnailURL, err := h.extractVideoFrames(tempVideoPath, fileName)
	if err != nil {
		logger.Error("Failed to extract video frames", zap.Error(err))
		// 不阻断上传，继续保存视频记录
	}

	// 生成媒体记录ID
	mediaID := uuid.New().String()

	// 创建媒体记录
	media := &models.Media{
		ID:         mediaID,
		UploaderID: userID,
		Title:      title,
		Description: func() *string {
			if description == "" {
				return nil
			} else {
				return &description
			}
		}(),
		FileName:    fileName,
		FileKey:     fileKey,
		FileURL:     fileURL,
		FileSize:    fileSize,
		ContentType: contentType,
		MediaType:   models.MediaTypeVideo,
		Category:    category,
		Status:      models.MediaStatusReady,
		Preview:     previewURL,
		Thumbnail:   thumbnailURL,
	}

	// 保存媒体记录到数据库
	if err := h.repo.CreateMedia(media); err != nil {
		logger.Error("Failed to create media record", zap.Error(err))
	}

	// 返回上传成功结果
	c.JSON(http.StatusOK, gin.H{
		"id":          mediaID,
		"url":         fileURL,
		"fileName":    fileName,
		"fileSize":    fileSize,
		"contentType": contentType,
		"category":    category,
		"preview":     previewURL,
		"thumbnail":   thumbnailURL,
		"message":     "Video upload successful",
	})
}

// UploadAudio 处理音频文件上传请求
func (h *mediaHandler) UploadAudio(c *gin.Context) {
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

	// 验证文件类型
	if !strings.Contains(contentType, "audio") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is not an audio"})
		return
	}

	// 获取标题（音频只需要标题）
	title := c.PostForm("title")
	if title == "" {
		title = fileName
	}

	logger.Info("Starting to upload audio file",
		zap.String("userID", userID),
		zap.String("fileName", fileName),
		zap.Int64("fileSize", fileSize),
		zap.String("contentType", contentType))

	// 上传音频文件到存储
	fileURL, err := h.repo.UploadMedia(file, fileSize, fileName, contentType)
	if err != nil {
		logger.Error("Failed to upload audio file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// 从URL中提取文件键
	fileKey := strings.TrimPrefix(fileURL, strings.Split(fileURL, "/")[0]+"//"+strings.Split(fileURL, "/")[2]+"/")

	// 生成媒体记录ID
	mediaID := uuid.New().String()

	// 创建媒体记录
	media := &models.Media{
		ID:          mediaID,
		UploaderID:  userID,
		Title:       title,
		FileName:    fileName,
		FileKey:     fileKey,
		FileURL:     fileURL,
		FileSize:    fileSize,
		ContentType: contentType,
		MediaType:   models.MediaTypeAudio,
		Category:    "音频", // 音频默认分类
		Status:      models.MediaStatusReady,
	}

	// 保存媒体记录到数据库
	if err := h.repo.CreateMedia(media); err != nil {
		logger.Error("Failed to create media record", zap.Error(err))
	}

	// 返回上传成功结果
	c.JSON(http.StatusOK, gin.H{
		"id":          mediaID,
		"url":         fileURL,
		"fileName":    fileName,
		"fileSize":    fileSize,
		"contentType": contentType,
		"message":     "Audio upload successful",
	})
}

// extractVideoFrames 提取视频第一帧并生成预览图和缩略图
func (h *mediaHandler) extractVideoFrames(videoPath, originalFileName string) (*string, *string, error) {
	tempDir := os.TempDir()

	// 生成输出文件名
	baseFileName := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	previewPath := filepath.Join(tempDir, fmt.Sprintf("preview_%s_%s.jpg", baseFileName, uuid.New().String()))
	thumbnailPath := filepath.Join(tempDir, fmt.Sprintf("thumbnail_%s_%s.jpg", baseFileName, uuid.New().String()))

	defer os.Remove(previewPath)
	defer os.Remove(thumbnailPath)

	// 添加调试信息
	logger.Info("Extracting video frames",
		zap.String("videoPath", videoPath),
		zap.String("previewPath", previewPath),
		zap.String("thumbnailPath", thumbnailPath))

	// 提取第一帧 (高质量预览图) - 简化命令
	err := ffmpeg_go.Input(videoPath).
		Output(previewPath, ffmpeg_go.KwArgs{
			"vframes": 1,
			"ss":      "00:00:01", // 从第1秒开始提取，避免全黑帧
			"q:v":     2,          // 高质量
			"vf":      "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2:black",
		}).
		OverWriteOutput().
		Run()

	if err != nil {
		logger.Error("Failed to extract preview frame", zap.Error(err))
		// 尝试更简单的命令
		err = ffmpeg_go.Input(videoPath).
			Output(previewPath, ffmpeg_go.KwArgs{
				"vframes": 1,
				"ss":      "00:00:01",
			}).
			OverWriteOutput().
			Run()

		if err != nil {
			logger.Error("Failed to extract preview frame with simple command", zap.Error(err))
			return nil, nil, err
		}
	}

	// 提取第一帧 (缩略图) - 简化命令
	err = ffmpeg_go.Input(videoPath).
		Output(thumbnailPath, ffmpeg_go.KwArgs{
			"vframes": 1,
			"ss":      "00:00:01", // 从第1秒开始提取，避免全黑帧
			"q:v":     8,          // 较低质量，用于缩略图
			"vf":      "scale=320:180:force_original_aspect_ratio=decrease,pad=320:180:(ow-iw)/2:(oh-ih)/2:black",
		}).
		OverWriteOutput().
		Run()

	if err != nil {
		logger.Error("Failed to extract thumbnail frame", zap.Error(err))
		// 尝试更简单的命令
		err = ffmpeg_go.Input(videoPath).
			Output(thumbnailPath, ffmpeg_go.KwArgs{
				"vframes": 1,
				"ss":      "00:00:01",
				"vf":      "scale=320:180",
			}).
			OverWriteOutput().
			Run()

		if err != nil {
			logger.Error("Failed to extract thumbnail frame with simple command", zap.Error(err))
			return nil, nil, err
		}
	}

	// 上传预览图到R2
	previewURL, err := h.uploadImageToR2(previewPath, fmt.Sprintf("preview_%s.jpg", baseFileName))
	if err != nil {
		logger.Error("Failed to upload preview image", zap.Error(err))
		return nil, nil, err
	}

	// 上传缩略图到R2
	thumbnailURL, err := h.uploadImageToR2(thumbnailPath, fmt.Sprintf("thumbnail_%s.jpg", baseFileName))
	if err != nil {
		logger.Error("Failed to upload thumbnail image", zap.Error(err))
		return nil, nil, err
	}

	return previewURL, thumbnailURL, nil
}

// uploadImageToR2 上传图片到R2存储
func (h *mediaHandler) uploadImageToR2(imagePath, fileName string) (*string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// 上传到R2
	imageURL, err := h.repo.UploadMedia(file, fileInfo.Size(), fileName, "image/jpeg")
	if err != nil {
		return nil, err
	}

	return &imageURL, nil
}

// GetAudios 获取公开音频列表 - 前端调用 /public/media/audio
func (h *mediaHandler) GetAudios(c *gin.Context) {
	// 获取音频列表 (不需要分页参数，前端未使用)
	audios, err := h.repo.GetAudios(1, 100) // 获取前100个音频
	if err != nil {
		logger.Error("Failed to get audios", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audios"})
		return
	}

	// 转换为列表格式
	audioList := make([]models.PublicAudioList, len(audios))
	for i, audio := range audios {
		audioList[i] = audio.ToPublicAudioList()
	}

	// 按前端期望的格式返回
	c.JSON(http.StatusOK, audioList)
}

// GetAudioDetail 获取音频详情
func (h *mediaHandler) GetAudioDetail(c *gin.Context) {
	audioID := c.Param("id")
	if audioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio ID cannot be empty"})
		return
	}

	// 根据ID获取媒体信息
	media, err := h.repo.GetMediaByID(audioID)
	if err != nil {
		logger.Error("Failed to get audio by ID", zap.Error(err), zap.String("audioID", audioID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if media == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audio not found"})
		return
	}

	// 检查是否为音频类型
	if media.MediaType != models.MediaTypeAudio {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media is not an audio"})
		return
	}

	// 返回音频详情
	audioDetail := media.ToAudioDetail()
	c.JSON(http.StatusOK, audioDetail)
}
