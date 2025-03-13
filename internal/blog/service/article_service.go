package service

import (
	"blog-server/internal/blog/models"
	"blog-server/internal/blog/repository"
)

// ArticleService 文章服务接口
type ArticleService interface {
	CreateArticle(article *models.Article) error
	UpdateArticle(article *models.Article) error
	DeleteArticle(id uint) error
	GetArticle(id uint) (*models.Article, error)
	ListArticles(page, pageSize int) ([]models.ArticleListItem, int64, error)
}

// articleService 文章服务实现
type articleService struct {
	repo repository.ArticleRepository
}

// NewArticleService 创建文章服务实例
func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

// CreateArticle 创建文章
func (s *articleService) CreateArticle(article *models.Article) error {
	return s.repo.Create(article)
}

// UpdateArticle 更新文章
func (s *articleService) UpdateArticle(article *models.Article) error {
	return s.repo.Update(article)
}

// DeleteArticle 删除文章
func (s *articleService) DeleteArticle(id uint) error {
	return s.repo.Delete(id)
}

// GetArticle 获取文章
func (s *articleService) GetArticle(id uint) (*models.Article, error) {
	return s.repo.FindByID(id)
}

// ListArticles 获取文章列表
func (s *articleService) ListArticles(page, pageSize int) ([]models.ArticleListItem, int64, error) {
	articles, total, err := s.repo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为列表项
	listItems := make([]models.ArticleListItem, len(articles))
	for i, article := range articles {
		listItems[i] = article.ToListItem()
	}

	return listItems, total, nil
}
