package main

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/handler"
	"betalyr-learning-server/internal/models"
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/pkg/middleware"
	"betalyr-learning-server/internal/repository"
	"betalyr-learning-server/internal/service"
	"betalyr-learning-server/internal/storage"
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// 应用程序启动时间
var startTime = time.Now()

func main() {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		// 如果.env文件不存在，只记录警告，不终止程序
		// 因为配置可以通过环境变量设置
	}

	// 设置 Gin 为发布模式，禁用控制台颜色
	gin.SetMode(gin.ReleaseMode)

	// 初始化日志
	logger.InitLogger("development")
	defer logger.Log.Sync()

	// 加载配置
	cfg := config.NewConfig()
	logger.Info("Configuration loaded")

	// 初始化数据库连接
	if err := database.Initialize(cfg); err != nil {
		logger.Fatal("Database initialization failed", zap.Error(err))
	}
	logger.Info("Database connection successful")

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(&models.Document{}); err != nil {
		logger.Fatal("Database migration failed", zap.Error(err))
	}
	logger.Info("Database migration completed")

	// 初始化R2对象存储 (如果配置了R2)
	if cfg.R2.Endpoint != "" {
		if err := storage.InitializeR2(cfg); err != nil {
			logger.Warn("R2 initialization failed, media storage will not be available", zap.Error(err))
		} else {
			logger.Info("R2 object storage initialization successful")
		}
	} else {
		logger.Warn("R2 object storage not configured, media storage will not be available")
	}

	// 初始化文档相关依赖
	documentRepo := repository.NewDocumentRepository()
	documentService := service.NewDocumentService(documentRepo)

	// // 初始化媒体存储库 (R2初始化成功时才创建)
	// var mediaRepo repository.MediaRepository
	// if storage.R2Client != nil {
	// 	mediaRepo = repository.NewMediaRepository()
	// 	logger.Info("媒体存储库初始化成功")
	// }

	// 初始化Cloudinary服务
	cloudinaryService := service.NewCloudinaryService(cfg)

	// 输出Cloudinary配置信息
	logger.Info("Cloudinary configuration",
		zap.String("cloudName", cfg.Cloudinary.CloudName),
		zap.String("apiKey", cfg.Cloudinary.APIKey),
		zap.String("apiSecretLength", fmt.Sprintf("%d chars", len(cfg.Cloudinary.APISecret))))

	// 初始化文档处理器，注入Cloudinary服务
	documentHandler := handler.NewDocumentHandler(documentService, cloudinaryService)

	// 初始化用户处理器
	userHandler := handler.NewUserHandler(documentRepo)

	// 设置路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger()) // 使用统一的日志中间件

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3030", "https://375566.xyz"},
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

	// 主页/健康检查路由不需要身份验证
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"version": "1.0.0",
			"uptime":  time.Since(startTime).String(),
		})
	})

	// 专用健康检查端点不需要身份验证
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 公开文章列表不需要身份验证
	public := r.Group("/public")
	{
		public.GET("/documents", documentHandler.GetPublishedDocs)
		// Cloudinary签名接口
		public.POST("/sign-cloudinary", documentHandler.CloudinarySignRequest)
	}

	// 需要验证的API路由
	api := r.Group("")
	// 应用身份验证中间件
	api.Use(middleware.AuthChecker())

	// 用户相关路由
	api.PUT("/update-stories-user", userHandler.UpdateStoriesUser)

	// 文档相关路由
	documents := api.Group("/documents")
	{
		// 创建空文档（放在最前面）
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

		// 获取文档详情（通用路由放在最后）
		documents.GET("/:id", documentHandler.GetDoc)
	}

	// 获取PORT环境变量(Render优先使用这个)
	port := os.Getenv("PORT")
	if port == "" {
		// 若没有PORT环境变量，则使用配置中的端口
		if cfg.Server.Port == "" {
			// 若配置也没有指定，则使用默认端口8000
			port = "8000"
		} else {
			port = cfg.Server.Port
		}
	}

	// 启动服务器
	logger.Info("Server started", zap.String("port", port))
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("Server startup failed", zap.Error(err))
	}
}
