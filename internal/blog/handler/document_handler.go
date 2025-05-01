package handler

import (
	"betalyr-learning-server/internal/blog/service"
	"betalyr-learning-server/internal/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DocumentHandler 定义文档处理器接口
type DocumentHandler interface {
	FindDoc(c *gin.Context)
	GetDoc(c *gin.Context)
	CreateEmptyDoc(c *gin.Context)
	GetUserDocs(c *gin.Context)
	UpdateDoc(c *gin.Context)
	PublishDoc(c *gin.Context)
	DeleteDoc(c *gin.Context)
}

// documentHandler 实现文档处理器接口
type documentHandler struct {
	service service.DocumentService
}

// NewDocumentHandler 创建新的文档处理器实例
func NewDocumentHandler(service service.DocumentService) DocumentHandler {
	return &documentHandler{
		service: service,
	}
}

// FindDoc 检查文档是否存在
func (h *documentHandler) FindDoc(c *gin.Context) {
	documentID := c.Param("id")

	exists, err := h.service.FindDoc(documentID)
	if err != nil {
		logger.Error("查询文档是否存在失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	c.JSON(http.StatusOK, exists)
}

// GetDoc 获取文档详情
func (h *documentHandler) GetDoc(c *gin.Context) {
	documentID := c.Param("id")

	doc, err := h.service.GetDoc(documentID)
	if err != nil {
		logger.Error("获取文档失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// CreateEmptyDoc 创建空文档
func (h *documentHandler) CreateEmptyDoc(c *gin.Context) {
	// 从请求头中获取虚拟用户ID
	virtualUserID := c.GetHeader("X-Virtual-User-ID")
	doc, err := h.service.CreateEmptyDoc(virtualUserID)
	if err != nil {
		logger.Error("创建空文档失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetUserDocs 获取用户的文档列表
func (h *documentHandler) GetUserDocs(c *gin.Context) {
	// 从请求头中获取虚拟用户ID
	virtualUserID := c.GetHeader("X-Virtual-User-ID")

	docs, err := h.service.GetUserDocs(virtualUserID)
	if err != nil {
		logger.Error("获取用户文档列表失败", zap.Error(err), zap.String("userID", virtualUserID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	c.JSON(http.StatusOK, docs)
}

// UpdateDoc 更新文档内容
func (h *documentHandler) UpdateDoc(c *gin.Context) {
	documentID := c.Param("id")

	// 解析请求体
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		logger.Error("解析更新内容失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	doc, err := h.service.UpdateDoc(documentID, updates)
	if err != nil {
		logger.Error("更新文档失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// PublishDoc 将文档设为公开
func (h *documentHandler) PublishDoc(c *gin.Context) {
	documentID := c.Param("id")

	doc, err := h.service.PublishDoc(documentID)
	if err != nil {
		logger.Error("发布文档失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// DeleteDoc 删除文档
func (h *documentHandler) DeleteDoc(c *gin.Context) {
	documentID := c.Param("id")

	// 从请求头中获取虚拟用户ID
	virtualUserID := c.GetHeader("X-Virtual-User-ID")
	logger.Info("删除文档", zap.String("documentID", documentID), zap.String("virtualUserID", virtualUserID))

	success, err := h.service.DeleteDoc(documentID, virtualUserID)
	if err != nil {
		logger.Error("删除文档失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	if !success {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限删除此文档或文档不存在"})
		return
	}

	c.JSON(http.StatusOK, true)
}
