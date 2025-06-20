package service

import (
	"betalyr-learning-server/internal/models"
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/repository"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DocumentService 定义文档服务接口
type DocumentService interface {
	FindDoc(id string) (bool, error)
	GetDoc(id string) (*models.Document, error)
	CreateEmptyDoc(ownerID string) (*models.Document, error)
	GetUserDocs(userID string) ([]models.DocumentList, error)
	UpdateDoc(id string, ownerID string, updates map[string]interface{}) (*models.Document, error)
	PublishDoc(id string, ownerID string) (*models.Document, error)
	UnpublishDoc(id string, ownerID string) (*models.Document, error)
	DeleteDoc(id string, ownerID string) (bool, error)
	GetPublishedDocs(page, limit int) ([]models.PublicDocumentList, int64, error)
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
func (s *documentService) UpdateDoc(id string, ownerID string, updates map[string]interface{}) (*models.Document, error) {
	// 获取现有文档
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // 文档不存在
	}

	// 验证文档所有者
	if doc.OwnerID != ownerID {
		logger.Warn("User attempted to update a document that is not their own",
			zap.String("documentID", id),
			zap.String("docOwnerID", doc.OwnerID),
			zap.String("requestUserID", ownerID))
		return nil, nil // 不是文档所有者，返回nil表示无权更新
	}

	// 记录更新前的文档状态
	beforeJson, _ := json.Marshal(doc)
	logger.Info("Document state before update", zap.String("documentID", id), zap.String("before", string(beforeJson)))

	// 应用更新
	if title, ok := updates["title"].(string); ok {
		doc.Title = title
	}

	if ownerId, ok := updates["ownerId"].(string); ok && doc.OwnerID == "" {
		doc.OwnerID = ownerId
	}

	// 处理iconImage字段
	if iconImage, exists := updates["iconImage"]; exists {
		// 检查是否为null
		if iconImage == nil {
			logger.Info("Setting iconImage to null", zap.String("documentID", id))
			doc.IconImage = nil // 如果前端传入null，则将字段设置为nil
		} else if iconImageMap, ok := iconImage.(map[string]interface{}); ok {
			// 如果是map类型，则按照原来的逻辑处理
			if url, urlOk := iconImageMap["url"].(string); urlOk {
				timestamp, _ := iconImageMap["timeStamp"].(float64)
				logger.Info("Setting new iconImage",
					zap.String("documentID", id),
					zap.String("url", url),
					zap.Float64("timeStamp", timestamp))
				doc.IconImage = &models.Image{
					URL:       url,
					TimeStamp: int64(timestamp),
				}
			}
		} else {
			// 记录意外的类型
			logger.Warn("Unexpected type for iconImage",
				zap.String("documentID", id),
				zap.String("type", fmt.Sprintf("%T", iconImage)))
		}
	}

	// 处理coverImage字段
	if coverImage, exists := updates["coverImage"]; exists {
		// 检查是否为null
		if coverImage == nil {
			doc.CoverImage = nil // 如果前端传入null，则将字段设置为nil
		} else if coverImageMap, ok := coverImage.(map[string]interface{}); ok {
			// 如果是map类型，则按照原来的逻辑处理
			if url, urlOk := coverImageMap["url"].(string); urlOk {
				timestamp, _ := coverImageMap["timeStamp"].(float64)
				doc.CoverImage = &models.Image{
					URL:       url,
					TimeStamp: int64(timestamp),
				}
			}
		}
	}

	// 处理editorJson字段
	if editorJson, exists := updates["editorJson"]; exists {
		// 检查是否为null
		if editorJson == nil {
			doc.EditorJSON = nil // 如果前端传入null，则将字段设置为nil
		} else if editorJsonMap, ok := editorJson.(map[string]interface{}); ok {
			// 如果是map类型，则按照原来的逻辑处理
			jsonContent := models.JSONContent(editorJsonMap)
			doc.EditorJSON = &jsonContent
		}
	}

	// 更新时间
	doc.UpdatedAt = time.Now()

	// 记录更新后的文档状态
	afterJson, _ := json.Marshal(doc)
	logger.Info("Document state after update", zap.String("documentID", id), zap.String("after", string(afterJson)))

	// 保存更新
	err = s.repo.Update(doc)
	if err != nil {
		logger.Error("Failed to save update", zap.String("documentID", id), zap.Error(err))
		return nil, err
	}

	return doc, nil
}

// PublishDoc 将文档设为公开
func (s *documentService) PublishDoc(id string, ownerID string) (*models.Document, error) {
	// 获取现有文档
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // 文档不存在
	}

	// 验证文档所有者
	if doc.OwnerID != ownerID {
		logger.Warn("User attempted to publish a document that is not their own",
			zap.String("documentID", id),
			zap.String("docOwnerID", doc.OwnerID),
			zap.String("requestUserID", ownerID))
		return nil, nil // 不是文档所有者，返回nil表示无权更新
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

// UnpublishDoc 将文档设为非公开
func (s *documentService) UnpublishDoc(id string, ownerID string) (*models.Document, error) {
	doc, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil // 文档不存在
	}

	// 验证文档所有者
	if doc.OwnerID != ownerID {
		logger.Warn("User attempted to unpublish a document that is not their own",
			zap.String("documentID", id),
			zap.String("docOwnerID", doc.OwnerID),
			zap.String("requestUserID", ownerID))
		return nil, nil // 不是文档所有者，返回nil表示无权更新
	}

	// 设置为非公开
	isPublic := false
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

// GetPublishedDocs 获取所有公开的文档
func (s *documentService) GetPublishedDocs(page, limit int) ([]models.PublicDocumentList, int64, error) {
	// 设置默认值
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // 限制最大为100，防止请求过大
	}

	// 获取公开文档
	docs, err := s.repo.GetPublishedDocs(page, limit)
	if err != nil {
		return nil, 0, err
	}

	// 获取总数
	totalCount, err := s.repo.CountPublishedDocs()
	if err != nil {
		return nil, 0, err
	}

	// 转换为公开列表格式
	result := make([]models.PublicDocumentList, len(docs))
	for i, doc := range docs {
		result[i] = doc.ToPublicDocumentList()
	}

	return result, totalCount, nil
}
