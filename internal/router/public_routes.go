package router

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/handler"
	"betalyr-learning-server/internal/repository"
	"betalyr-learning-server/internal/service"

	"github.com/gin-gonic/gin"
)

// registerPublicRoutes 注册不需要身份验证的公共路由
func registerPublicRoutes(r *gin.Engine, cfg *config.Config) {
	// 初始化文档相关依赖
	documentRepo := repository.NewDocumentRepository()
	documentService := service.NewDocumentService(documentRepo)
	cloudinaryService := service.NewCloudinaryService(cfg)

	// 初始化处理器
	documentHandler := handler.NewDocumentHandler(documentService, cloudinaryService)

	// 公开文章列表不需要身份验证
	public := r.Group("/public")
	{
		public.GET("/documents", documentHandler.GetPublishedDocs)
		// Cloudinary签名接口
		public.POST("/sign-cloudinary", documentHandler.CloudinarySignRequest)
	}
}
