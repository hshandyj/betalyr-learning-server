#!/bin/bash

# 监控和维护脚本
set -e

LOGFILE="/var/log/betalyr-monitoring.log"
PROJECT_DIR="/opt/betalyr-learning"

# 日志函数
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOGFILE
}

# 检查服务状态
check_services() {
    log "🔍 检查服务状态..."
    
    cd $PROJECT_DIR
    
    # 检查容器状态
    if ! docker compose -f docker-compose.prod.yml ps | grep -q "Up"; then
        log "❌ 容器服务异常，尝试重启..."
        docker compose -f docker-compose.prod.yml up -d
        sleep 30
    fi
    
    # 检查健康端点
    if ! curl -f http://localhost:8000/health > /dev/null 2>&1; then
        log "❌ 健康检查失败，尝试重启服务..."
        docker compose -f docker-compose.prod.yml restart app
        sleep 30
        
        # 再次检查
        if ! curl -f http://localhost:8000/health > /dev/null 2>&1; then
            log "❌ 服务重启后仍然异常，需要人工干预"
            return 1
        fi
    fi
    
    log "✅ 服务运行正常"
    return 0
}

# 检查磁盘空间
check_disk_space() {
    log "💾 检查磁盘空间..."
    
    DISK_USAGE=$(df / | awk 'NR==2{print $5}' | sed 's/%//')
    
    if [ $DISK_USAGE -gt 80 ]; then
        log "⚠️ 磁盘空间不足 ($DISK_USAGE%)，开始清理..."
        
        # 清理Docker
        docker system prune -f
        docker volume prune -f
        
        # 清理旧日志
        find /var/log -name "*.log" -mtime +7 -delete
        find $PROJECT_DIR/logs -name "*.log" -mtime +7 -delete
        
        # 清理旧备份（保留7天）
        find $PROJECT_DIR/backups -name "*.sql" -mtime +7 -delete
        
        log "🧹 清理完成"
    else
        log "✅ 磁盘空间充足 ($DISK_USAGE%)"
    fi
}

# 检查内存使用
check_memory() {
    log "🧠 检查内存使用..."
    
    MEMORY_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    
    if [ $MEMORY_USAGE -gt 85 ]; then
        log "⚠️ 内存使用率过高 ($MEMORY_USAGE%)，尝试优化..."
        
        # 重启应用容器释放内存
        docker compose -f docker-compose.prod.yml restart app
        
        log "🔄 已重启应用容器"
    else
        log "✅ 内存使用正常 ($MEMORY_USAGE%)"
    fi
}

# 备份数据库
backup_database() {
    log "💾 开始数据库备份..."
    
    cd $PROJECT_DIR
    
    # 获取容器中的数据库连接信息
    DB_CONTAINER=$(docker compose -f docker-compose.prod.yml ps -q postgres)
    
    if [ -n "$DB_CONTAINER" ]; then
        BACKUP_FILE="backups/manual_backup_$(date +%Y%m%d_%H%M%S).sql"
        
        docker exec $DB_CONTAINER pg_dump -U betalyr_user betalyr_learning > $BACKUP_FILE
        
        if [ $? -eq 0 ]; then
            log "✅ 数据库备份完成: $BACKUP_FILE"
        else
            log "❌ 数据库备份失败"
            return 1
        fi
    else
        log "❌ 找不到数据库容器"
        return 1
    fi
}

# 更新SSL证书
renew_ssl() {
    log "🔐 检查SSL证书更新..."
    
    if sudo certbot renew --quiet; then
        log "✅ SSL证书检查完成"
        sudo systemctl reload nginx
    else
        log "❌ SSL证书更新失败"
        return 1
    fi
}

# 主函数
main() {
    log "🚀 开始系统监控和维护..."
    
    # 检查服务
    if ! check_services; then
        log "❌ 服务检查失败"
        exit 1
    fi
    
    # 检查资源
    check_disk_space
    check_memory
    
    # 如果是周日，执行备份
    if [ $(date +%u) -eq 7 ]; then
        backup_database
    fi
    
    # 如果是每月1号，更新SSL证书
    if [ $(date +%d) -eq 01 ]; then
        renew_ssl
    fi
    
    log "✅ 监控和维护完成"
}

# 执行参数处理
case "${1:-main}" in
    "check")
        check_services
        ;;
    "backup")
        backup_database
        ;;
    "ssl")
        renew_ssl
        ;;
    "clean")
        check_disk_space
        ;;
    "main")
        main
        ;;
    *)
        echo "用法: $0 [check|backup|ssl|clean|main]"
        exit 1
        ;;
esac 