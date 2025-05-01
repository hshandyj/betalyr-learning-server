package main

import (
	"betalyr-learning-server/internal/blog/handler"
	"betalyr-learning-server/internal/blog/models"
	"betalyr-learning-server/internal/blog/repository"
	"betalyr-learning-server/internal/blog/service"
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/pkg/logger"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 应用程序启动时间
var startTime = time.Now()

func main() {
	// 设置 Gin 为发布模式，禁用控制台颜色
	gin.SetMode(gin.ReleaseMode)

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
	if err := database.DB.AutoMigrate(&models.Document{}); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}
	logger.Info("数据库迁移完成")

	// 初始化文档相关依赖
	documentRepo := repository.NewDocumentRepository()
	documentService := service.NewDocumentService(documentRepo)
	documentHandler := handler.NewDocumentHandler(documentService)

	// 设置路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger()) // 使用统一的日志中间件

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3030"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Virtual-User-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		AllowWildcard:    true,
		MaxAge:           12 * time.Hour,
	}))

	// 添加预检请求处理
	r.OPTIONS("/*path", func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Virtual-User-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Status(204)
	})

	// 主页/健康检查
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "success",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 文档相关路由
	documents := r.Group("/documents")
	{
		// 创建空文档（放在最前面）
		documents.POST("/createEmptyDoc", documentHandler.CreateEmptyDoc)

		// 查找文档是否存在
		documents.GET("/findDoc/:id", documentHandler.FindDoc)

		// 获取用户文档列表
		documents.GET("/user/:userId", documentHandler.GetUserDocs)

		// 发布文档
		documents.PATCH("/:id/publish", documentHandler.PublishDoc)

		// 更新文档
		documents.PATCH("/:id", documentHandler.UpdateDoc)

		// 获取文档详情（通用路由放在最后）
		documents.GET("/:id", documentHandler.GetDoc)
	}

	// 启动服务器
	logger.Info("服务器启动", zap.String("port", cfg.Server.Port))
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}
}
