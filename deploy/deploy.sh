#!/bin/bash

# 服务器部署脚本
set -e

echo "🚀 开始部署 Betalyr Learning Server..."

# 检查必要的文件
if [ ! -f ".env.prod" ]; then
    echo "❌ .env.prod 文件不存在，请先创建生产环境配置文件"
    exit 1
fi

if [ ! -f "docker-compose.prod.yml" ]; then
    echo "❌ docker-compose.prod.yml 文件不存在"
    exit 1
fi

# 创建必要的目录
mkdir -p logs backups scripts

# 停止现有服务
echo "🛑 停止现有服务..."
docker compose -f docker-compose.prod.yml down --remove-orphans || true

# 拉取最新镜像
echo "📥 拉取最新镜像..."
docker compose -f docker-compose.prod.yml pull

# 启动服务
echo "🔄 启动服务..."
docker compose -f docker-compose.prod.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo "🔍 检查服务状态..."
docker compose -f docker-compose.prod.yml ps

# 验证健康检查
echo "🩺 验证服务健康状态..."
for i in {1..5}; do
    if curl -f http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ 服务运行正常！"
        break
    else
        if [ $i -eq 5 ]; then
            echo "❌ 服务启动失败，请检查日志"
            docker compose -f docker-compose.prod.yml logs app
            exit 1
        fi
        echo "⏳ 等待服务启动... ($i/5)"
        sleep 10
    fi
done

# 清理旧镜像
echo "🧹 清理旧镜像..."
docker image prune -f

echo "🎉 部署完成！"
echo "📊 服务状态："
docker compose -f docker-compose.prod.yml ps
echo ""
echo "🌐 访问地址："
echo "  - 主域名: https://735566.xyz"
echo "  - API: https://api.735566.xyz"
echo "  - 健康检查: https://735566.xyz/health" 