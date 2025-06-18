# Betalyr Learning Server 生产环境部署指南

## 📋 概述

本指南将帮助您将 Betalyr Learning Server 部署到京东云服务器，并配置域名、SSL证书和自动化部署流程。

## 🏗️ 架构图

```
用户 → Cloudflare DNS → 京东云服务器 → Nginx → Go应用 → PostgreSQL/Redis
```

## 🚀 部署步骤

### 1. 域名和DNS配置

#### 1.1 配置Cloudflare DNS
1. 登录 [Cloudflare Dashboard](https://dash.cloudflare.com)
2. 添加域名 `735566.xyz`
3. 在DNS设置中添加以下记录：
   ```
   Type: A, Name: @, Content: 117.72.96.174, TTL: Auto
   Type: A, Name: www, Content: 117.72.96.174, TTL: Auto
   Type: A, Name: api, Content: 117.72.96.174, TTL: Auto
   ```
4. 在 Spaceship 控制台将域名的 Name Servers 更改为 Cloudflare 提供的地址

### 2. 服务器初始化

#### 2.1 连接到京东云服务器
```bash
ssh root@117.72.96.174
```

#### 2.2 运行服务器初始化脚本
```bash
# 下载脚本
wget https://raw.githubusercontent.com/your-username/betalyr-learning-server/main/deploy/server-setup.sh
chmod +x server-setup.sh
./server-setup.sh
```

#### 2.3 重新登录使docker组权限生效
```bash
exit
ssh root@117.72.96.174
```

### 3. 部署应用

#### 3.1 创建项目目录并下载配置文件
```bash
cd /opt/betalyr-learning
git clone https://github.com/your-username/betalyr-learning-server.git .
```

#### 3.2 配置环境变量
```bash
cp deploy/env.prod.example .env.prod
vim .env.prod
```

填入以下配置：
```env
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_USER=betalyr_user
DB_PASSWORD=your_secure_password_here
DB_NAME=betalyr_learning

# 服务器配置
SERVER_PORT=8000

# Cloudinary配置
CLOUDINARY_CLOUD_NAME=your_cloudinary_cloud_name
CLOUDINARY_API_KEY=your_cloudinary_api_key
CLOUDINARY_API_SECRET=your_cloudinary_api_secret

# R2配置
R2_ENDPOINT=your_r2_endpoint
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key_id
R2_SECRET_ACCESS_KEY=your_r2_secret_access_key
R2_BUCKET=your_r2_bucket_name
R2_PUBLIC_URL=your_r2_public_url
```

#### 3.3 设置SSL证书
```bash
cd /opt/betalyr-learning/deploy
# 修改邮箱地址
vim setup-ssl.sh
chmod +x setup-ssl.sh
./setup-ssl.sh
```

#### 3.4 首次部署
```bash
cd /opt/betalyr-learning
chmod +x deploy/deploy.sh
./deploy/deploy.sh
```

### 4. GitHub Actions配置

#### 4.1 在GitHub仓库中设置Secrets
在 GitHub 仓库的 Settings → Secrets and variables → Actions 中添加：

```
DOCKER_USERNAME: 您的Docker Hub用户名
DOCKER_PASSWORD: 您的Docker Hub密码
HOST: 117.72.96.174
USERNAME: root
PRIVATE_KEY: 您的SSH私钥内容
```

#### 4.2 生成SSH密钥对（如果没有）
```bash
# 在本地机器上运行
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
cat ~/.ssh/id_rsa.pub
```

将公钥内容添加到服务器的 `~/.ssh/authorized_keys` 文件中：
```bash
# 在服务器上运行
mkdir -p ~/.ssh
echo "your_public_key_content" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

### 5. 自动化部署测试

#### 5.1 创建并推送tag
```bash
git tag v1.0.0
git push origin v1.0.0
```

这将触发自动部署流程。

### 6. 监控和维护

#### 6.1 设置定时监控
```bash
cd /opt/betalyr-learning
chmod +x deploy/monitoring.sh

# 添加到crontab
crontab -e
```

添加以下行：
```cron
# 每5分钟检查一次服务状态
*/5 * * * * /opt/betalyr-learning/deploy/monitoring.sh check

# 每天凌晨2点运行完整监控
0 2 * * * /opt/betalyr-learning/deploy/monitoring.sh main
```

## 🔧 维护命令

### 日常维护
```bash
# 查看服务状态
docker compose -f deploy/docker-compose.prod.yml ps

# 查看日志
docker compose -f deploy/docker-compose.prod.yml logs app

# 重启服务
docker compose -f deploy/docker-compose.prod.yml restart app

# 手动备份数据库
./deploy/monitoring.sh backup

# 检查SSL证书
./deploy/monitoring.sh ssl
```

### 紧急恢复
```bash
# 停止所有服务
docker compose -f deploy/docker-compose.prod.yml down

# 恢复数据库备份
docker exec -i betalyr-postgres psql -U betalyr_user -d betalyr_learning < backups/backup_YYYYMMDD_HHMMSS.sql

# 重新启动服务
docker compose -f deploy/docker-compose.prod.yml up -d
```

## 📊 访问地址

部署完成后，您可以通过以下地址访问服务：

- **主域名**: https://735566.xyz
- **API接口**: https://api.735566.xyz
- **健康检查**: https://735566.xyz/health

## ⚠️ 注意事项

1. **安全性**：
   - 定期更新系统和Docker镜像
   - 使用强密码
   - 定期备份数据库

2. **监控**：
   - 定期检查服务运行状态
   - 监控磁盘空间和内存使用
   - 关注SSL证书到期时间

3. **备份**：
   - 自动备份保留7天
   - 重要更新前手动备份
   - 定期测试备份恢复

## 🆘 常见问题

### Q: 部署失败怎么办？
A: 查看GitHub Actions日志，检查服务器日志，确保环境变量配置正确。

### Q: SSL证书获取失败？
A: 确保域名解析正确，检查防火墙设置，确保80和443端口开放。

### Q: 服务访问缓慢？
A: 检查服务器资源使用情况，考虑优化数据库查询或增加服务器配置。

## 📞 技术支持

如遇到问题，请检查：
1. GitHub Actions 构建日志
2. 服务器 `/var/log/betalyr-monitoring.log` 日志
3. Docker 容器日志：`docker logs betalyr-learning-app` 