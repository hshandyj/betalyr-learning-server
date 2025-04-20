package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	DB     DBConfig     `yaml:"db"`
	Server ServerConfig `yaml:"server"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// expandEnvVars 展开环境变量
func expandEnvVars(value string) string {
	return os.ExpandEnv(strings.ReplaceAll(value, "${", "$"))
}

// processConfig 处理配置中的环境变量
func processConfig(cfg *Config) {
	cfg.DB.Host = expandEnvVars(cfg.DB.Host)
	cfg.DB.Port = expandEnvVars(cfg.DB.Port)
	cfg.DB.User = expandEnvVars(cfg.DB.User)
	cfg.DB.Password = expandEnvVars(cfg.DB.Password)
	cfg.DB.DBName = expandEnvVars(cfg.DB.DBName)
	cfg.Server.Port = expandEnvVars(cfg.Server.Port)
}

// NewConfig 创建配置
func NewConfig() *Config {
	// 默认配置
	config := &Config{
		DB: DBConfig{
			Host:     "postgres",
			Port:     "5432",
			User:     "betalyr_lerning_dev",
			Password: "dev123",
			DBName:   "betalyr_lerningdb_dev",
		},
		Server: ServerConfig{
			Port: "8000",
		},
	}

	// 尝试从配置文件加载
	configPath := filepath.Join("configs", "config.yaml")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			panic("Failed to parse config file: " + err.Error())
		}
		// 处理环境变量
		processConfig(config)
	}

	return config
}
