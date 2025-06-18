package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// MediaType 媒体类型枚举
type MediaType string

const (
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeImage MediaType = "image"
	MediaTypeOther MediaType = "other"
)

// MediaStatus 媒体状态枚举
type MediaStatus string

const (
	MediaStatusUploading  MediaStatus = "uploading"  // 上传中
	MediaStatusProcessing MediaStatus = "processing" // 处理中
	MediaStatusReady      MediaStatus = "ready"      // 就绪
	MediaStatusError      MediaStatus = "error"      // 错误
)

// MediaMeta 媒体元数据
type MediaMeta struct {
	Duration    *int64  `json:"duration,omitempty"`    // 时长（秒），视频/音频专用
	Width       *int    `json:"width,omitempty"`       // 宽度，视频/图片专用
	Height      *int    `json:"height,omitempty"`      // 高度，视频/图片专用
	Resolution  *string `json:"resolution,omitempty"`  // 分辨率，如"1920x1080"
	Bitrate     *int64  `json:"bitrate,omitempty"`     // 比特率
	FrameRate   *string `json:"frameRate,omitempty"`   // 帧率，视频专用
	Codec       *string `json:"codec,omitempty"`       // 编码格式
	AspectRatio *string `json:"aspectRatio,omitempty"` // 宽高比，如"16:9"
}

// Value 实现driver.Valuer接口
func (m MediaMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan 实现sql.Scanner接口
func (m *MediaMeta) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &m)
}

// Media 媒体文件模型
type Media struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	UploaderID  string          `gorm:"index" json:"uploaderId"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	FileName    string          `json:"fileName"`
	FileKey     string          `gorm:"uniqueIndex" json:"fileKey"` // 存储的文件键
	FileURL     string          `json:"fileURL"`                    // 公开访问URL
	FileSize    int64           `json:"fileSize"`                   // 文件大小（字节）
	ContentType string          `json:"contentType"`                // MIME类型
	MediaType   MediaType       `gorm:"index" json:"mediaType"`     // 媒体类型
	Status      MediaStatus     `gorm:"default:'uploading'" json:"status"`
	Thumbnail   *string         `json:"thumbnail,omitempty"` // 缩略图URL
	Preview     *string         `json:"preview,omitempty"`   // 预览图URL
	Meta        *MediaMeta      `gorm:"type:jsonb" json:"meta,omitempty"`
	Category    string          `gorm:"index" json:"category"` // 分类
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   *gorm.DeletedAt `gorm:"index" json:"-"`
}

// PublicVideoList 公开视频列表项模型
type PublicVideoList struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Thumbnail   *string   `json:"thumbnail,omitempty"`
	Duration    string    `json:"duration,omitempty"` // 时长，格式如"25:30"
	Category    string    `json:"category"`           // 分类
	UploadTime  time.Time `json:"uploadTime"`
}

// PublicAudioList 公开音频列表项模型
type PublicAudioList struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Duration   string    `json:"duration,omitempty"` // 时长，格式如"25:30"
	UploadTime time.Time `json:"uploadTime"`         // 上传时间，格式如"2024-01-15"
}

// AudioDetail 音频详情模型
type AudioDetail struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	MediaUrl string `json:"mediaUrl"` // 音频URL
}

// VideoDetail 视频详情模型
type VideoDetail struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	MediaUrl    string     `json:"mediaUrl"`           // 改名为 mediaUrl 匹配前端
	Preview     *string    `json:"preview,omitempty"`  // 预览图URL
	Duration    string     `json:"duration,omitempty"` // 时长，格式如"25:30"
	UploadTime  time.Time  `json:"uploadTime"`         // 上传时间，格式如"2024-01-15"
	Category    string     `json:"category"`           // 分类
	Meta        *MediaMeta `json:"meta,omitempty"`
}

// BeforeCreate 在创建媒体记录前设置默认值
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	// 根据ContentType自动设置MediaType
	if m.MediaType == "" {
		switch {
		case m.ContentType == "video/mp4" || m.ContentType == "video/webm" || m.ContentType == "video/quicktime":
			m.MediaType = MediaTypeVideo
		case m.ContentType == "audio/mpeg" || m.ContentType == "audio/wav" || m.ContentType == "audio/ogg":
			m.MediaType = MediaTypeAudio
		case m.ContentType == "image/jpeg" || m.ContentType == "image/png" || m.ContentType == "image/gif" || m.ContentType == "image/webp":
			m.MediaType = MediaTypeImage
		default:
			m.MediaType = MediaTypeOther
		}
	}

	return nil
}

// FormatDuration 将秒数转换为 "MM:SS" 格式
func formatDuration(seconds int64) string {
	if seconds == 0 {
		return ""
	}
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

// ToPublicVideoList 将Media转换为PublicVideoList
func (m *Media) ToPublicVideoList() PublicVideoList {
	var duration string
	if m.Meta != nil && m.Meta.Duration != nil {
		duration = formatDuration(*m.Meta.Duration)
	}

	return PublicVideoList{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Thumbnail:   m.Thumbnail,
		Duration:    duration,
		Category:    m.Category,
		UploadTime:  m.CreatedAt,
	}
}

// ToPublicAudioList 将Media转换为PublicAudioList
func (m *Media) ToPublicAudioList() PublicAudioList {
	var duration string
	if m.Meta != nil && m.Meta.Duration != nil {
		duration = formatDuration(*m.Meta.Duration)
	}

	return PublicAudioList{
		ID:         m.ID,
		Title:      m.Title,
		Duration:   duration,
		UploadTime: m.CreatedAt,
	}
}

// ToAudioDetail 将Media转换为AudioDetail
func (m *Media) ToAudioDetail() AudioDetail {
	return AudioDetail{
		ID:       m.ID,
		Title:    m.Title,
		MediaUrl: m.FileURL, // 映射到 mediaUrl
	}
}

// ToVideoDetail 将Media转换为VideoDetail
func (m *Media) ToVideoDetail() VideoDetail {
	var duration string
	if m.Meta != nil && m.Meta.Duration != nil {
		duration = formatDuration(*m.Meta.Duration)
	}

	return VideoDetail{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		MediaUrl:    m.FileURL, // 映射到 mediaUrl
		Preview:     m.Preview,
		Duration:    duration,
		UploadTime:  m.CreatedAt,
		Category:    m.Category,
		Meta:        m.Meta,
	}
}
