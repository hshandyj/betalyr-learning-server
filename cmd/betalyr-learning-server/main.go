package main

import (
	"betalyr-learning-server/internal/blog/handler"
	"betalyr-learning-server/internal/blog/models"
	"betalyr-learning-server/internal/blog/repository"
	"betalyr-learning-server/internal/blog/service"
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/pkg/logger"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 健康状态信息
type HealthStatus struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
	Uptime      string    `json:"uptime"`
	Host        string    `json:"host"`
	OS          string    `json:"os"`
	DBStatus    string    `json:"db_status"`
}

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
	if err := database.DB.AutoMigrate(&models.Article{}); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}
	logger.Info("数据库迁移完成")

	// 初始化依赖
	articleRepo := repository.NewArticleRepository()
	articleService := service.NewArticleService(articleRepo)
	articleHandler := handler.NewArticleHandler(articleService)

	// 设置路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLogger()) // 使用统一的日志中间件

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3030"}, // 允许的前端域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowWildcard:    true, // 允许通配符
		MaxAge:           12 * time.Hour,
	}))

	// 添加 OPTIONS 请求的全局处理
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(200)
	})

	// 健康检查路由
	r.GET("/", func(c *gin.Context) {
		hostname, _ := os.Hostname()

		// 检查数据库连接状态
		dbStatus := "正常"
		sqlDB, err := database.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "异常"
		}

		// 版本信息（可以从环境变量获取或硬编码）
		version := os.Getenv("APP_VERSION")
		if version == "" {
			version = "1.0.0" // 默认版本号
		}

		// 环境信息
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "production" // 默认环境
		}

		health := HealthStatus{
			Status:      "运行中",
			Version:     version,
			Environment: env,
			Timestamp:   time.Now(),
			Uptime:      time.Since(startTime).String(),
			Host:        hostname,
			OS:          runtime.GOOS + "/" + runtime.GOARCH,
			DBStatus:    dbStatus,
		}

		c.JSON(http.StatusOK, health)
	})

	// 健康检查接口（用于Fly.io健康检查）
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	})

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
