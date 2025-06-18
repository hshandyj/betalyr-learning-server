package router

import (
	"betalyr-learning-server/internal/config"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 初始化并配置Gin路由器
func SetupRouter(cfg *config.Config) *gin.Engine {
	// 初始化gin
	r := gin.New()
	r.Use(gin.Recovery())

	// 配置CORS
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

	// 注册各个模块的路由
	registerHealthRoutes(r)
	registerPublicRoutes(r, cfg)
	registerDocumentRoutes(r, cfg)
	registerMediaRoutes(r, cfg)
	return r
}
