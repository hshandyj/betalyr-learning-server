package app

import (
	"betalyr-learning-server/internal/config"
	"betalyr-learning-server/internal/database"
	"betalyr-learning-server/internal/models"
	"betalyr-learning-server/internal/pkg/logger"
	"betalyr-learning-server/internal/router"
	"betalyr-learning-server/internal/storage"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// 应用程序启动时间
var startTime = time.Now()

// App 应用程序结构体
type App struct {
	Config *config.Config
	Router *gin.Engine
}

// New 创建新的应用程序实例
func New() *App {
	return &App{}
}

// Initialize 初始化应用程序
func (a *App) Initialize() error {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		// 如果.env文件不存在，只记录警告，不终止程序
		// 因为配置可以通过环境变量设置
	}

	// 设置 Gin 为发布模式，禁用控制台颜色
	gin.SetMode(gin.ReleaseMode)

	// 初始化日志
	logger.InitLogger("development")

	// 加载配置
	a.Config = config.NewConfig()
	logger.Info("server config loaded")

	// 初始化数据库连接
	if err := database.Initialize(a.Config); err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}
	logger.Info("database connected")

	// 自动迁移数据库表
	if err := database.DB.AutoMigrate(&models.Document{}); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	logger.Info("database migrated")

	// 初始化R2对象存储 (如果配置了R2)
	if a.Config.R2.Endpoint != "" {
		if err := storage.InitializeR2(a.Config); err != nil {
			logger.Warn("R2 initialization failed, media storage will not be available", zap.Error(err))
		} else {
			logger.Info("R2 object storage initialized")
		}
	} else {
		logger.Warn("R2 object storage not configured, media storage will not be available")
	}

	// 初始化路由器
	a.Router = router.SetupRouter(a.Config)
	logger.Info("router initialized")

	return nil
}

// Run 运行应用程序
func (a *App) Run() error {
	// 获取PORT环境变量(Render优先使用这个)
	port := os.Getenv("PORT")
	if port == "" {
		// 若没有PORT环境变量，则使用配置中的端口
		if a.Config.Server.Port == "" {
			// 若配置也没有指定，则使用默认端口8000
			port = "8000"
		} else {
			port = a.Config.Server.Port
		}
	}

	// 启动服务器
	logger.Info("server started", zap.String("port", port))
	return a.Router.Run(":" + port)
}

// Close 关闭应用程序资源
func (a *App) Close() {
	// 这里可以添加需要清理的资源，比如关闭数据库连接等
	if err := logger.Log.Sync(); err != nil {
		fmt.Printf("failed to sync logger: %v\n", err)
	}
}
