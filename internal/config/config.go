package config

// Config 应用配置
type Config struct {
	DB     DBConfig
	Server ServerConfig
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:     "postgres",
			Port:     "5432",
			User:     "blog_dev",
			Password: "dev123",
			DBName:   "blogdb_dev",
		},
		Server: ServerConfig{
			Port: "8000",
		},
	}
}
