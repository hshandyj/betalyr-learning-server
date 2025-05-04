package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Image 结构用于存储图片URL和时间戳
type Image struct {
	URL       string `json:"url"`
	TimeStamp int64  `json:"timeStamp"`
}

// Value 实现driver.Valuer接口，用于将Image结构转换为数据库值
func (i Image) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan 实现sql.Scanner接口，用于将数据库值转换为Image结构
func (i *Image) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &i)
}

// JSONContent 存储编辑器的JSON内容
type JSONContent map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONContent) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONContent) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j)
}

// Document 文档模型
type Document struct {
	ID         string       `gorm:"primaryKey" json:"id"`
	OwnerID    string       `gorm:"index" json:"ownerId"`
	Title      string       `json:"title"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	IconImage  *Image       `gorm:"type:jsonb" json:"iconImage,omitempty"`
	CoverImage *Image       `gorm:"type:jsonb" json:"coverImage,omitempty"`
	EditorJSON *JSONContent `gorm:"type:jsonb" json:"editorJson,omitempty"`
	IsPublic   *bool        `gorm:"default:false" json:"isPublic,omitempty"`
}

// DocumentList 文档列表项模型
type DocumentList struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	CoverImage *Image `json:"coverImage,omitempty"`
}

// BeforeCreate 在创建文档前设置默认值
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	// 如果没有设置IsPublic，默认为false
	if d.IsPublic == nil {
		isPublic := false
		d.IsPublic = &isPublic
	}
	return nil
}

// ToDocumentList 将Document转换为DocumentList
func (d *Document) ToDocumentList() DocumentList {
	return DocumentList{
		ID:         d.ID,
		Title:      d.Title,
		CoverImage: d.CoverImage,
	}
}

// ToEmptyDocument 将Document转换为不包含内容的空文档
func (d *Document) ToEmptyDocument() map[string]interface{} {
	return map[string]interface{}{
		"id":        d.ID,
		"ownerId":   d.OwnerID,
		"title":     d.Title,
		"createdAt": d.CreatedAt,
		"updatedAt": d.UpdatedAt,
	}
}
