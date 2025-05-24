package router

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/handler"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"
	"betalyr-learning-server/internal/service"

	"github.com/gin-gonic/gin"
)

// registerDocumentRoutes 注册文档相关路由
func registerDocumentRoutes(r *gin.Engine, cfg *config.Config) {
	// 初始化文档相关依赖
	documentRepo := repository.NewDocumentRepository()
	documentService := service.NewDocumentService(documentRepo)
	cloudinaryService := service.NewCloudinaryService(cfg)

	// 初始化处理器
	documentHandler := handler.NewDocumentHandler(documentService, cloudinaryService)

	// 初始化用户处理器
	userHandler := handler.NewUserHandler(documentRepo)

	// 需要验证的API路由
	api := r.Group("")
	// 应用身份验证中间件
	api.Use(middleware.AuthChecker())

	// 用户相关路由
	api.PUT("/update-stories-user", userHandler.UpdateStoriesUser)

	// 文档相关路由
	documents := api.Group("/documents")
	{
		// 创建空文档
		documents.POST("/createEmptyDoc", documentHandler.CreateEmptyDoc)

		// 查找文档是否存在
		documents.GET("/findDoc/:id", documentHandler.FindDoc)

		// 删除文档
		documents.DELETE("/deleteDoc/:id", documentHandler.DeleteDoc)

		// 获取用户文档列表
		documents.GET("/user", documentHandler.GetUserDocs)

		// 发布文档
		documents.PATCH("/:id/publish", documentHandler.PublishDoc)

		// 取消发布文档
		documents.PATCH("/:id/unpublish", documentHandler.UnpublishDoc)

		// 更新文档
		documents.PUT("/:id", documentHandler.UpdateDoc)

		// 获取文档详情
		documents.GET("/:id", documentHandler.GetDoc)
	}

}
