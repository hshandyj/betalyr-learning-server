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

### 健康检查接口

#### 根路径健康状态
- **URL**: `/`
- **方法**: GET
- **响应**:
  ```json
  {
    "status": "success",
  }
  ```
  ```
####重启服务
lsof -i :8000
kill -9 [PID]