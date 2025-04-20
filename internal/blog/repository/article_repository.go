package repository

import (
	"betalyr-learning-server/internal/blog/models"
	"betalyr-learning-server/internal/database"
)

// ArticleRepository 文章仓库接口
type ArticleRepository interface {
	Create(article *models.Article) error
	Update(article *models.Article) error
	Delete(id uint) error
	FindByID(id uint) (*models.Article, error)
	List(page, pageSize int) ([]models.Article, int64, error)
}

// articleRepository 文章仓库实现
type articleRepository struct{}

// NewArticleRepository 创建文章仓库实例
func NewArticleRepository() ArticleRepository {
	return &articleRepository{}
}

// Create 创建文章
func (r *articleRepository) Create(article *models.Article) error {
	return database.DB.Create(article).Error
}

// Update 更新文章
func (r *articleRepository) Update(article *models.Article) error {
	return database.DB.Save(article).Error
}

// Delete 删除文章
func (r *articleRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Article{}, id).Error
}

// FindByID 根据ID查找文章
func (r *articleRepository) FindByID(id uint) (*models.Article, error) {
	var article models.Article
	err := database.DB.First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// List 获取文章列表
func (r *articleRepository) List(page, pageSize int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	offset := (page - 1) * pageSize

	err := database.DB.Model(&models.Article{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = database.DB.Offset(offset).Limit(pageSize).Find(&articles).Error
	if err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}
