package router

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/handler"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"

	"github.com/gin-gonic/gin"
)

// registerMediaRoutes 注册媒体相关路由
func registerMediaRoutes(r *gin.Engine, cfg *config.Config) {
	mediaRepo := repository.NewMediaRepository()
	mediaHandler := handler.NewMediaHandler(mediaRepo)

	api := r.Group("")
	api.Use(middleware.AuthChecker())

	media := api.Group("/media")
	{
		// 上传媒体文件
		media.POST("/upload", mediaHandler.UploadMedia)
		// 获取媒体文件URL
		media.GET("/:key", mediaHandler.GetMediaURL)
		// 删除媒体文件
		media.DELETE("/:key", mediaHandler.DeleteMedia)
	}
}
