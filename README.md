# 博客服务端

这是一个使用 Go 语言开发的博客系统后端服务，采用清晰的分层架构设计，提供 RESTful API 接口。

## 项目特性

- 使用 Gin 框架构建 RESTful API
- PostgreSQL 数据库存储
- Docker 容器化部署
- 支持热重载开发
- 清晰的分层架构（Handler、Service、Repository）

## 项目结构

```
.
├── cmd/
│   └── betalyr-learning-server/        # 主程序入口
├── internal/
│   ├── blog/              
│   │   ├── handler/       # HTTP 处理器
│   │   ├── service/       # 业务逻辑层
│   │   ├── repository/    # 数据访问层
│   │   └── models/        # 数据模型
│   ├── config/            # 配置管理
│   └── database/          # 数据库连接
├── migrations/            # 数据库迁移文件
├── scripts/              # 工具脚本
├── .air.toml             # Air 配置文件
├── docker-compose.yml    # Docker Compose 配置
└── Dockerfile            # Docker 构建文件
```

## 环境要求

- Go 1.21+
- Docker 和 Docker Compose
- PostgreSQL 15
- Redis 7

## 快速开始

1. **克隆项目**
   ```bash
   git clone <项目地址>
   cd betalyr-learning-server
   ```

2. **启动开发环境**
   ```bash
   # 启动 Docker 容器
   docker-compose up -d
   ```

3. **运行服务**
   ```bash
   # 使用 air 启动开发服务器（支持热重载）
   air
   ```

服务将在 http://localhost:8000 启动

## API 文档

### 文章管理接口

#### 创建文章
- **URL**: `/api/articles`
- **方法**: POST
- **请求体**:
  ```json
  {
    "userId": "用户ID",
    "title": "文章标题",
    "excerpt": "文章摘要",
    "content": "文章内容",
    "author": "作者",
    "tags": ["标签1", "标签2"]
  }
  ```

#### 获取文章列表
- **URL**: `/api/articles?page=1&page_size=10`
- **方法**: GET
- **参数**:
  - page: 页码
  - page_size: 每页数量

#### 获取单篇文章
- **URL**: `/api/articles/:id`
- **方法**: GET

#### 更新文章
- **URL**: `/api/articles/:id`
- **方法**: PUT

#### 删除文章
- **URL**: `/api/articles/:id`
- **方法**: DELETE

## 数据库配置

数据库配置信息在 `internal/config/config.go` 中：

```go
DB: DBConfig{
    Host:     "postgres",
    Port:     "5432",
    User:     "betalyr_lerning_dev",
    Password: "dev123",
    DBName:   "betalyr_lerningdb_dev",
}
```

## 开发指南

1. **修改配置**
   - 数据库配置在 `internal/config/config.go`
   - 服务器配置也在同一文件中

2. **添加新功能**
   - 在 `internal/blog/models` 中添加新的数据模型
   - 在 `internal/blog/repository` 中实现数据访问逻辑
   - 在 `internal/blog/service` 中实现业务逻辑
   - 在 `internal/blog/handler` 中添加新的 API 处理器

3. **数据库迁移**
   - 数据库表会在服务启动时自动创建
   - 使用 GORM 的 AutoMigrate 功能

## 部署

1. **使用 Docker Compose 部署**
   ```bash
   # 构建镜像
   docker-compose build
   
   # 启动服务
   docker-compose up -d
   ```

2. **手动部署**
   ```bash
   # 编译
   go build -o betalyr-learning-server ./cmd/betalyr-learning-server
   
   # 运行
   ./betalyr-learning-server
   ```

## 注意事项

1. 确保 PostgreSQL 和 Redis 服务正常运行
2. 首次运行时会自动创建数据库表
3. 开发环境建议使用 `air` 实现热重载
4. 生产环境部署时注意修改配置文件中的敏感信息

## 自动部署配置

本项目配置了GitHub Actions，在main分支打tag时自动部署到Fly.io。

### 设置步骤

1. 获取Fly.io API令牌:
   ```bash
   fly auth token
   ```

2. 在GitHub仓库中添加密钥:
   - 进入仓库设置 -> Secrets and variables -> Actions
   - 点击"New repository secret"
   - 名称填写: `FLY_API_TOKEN`
   - 值填写: 上一步获取的令牌
   - 点击"Add secret"

3. 推送tag触发部署:
   ```bash
   git tag v1.0.0  # 创建标签
   git push origin v1.0.0  # 推送标签
   ```

4. 自动部署将在GitHub Actions中运行，完成后应用将可通过以下地址访问:
   https://betalyr-learning-server.fly.dev

### 注意事项

- `.dockerignore`文件格式：如果您在Windows环境中开发，请确保.dockerignore文件使用Linux风格的路径分隔符(/)而不是Windows风格(\)。GitHub Actions工作流会自动修复这个问题，但本地构建时您可能需要手动调整。

- 部署日志：可以在GitHub仓库的Actions页面查看部署日志和状态。

## 贡献指南

欢迎提交 Issue 和 Pull Request

## 许可证

MIT License