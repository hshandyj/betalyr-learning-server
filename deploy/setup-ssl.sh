#!/bin/bash

# SSL证书设置脚本
set -e

DOMAIN="735566.xyz"
EMAIL="your-email@example.com"  # 请更改为您的邮箱

echo "🔐 开始设置SSL证书..."

# 检查域名解析
echo "🔍 检查域名解析..."
for subdomain in "" "www" "api"; do
    if [ -n "$subdomain" ]; then
        check_domain="${subdomain}.${DOMAIN}"
    else
        check_domain="$DOMAIN"
    fi
    
    echo "检查 $check_domain..."
    if ! nslookup $check_domain > /dev/null 2>&1; then
        echo "❌ $check_domain 域名解析失败，请检查DNS设置"
        exit 1
    fi
done

echo "✅ 域名解析正常"

# 停止nginx（如果正在运行）
sudo systemctl stop nginx || true

# 获取SSL证书
echo "📜 获取SSL证书..."
sudo certbot certonly \
    --standalone \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --domains $DOMAIN,www.$DOMAIN,api.$DOMAIN

# 复制nginx配置
echo "⚙️ 配置Nginx..."
sudo cp nginx.conf /etc/nginx/sites-available/betalyr-learning
sudo ln -sf /etc/nginx/sites-available/betalyr-learning /etc/nginx/sites-enabled/

# 移除默认配置
sudo rm -f /etc/nginx/sites-enabled/default

# 测试nginx配置
echo "🧪 测试Nginx配置..."
sudo nginx -t

# 启动nginx
echo "🔄 启动Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# 设置证书自动续期
echo "🔄 设置证书自动续期..."
sudo crontab -l 2>/dev/null | grep -v "certbot renew" | sudo crontab -
(sudo crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | sudo crontab -

# 测试证书续期
echo "🧪 测试证书续期..."
sudo certbot renew --dry-run

echo "✅ SSL证书设置完成！"
echo "🌐 您的网站现在可以通过HTTPS访问："
echo "  - https://$DOMAIN"
echo "  - https://www.$DOMAIN"
echo "  - https://api.$DOMAIN"

# 检查SSL证书状态
echo "🔍 检查SSL证书状态..."
openssl x509 -in /etc/letsencrypt/live/$DOMAIN/cert.pem -text -noout | grep "Not After" 