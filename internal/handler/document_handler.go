package handler

import (
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/service"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

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
	UnpublishDoc(c *gin.Context)
	DeleteDoc(c *gin.Context)
	CloudinarySignRequest(c *gin.Context)
	GetPublishedDocs(c *gin.Context)
}

// documentHandler 实现文档处理器接口
type documentHandler struct {
	service      service.DocumentService
	cloudService service.CloudinaryService
}

// NewDocumentHandler 创建新的文档处理器实例
func NewDocumentHandler(service service.DocumentService, cloudService service.CloudinaryService) DocumentHandler {
	return &documentHandler{
		service:      service,
		cloudService: cloudService,
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
	// 使用辅助函数从上下文中获取用户ID
	userIdStr, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("未找到用户ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	logger.Info("创建空文档", zap.String("userID", userIdStr))

	doc, err := h.service.CreateEmptyDoc(userIdStr)
	if err != nil {
		logger.Error("创建空文档失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// GetUserDocs 获取用户的文档列表
func (h *documentHandler) GetUserDocs(c *gin.Context) {
	// 使用辅助函数从上下文中获取用户ID
	userIdStr, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("未找到用户ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	docs, err := h.service.GetUserDocs(userIdStr)
	if err != nil {
		logger.Error("获取用户文档列表失败", zap.Error(err), zap.String("userID", userIdStr))
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

	// 读取请求体内容用于日志记录
	bodyBytes, _ := c.GetRawData()
	bodyString := string(bodyBytes)
	logger.Info("更新文档请求体", zap.String("documentID", documentID), zap.String("body", bodyString))

	// 重新设置请求体以供后续绑定使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&updates); err != nil {
		logger.Error("解析更新内容失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 记录解析后的更新内容
	updatesJson, _ := json.Marshal(updates)
	logger.Info("文档更新内容", zap.String("documentID", documentID), zap.String("updates", string(updatesJson)))

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

	// 返回成功状态
	c.JSON(http.StatusOK, true)
}

// UnpublishDoc 将文档设为非公开
func (h *documentHandler) UnpublishDoc(c *gin.Context) {
	documentID := c.Param("id")

	doc, err := h.service.UnpublishDoc(documentID)
	if err != nil {
		logger.Error("取消发布文档失败", zap.Error(err), zap.String("documentID", documentID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	if doc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文档不存在"})
		return
	}

	// 返回成功状态
	c.JSON(http.StatusOK, true)
}

// DeleteDoc 删除文档
func (h *documentHandler) DeleteDoc(c *gin.Context) {
	documentID := c.Param("id")

	// 使用辅助函数从上下文中获取用户ID
	userIdStr, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("未找到用户ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	logger.Info("删除文档", zap.String("documentID", documentID), zap.String("userID", userIdStr))

	success, err := h.service.DeleteDoc(documentID, userIdStr)
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

// CloudinarySignRequest 处理Cloudinary签名请求
func (h *documentHandler) CloudinarySignRequest(c *gin.Context) {
	// 获取请求体中的参数
	var requestData struct {
		ParamsToSign map[string]interface{} `json:"paramsToSign"`
	}

	// 读取请求体内容用于日志记录
	bodyBytes, _ := c.GetRawData()
	bodyString := string(bodyBytes)
	logger.Info("Cloudinary签名请求体", zap.String("body", bodyString))

	// 重新设置请求体以供后续绑定使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Error("解析Cloudinary签名参数失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 记录要签名的参数
	paramsJson, _ := json.Marshal(requestData.ParamsToSign)
	logger.Info("Cloudinary参数", zap.String("params", string(paramsJson)))

	// // 验证用户权限
	// virtualUserID := c.GetHeader("X-Virtual-User-ID")
	// if virtualUserID == "" {
	// 	logger.Error("未提供用户ID，无法生成Cloudinary签名")
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
	// 	return
	// }

	// 签名参数
	signature, err := h.cloudService.SignRequest(requestData.ParamsToSign)
	if err != nil {
		logger.Error("生成Cloudinary签名失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	logger.Info("生成Cloudinary签名成功", zap.String("signature", signature))

	// 返回签名结果
	c.JSON(http.StatusOK, gin.H{"signature": signature})
}

// GetPublishedDocs 获取所有公开发布的文章
func (h *documentHandler) GetPublishedDocs(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	// 转换为整数
	page := 1
	limit := 20
	if pageInt, err := strconv.Atoi(pageStr); err == nil && pageInt > 0 {
		page = pageInt
	}
	if limitInt, err := strconv.Atoi(limitStr); err == nil && limitInt > 0 {
		limit = limitInt
	}

	// 限制最大数量
	if limit > 100 {
		limit = 100
	}

	logger.Info("获取公开文章列表",
		zap.Int("page", page),
		zap.Int("limit", limit))

	// 调用服务层获取数据
	docs, total, err := h.service.GetPublishedDocs(page, limit)
	if err != nil {
		logger.Error("获取公开文章失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data": docs,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
