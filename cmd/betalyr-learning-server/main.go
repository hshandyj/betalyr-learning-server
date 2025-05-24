package main

import (
	"betalyr-learning-server/internal/app"
	"betalyr-learning-server/internal/pkg/logger"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	// 创建应用程序实例
	application := app.New()

	// 初始化应用程序
	if err := application.Initialize(); err != nil {
		fmt.Printf("应用程序初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 启动应用程序并捕获错误
	if err := application.Run(); err != nil {
		logger.Fatal("服务器启动失败", zap.Error(err))
	}

	// 优雅关闭
	defer application.Close()
}
