package database

import (
	"fmt"
	"log"

	"betalyr-learning-server/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB 全局数据库连接
var DB *gorm.DB

// Initialize 初始化数据库连接
func Initialize(config *config.Config) error {
	var dsn string

	// 优先使用数据库URL (如果存在)
	if config.DB.URL != "" {
		dsn = config.DB.URL
	} else {
		// 否则使用单独的连接参数构建DSN
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			config.DB.Host,
			config.DB.User,
			config.DB.Password,
			config.DB.DBName,
			config.DB.Port,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	DB = db
	return nil
}
