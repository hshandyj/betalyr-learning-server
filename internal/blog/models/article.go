package models

import (
	"time"
)

// Article 博客文章模型
type Article struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserId    uint      `gorm:"not null" json:"userId"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Status    int       `gorm:"default:0" json:"status"` // 0: 草稿, 1: 已发布
	Tags      string    `gorm:"size:200" json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Author    string    `gorm:"size:100;not null" json:"author"`
	Excerpt   string    `gorm:"type:text" json:"excerpt"`
}

// ArticleListItem 文章列表项（不包含文章内容）
type ArticleListItem struct {
	ID        uint      `json:"id"`
	UserId    uint      `json:"userId"`
	Title     string    `json:"title"`
	Status    int       `json:"status"`
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Author    string    `json:"author"`
	Excerpt   string    `json:"excerpt"`
}

// ToListItem 将完整文章转换为列表项
func (a *Article) ToListItem() ArticleListItem {
	return ArticleListItem{
		ID:        a.ID,
		UserId:    a.UserId,
		Title:     a.Title,
		Status:    a.Status,
		Tags:      a.Tags,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Author:    a.Author,
		Excerpt:   a.Excerpt,
	}
}
