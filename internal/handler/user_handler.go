package handler

import (
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserHandler 定义用户处理器接口
type UserHandler interface {
	UpdateStoriesUser(c *gin.Context)
}

// userHandler 实现用户处理器接口
type userHandler struct {
	docRepo repository.DocumentRepository
}

// NewUserHandler 创建新的用户处理器实例
func NewUserHandler(docRepo repository.DocumentRepository) UserHandler {
	return &userHandler{
		docRepo: docRepo,
	}
}

// UpdateStoriesUser 将虚拟用户的所有文章更新为新的用户ID
func (h *userHandler) UpdateStoriesUser(c *gin.Context) {
	// 从上下文中获取当前用户ID（通过JWT令牌或登录后获取的用户ID）
	newUserId, exists := middleware.GetUserID(c)
	if !exists {
		logger.Error("未找到用户ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	// 从请求头中获取要迁移的虚拟用户ID
	virtualUserId := c.GetHeader("X-Virtual-User-ID")
	if virtualUserId == "" {
		logger.Error("未提供虚拟用户ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "未提供虚拟用户ID"})
		return
	}

	// 确保两个ID不同
	if virtualUserId == newUserId {
		logger.Warn("尝试将文章迁移到相同的用户ID",
			zap.String("userId", newUserId))
		c.JSON(http.StatusBadRequest, gin.H{"error": "虚拟用户ID和目标用户ID相同，无需迁移"})
		return
	}

	logger.Info("开始迁移用户文章",
		zap.String("virtualUserId", virtualUserId),
		zap.String("newUserId", newUserId))

	// 调用仓库层方法更新所有文章
	count, err := h.docRepo.UpdateOwnerID(virtualUserId, newUserId)
	if err != nil {
		logger.Error("迁移用户文章失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		return
	}

	// 返回迁移成功的信息
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户文章迁移成功",
		"count":   count,
	})
}
