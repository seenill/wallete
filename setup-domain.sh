#!/bin/bash

# 域名配置自动化脚本
# 用于快速配置域名和SSL证书

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 显示菜单
show_menu() {
    echo "========================================"
    echo "    域名配置自动化脚本"
    echo "========================================"
    echo "选择操作："
    echo "1) 配置域名和Nginx"
    echo "2) 申请Let's Encrypt SSL证书"
    echo "3) 测试域名和SSL"
    echo "4) 查看当前配置"
    echo "5) 退出"
    echo "========================================"
}

# 检查依赖
check_dependencies() {
    print_step "检查系统依赖..."
    
    if ! command -v nginx &> /dev/null; then
        print_warning "Nginx未安装，尝试安装..."
        sudo apt update
        sudo apt install -y nginx
    fi
    
    if ! command -v certbot &> /dev/null; then
        print_warning "Certbot未安装，尝试安装..."
        sudo apt install -y certbot python3-certbot-nginx
    fi
    
    print_message "依赖检查完成"
}

# 配置域名
configure_domain() {
    print_step "开始域名配置..."
    
    # 获取用户输入
    read -p "请输入你的域名（例如：my-wallet.com）: " DOMAIN
    read -p "请输入你的服务器IP地址: " SERVER_IP
    
    # 验证输入
    if [ -z "$DOMAIN" ] || [ -z "$SERVER_IP" ]; then
        print_error "域名和IP地址不能为空"
        return 1
    fi
    
    print_message "域名: $DOMAIN"
    print_message "服务器IP: $SERVER_IP"
    
    # 测试DNS解析
    print_step "测试DNS解析..."
    if nslookup $DOMAIN | grep -q $SERVER_IP; then
        print_message "DNS解析正确"
    else
        print_warning "DNS解析可能还未生效，请确保已配置A记录："
        echo "  类型: A"
        echo "  主机记录: @"
        echo "  记录值: $SERVER_IP"
        echo ""
        read -p "是否继续配置？(y/n): " continue_setup
        if [ "$continue_setup" != "y" ]; then
            return 1
        fi
    fi
    
    # 备份原有配置
    if [ -f "nginx.conf" ]; then
        cp nginx.conf nginx.conf.backup.$(date +%Y%m%d_%H%M%S)
        print_message "已备份原有Nginx配置"
    fi
    
    # 更新Nginx配置
    print_step "更新Nginx配置..."
    sed -i.bak "s/your-domain\.com/$DOMAIN/g" nginx.conf
    
    # 创建简化的临时配置（仅HTTP，用于Let's Encrypt验证）
    cat > nginx.conf.temp << EOF
events {
    worker_connections 1024;
}

http {
    upstream wallet_backend {
        server wallet-backend:8087;
    }

    server {
        listen 80;
        server_name $DOMAIN api.$DOMAIN www.$DOMAIN;
        
        # Let's Encrypt验证路径
        location /.well-known/acme-challenge/ {
            root /var/www/html;
        }
        
        # API代理
        location /api/ {
            proxy_pass http://wallet_backend;
            proxy_set_header Host \$host;
            proxy_set_header X-Real-IP \$remote_addr;
            proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto \$scheme;
        }
        
        # 健康检查
        location /health {
            proxy_pass http://wallet_backend/health;
        }
        
        # 默认重定向到HTTPS（SSL配置后启用）
        # return 301 https://\$server_name\$request_uri;
    }
}
EOF

    print_message "Nginx配置已更新"
    
    # 重启服务
    print_step "重启Nginx服务..."
    if command -v docker-compose &> /dev/null; then
        docker-compose down
        cp nginx.conf.temp nginx.conf
        docker-compose up -d
        print_message "Docker服务已重启"
    else
        print_warning "请手动重启Nginx服务"
    fi
    
    print_message "域名配置完成！"
    print_message "临时访问地址: http://$DOMAIN/health"
}

# 申请SSL证书
setup_ssl() {
    print_step "开始SSL证书申请..."
    
    if [ -z "$DOMAIN" ]; then
        read -p "请输入你的域名: " DOMAIN
    fi
    
    # 检查域名解析
    print_step "检查域名解析..."
    if ! nslookup $DOMAIN > /dev/null 2>&1; then
        print_error "域名解析失败，请先配置DNS记录"
        return 1
    fi
    
    # 创建webroot目录
    sudo mkdir -p /var/www/html
    
    # 申请证书
    print_step "申请Let's Encrypt证书..."
    sudo certbot certonly \
        --webroot \
        --webroot-path=/var/www/html \
        --email admin@$DOMAIN \
        --agree-tos \
        --no-eff-email \
        -d $DOMAIN \
        -d api.$DOMAIN \
        -d www.$DOMAIN
    
    if [ $? -eq 0 ]; then
        print_message "SSL证书申请成功！"
        
        # 更新Nginx配置为完整版本（包含HTTPS）
        print_step "更新Nginx配置为HTTPS版本..."
        sed -i.bak "s/your-domain\.com/$DOMAIN/g" nginx.conf
        sed -i "s|/etc/nginx/ssl/cert.pem|/etc/letsencrypt/live/$DOMAIN/fullchain.pem|g" nginx.conf
        sed -i "s|/etc/nginx/ssl/key.pem|/etc/letsencrypt/live/$DOMAIN/privkey.pem|g" nginx.conf
        
        # 重启服务
        if command -v docker-compose &> /dev/null; then
            docker-compose down
            docker-compose up -d
        fi
        
        # 设置自动续期
        print_step "设置SSL证书自动续期..."
        (sudo crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet") | sudo crontab -
        
        print_message "SSL配置完成！"
        print_message "HTTPS访问地址: https://$DOMAIN/health"
    else
        print_error "SSL证书申请失败"
        return 1
    fi
}

# 测试配置
test_configuration() {
    print_step "测试域名和SSL配置..."
    
    if [ -z "$DOMAIN" ]; then
        read -p "请输入你的域名: " DOMAIN
    fi
    
    # 测试HTTP访问
    print_step "测试HTTP访问..."
    if curl -f http://$DOMAIN/health > /dev/null 2>&1; then
        print_message "HTTP访问正常"
    else
        print_error "HTTP访问失败"
    fi
    
    # 测试HTTPS访问
    print_step "测试HTTPS访问..."
    if curl -f https://$DOMAIN/health > /dev/null 2>&1; then
        print_message "HTTPS访问正常"
    else
        print_warning "HTTPS访问失败（可能还未配置SSL）"
    fi
    
    # 测试API接口
    print_step "测试API接口..."
    if curl -f https://$DOMAIN/api/v1/networks/list > /dev/null 2>&1; then
        print_message "API接口正常"
    else
        print_warning "API接口测试失败"
    fi
    
    # 显示证书信息
    if [ -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
        print_step "SSL证书信息:"
        sudo openssl x509 -in /etc/letsencrypt/live/$DOMAIN/fullchain.pem -text -noout | grep -E "(Issuer|Not Before|Not After|Subject:)"
    fi
    
    print_message "测试完成！"
}

# 查看当前配置
show_current_config() {
    print_step "当前配置信息:"
    
    # 显示Nginx状态
    if command -v docker-compose &> /dev/null; then
        echo "Docker服务状态:"
        docker-compose ps
    fi
    
    # 显示证书信息
    echo ""
    echo "SSL证书:"
    sudo ls -la /etc/letsencrypt/live/ 2>/dev/null || echo "未找到SSL证书"
    
    # 显示DNS信息
    echo ""
    if [ ! -z "$DOMAIN" ]; then
        echo "域名DNS信息:"
        nslookup $DOMAIN 2>/dev/null || echo "域名未设置"
    fi
    
    # 显示访问地址
    echo ""
    echo "访问地址:"
    echo "  HTTP:  http://$DOMAIN/health"
    echo "  HTTPS: https://$DOMAIN/health"
    echo "  API:   https://$DOMAIN/api/v1/networks/list"
}

# 主菜单循环
main() {
    check_dependencies
    
    while true; do
        show_menu
        read -p "请输入选项 (1-5): " choice
        
        case $choice in
            1)
                configure_domain
                ;;
            2)
                setup_ssl
                ;;
            3)
                test_configuration
                ;;
            4)
                show_current_config
                ;;
            5)
                print_message "退出域名配置脚本"
                exit 0
                ;;
            *)
                print_error "无效选项，请重新选择"
                ;;
        esac
        
        echo
        read -p "按回车键继续..."
        echo
    done
}

# 运行主程序
main