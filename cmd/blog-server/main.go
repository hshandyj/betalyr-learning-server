package main

import (
	"blog-server/internal/blog/handler"
	"blog-server/internal/blog/models"
	"blog-server/internal/blog/repository"
	"blog-server/internal/blog/service"
	"blog-server/internal/config"
	"blog-server/internal/database"
	"blog-server/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger.InitLogger("development")
	defer logger.Log.Sync()

	// 加载配置
	cfg := config.NewConfig()
	logger.Info("配置加载完成")

	// 初始化数据库连接
	if err := database.Initialize(cfg); err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(&models.Article{}); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}
	logger.Info("数据库迁移完成")

	// 初始化依赖
	articleRepo := repository.NewArticleRepository()
	articleService := service.NewArticleService(articleRepo)
	articleHandler := handler.NewArticleHandler(articleService)

	// 设置路由
	r := gin.Default()

	// 文章相关路由
	articles := r.Group("/api/articles")
	{
		articles.POST("", articleHandler.Create)
		articles.PUT("/:id", articleHandler.Update)
		articles.DELETE("/:id", articleHandler.Delete)
		articles.GET("/:id", articleHandler.Get)
		articles.GET("", articleHandler.List)
	}

	// 启动服务器
	logger.Info("服务器启动", zap.String("port", cfg.Server.Port))
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
