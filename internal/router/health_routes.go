package router

import (
	"time"

	"github.com/gin-gonic/gin"
)

// 应用启动时间
var startTime = time.Now()

// registerHealthRoutes 注册健康检查相关路由
func registerHealthRoutes(r *gin.Engine) {
	// 主页/健康检查路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"version": "1.0.0",
			"uptime":  time.Since(startTime).String(),
		})
	})

	// 专用健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
}
