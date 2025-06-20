package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	DB         DBConfig         `yaml:"db"`
	Server     ServerConfig     `yaml:"server"`
	Cloudinary CloudinaryConfig `yaml:"cloudinary"`
	R2         R2Config         `yaml:"r2"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	URL      string `yaml:"url"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"`
}

// CloudinaryConfig Cloudinary配置
type CloudinaryConfig struct {
	CloudName string `yaml:"cloud_name"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
}

// R2Config Cloudflare R2配置
type R2Config struct {
	Endpoint        string `yaml:"endpoint"`
	AccountID       string `yaml:"account_id"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
	PublicURL       string `yaml:"public_url"`
}

// expandEnvVars 展开环境变量
func expandEnvVars(value string) string {
	// 找到格式为 ${VAR:-default} 的模式
	if strings.Contains(value, "${") && strings.Contains(value, ":-") && strings.Contains(value, "}") {
		// 提取变量名和默认值
		start := strings.Index(value, "${") + 2
		end := strings.Index(value, "}")
		if start > 1 && end > start {
			varContent := value[start:end]
			parts := strings.Split(varContent, ":-")
			if len(parts) == 2 {
				varName := parts[0]
				defaultVal := parts[1]

				// 从环境变量获取值，如果不存在则使用默认值
				envVal := os.Getenv(varName)
				if envVal != "" {
					return envVal
				}
				return defaultVal
			}
		}
	}

	// 如果不是特殊格式，直接使用标准的环境变量替换
	return os.ExpandEnv(value)
}

// processConfig 处理配置中的环境变量
func processConfig(cfg *Config) {
	// 优先检查是否存在DATABASE_URL环境变量
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DB.URL = dbURL
	}

	cfg.DB.Host = expandEnvVars(cfg.DB.Host)
	cfg.DB.Port = expandEnvVars(cfg.DB.Port)
	cfg.DB.User = expandEnvVars(cfg.DB.User)
	cfg.DB.Password = expandEnvVars(cfg.DB.Password)
	cfg.DB.DBName = expandEnvVars(cfg.DB.DBName)
	cfg.Server.Port = expandEnvVars(cfg.Server.Port)

	// 处理Cloudinary配置
	cfg.Cloudinary.CloudName = expandEnvVars(cfg.Cloudinary.CloudName)
	cfg.Cloudinary.APIKey = expandEnvVars(cfg.Cloudinary.APIKey)
	cfg.Cloudinary.APISecret = expandEnvVars(cfg.Cloudinary.APISecret)

	// 处理R2配置
	cfg.R2.Endpoint = expandEnvVars(cfg.R2.Endpoint)
	cfg.R2.AccountID = expandEnvVars(cfg.R2.AccountID)
	cfg.R2.AccessKeyID = expandEnvVars(cfg.R2.AccessKeyID)
	cfg.R2.SecretAccessKey = expandEnvVars(cfg.R2.SecretAccessKey)
	cfg.R2.Bucket = expandEnvVars(cfg.R2.Bucket)
	cfg.R2.PublicURL = expandEnvVars(cfg.R2.PublicURL)
}

// NewConfig 创建配置
func NewConfig() *Config {
	// 默认配置
	config := &Config{
		DB: DBConfig{
			Host:     "",
			Port:     "",
			User:     "",
			Password: "",
			DBName:   "",
			URL:      "",
		},
		Server: ServerConfig{
			Port: "8000", // 默认端口为8000
		},
		Cloudinary: CloudinaryConfig{
			CloudName: "",
			APIKey:    "",
			APISecret: "",
		},
		R2: R2Config{
			Endpoint:        "",
			AccountID:       "",
			AccessKeyID:     "",
			SecretAccessKey: "",
			Bucket:          "",
			PublicURL:       "",
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

	// PORT环境变量优先级最高
	if port := os.Getenv("PORT"); port != "" {
		config.Server.Port = port
	}

	return config
}
