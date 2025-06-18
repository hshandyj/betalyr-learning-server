#!/bin/bash

# SSL证书设置脚本
set -e

DOMAIN="735566.xyz"
EMAIL="your-email@example.com"  # 请更改为您的邮箱

echo "🔐 Setting up SSL certificates..."

# 检查域名解析
echo "🔍 Checking domain name resolution..."
for subdomain in "" "www" "api"; do
    if [ -n "$subdomain" ]; then
        check_domain="${subdomain}.${DOMAIN}"
    else
        check_domain="$DOMAIN"
    fi
    
    echo "Checking $check_domain..."
    if ! nslookup $check_domain > /dev/null 2>&1; then
        echo "❌ $check_domain domain name resolution failed, please check DNS settings"
        exit 1
    fi
done

echo "✅ Domain name resolution is normal"

# 停止nginx（如果正在运行）
sudo systemctl stop nginx || true

# 获取SSL证书
echo "📜 Getting SSL certificates..."
sudo certbot certonly \
    --standalone \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    --domains $DOMAIN,www.$DOMAIN,api.$DOMAIN

# 复制nginx配置
echo "⚙️ Configuring Nginx..."
sudo cp nginx.conf /etc/nginx/sites-available/betalyr-learning
sudo ln -sf /etc/nginx/sites-available/betalyr-learning /etc/nginx/sites-enabled/

# 移除默认配置
sudo rm -f /etc/nginx/sites-enabled/default

# 测试nginx配置
echo "🧪 Testing Nginx configuration..."
sudo nginx -t

# 启动nginx
echo "🔄 Starting Nginx..."
sudo systemctl start nginx
sudo systemctl enable nginx

# 设置证书自动续期
echo "🔄 Setting up certificate auto-renewal..."
sudo crontab -l 2>/dev/null | grep -v "certbot renew" | sudo crontab -
(sudo crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | sudo crontab -

# 测试证书续期
echo "🧪 Testing certificate renewal..."
sudo certbot renew --dry-run

echo "✅ SSL certificates setup complete!"
echo "🌐 Your website is now accessible via HTTPS:"
echo "  - https://$DOMAIN"
echo "  - https://www.$DOMAIN"
echo "  - https://api.$DOMAIN"

# 检查SSL证书状态
echo "🔍 Checking SSL certificate status..."
openssl x509 -in /etc/letsencrypt/live/$DOMAIN/cert.pem -text -noout | grep "Not After" 