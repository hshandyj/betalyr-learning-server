package repository

import (
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/models"
	"errors"

	"gorm.io/gorm"
)

// DocumentRepository 定义文档仓库接口
type DocumentRepository interface {
	FindByID(id string) (*models.Document, error)
	CheckDocExists(id string) (bool, error)
	Create(doc *models.Document) error
	Update(doc *models.Document) error
	GetDocumentsByOwner(ownerID string) ([]models.Document, error)
	Delete(id string) error
}

// documentRepository 实现文档仓库接口
type documentRepository struct {
	db *gorm.DB
}

// NewDocumentRepository 创建新的文档仓库实例
func NewDocumentRepository() DocumentRepository {
	return &documentRepository{
		db: database.DB,
	}
}

// FindByID 根据ID查找文档
func (r *documentRepository) FindByID(id string) (*models.Document, error) {
	var doc models.Document
	result := r.db.Where("id = ?", id).First(&doc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到记录返回nil而不是错误
		}
		return nil, result.Error
	}
	return &doc, nil
}

// CheckDocExists 检查文档是否存在
func (r *documentRepository) CheckDocExists(id string) (bool, error) {
	var count int64
	result := r.db.Model(&models.Document{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Create 创建新文档
func (r *documentRepository) Create(doc *models.Document) error {
	return r.db.Create(doc).Error
}

// Update 更新文档
func (r *documentRepository) Update(doc *models.Document) error {
	return r.db.Save(doc).Error
}

// GetDocumentsByOwner 获取用户的所有文档
func (r *documentRepository) GetDocumentsByOwner(ownerID string) ([]models.Document, error) {
	var docs []models.Document
	result := r.db.Where("owner_id = ?", ownerID).Find(&docs)
	if result.Error != nil {
		return nil, result.Error
	}
	return docs, nil
}

// Delete 删除文档
func (r *documentRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Document{}).Error
}
