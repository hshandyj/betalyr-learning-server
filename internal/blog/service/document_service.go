package service

import (
	"betalyr-learning-server/internal/blog/models"
	"betalyr-learning-server/internal/blog/repository"
	"time"

	"github.com/google/uuid"
)

// DocumentService 定义文档服务接口
type DocumentService interface {
	FindDoc(id string) (bool, error)
	GetDoc(id string) (*models.Document, error)
	CreateEmptyDoc(ownerID string) (*models.Document, error)
	GetUserDocs(userID string) ([]models.DocumentList, error)
	UpdateDoc(id string, updates map[string]interface{}) (*models.Document, error)
	PublishDoc(id string) (*models.Document, error)
	DeleteDoc(id string, ownerID string) (bool, error)
}

// documentService 文档服务实现
type documentService struct {
	repo repository.DocumentRepository
}

// NewDocumentService 创建新的文档服务实例
func NewDocumentService(repo repository.DocumentRepository) DocumentService {
	return &documentService{
		repo: repo,
	}
}

// FindDoc 检查文档是否存在
func (s *documentService) FindDoc(id string) (bool, error) {
	return s.repo.CheckDocExists(id)
}

// GetDoc 获取文档详情
func (s *documentService) GetDoc(id string) (*models.Document, error) {
	return s.repo.FindByID(id)
}

// CreateEmptyDoc 创建空文档
func (s *documentService) CreateEmptyDoc(ownerID string) (*models.Document, error) {
	// 生成唯一ID
	id := uuid.New().String()

	// 创建默认文档
	now := time.Now()
	isPublic := false

	doc := &models.Document{
		ID:         id,
		OwnerID:    ownerID, // 设置从请求头中获取的虚拟用户ID
		Title:      "Untitled",
		CreatedAt:  now,
		UpdatedAt:  now,
		IconImage:  nil, // 确保明确设置为nil
		CoverImage: nil, // 确保明确设置为nil
		EditorJSON: nil, // 确保明确设置为nil
		IsPublic:   &isPublic,
	}

	// 保存文档
	err := s.repo.Create(doc)
	if err != nil {
		return nil, err
	}

	// 返回完整的文档对象
	return doc, nil
}

// GetUserDocs 获取用户的文档列表
func (s *documentService) GetUserDocs(userID string) ([]models.DocumentList, error) {
	docs, err := s.repo.GetDocumentsByOwner(userID)
	if err != nil {
		return nil, err
	}

	// 转换为列表格式
	result := make([]models.DocumentList, len(docs))
	for i, doc := range docs {
		result[i] = doc.ToDocumentList()
	}

	return result, nil
}

// UpdateDoc 更新文档
func (s *documentService) UpdateDoc(id string, updates map[string]interface{}) (*models.Document, error) {
	// 获取现有文档
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // 文档不存在
	}

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		doc.Title = title
	}

	if ownerId, ok := updates["ownerId"].(string); ok && doc.OwnerID == "" {
		doc.OwnerID = ownerId
	}

	if iconImage, ok := updates["iconImage"].(map[string]interface{}); ok {
		if url, urlOk := iconImage["url"].(string); urlOk {
			timestamp, _ := iconImage["timeStamp"].(float64)
			doc.IconImage = &models.Image{
				URL:       url,
				TimeStamp: int64(timestamp),
			}
		}
	}

	if coverImage, ok := updates["coverImage"].(map[string]interface{}); ok {
		if url, urlOk := coverImage["url"].(string); urlOk {
			timestamp, _ := coverImage["timeStamp"].(float64)
			doc.CoverImage = &models.Image{
				URL:       url,
				TimeStamp: int64(timestamp),
			}
		}
	}

	if editorJson, ok := updates["editorJson"].(map[string]interface{}); ok {
		jsonContent := models.JSONContent(editorJson)
		doc.EditorJSON = &jsonContent
	}

	// 更新时间
	doc.UpdatedAt = time.Now()

	// 保存更新
	err = s.repo.Update(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// PublishDoc 将文档设为公开
func (s *documentService) PublishDoc(id string) (*models.Document, error) {
	// 获取现有文档
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // 文档不存在
	}

	// 设置为公开
	isPublic := true
	doc.IsPublic = &isPublic
	doc.UpdatedAt = time.Now()

	// 保存更新
	err = s.repo.Update(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// DeleteDoc 删除文档
func (s *documentService) DeleteDoc(id string, ownerID string) (bool, error) {
	// 获取现有文档
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return false, err
	}

	// 如果文档不存在或用户ID不匹配，拒绝删除
	if doc == nil || doc.OwnerID != ownerID {
		return false, nil
	}

	// 删除文档
	err = s.repo.Delete(id)
	if err != nil {
		return false, err
	}

	return true, nil
}
