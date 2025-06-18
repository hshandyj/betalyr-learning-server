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
		// 上传视频文件
		media.POST("/upload/video", mediaHandler.UploadVideo)
		// 上传音频文件
		media.POST("/upload/audio", mediaHandler.UploadAudio)

		// 获取视频详情
		media.GET("/video/:id", mediaHandler.GetVideoDetail)
		// 获取音频详情
		media.GET("/audio/:id", mediaHandler.GetAudioDetail)

		// 删除媒体文件
		media.DELETE("/:id", mediaHandler.DeleteMedia)
	}
}
