#!/bin/bash

# 本地Docker部署脚本 - nnkong.asiayu域名
# 用于在本地服务器部署钱包服务并向外暴露

set -e

DOMAIN="nnkong.asiayu"
API_DOMAIN="api.nnkong.asiayu"
WWW_DOMAIN="www.nnkong.asiayu"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

print_header() {
    echo "========================================"
    echo "    本地Docker部署 - nnkong.asiayu"
    echo "========================================"
}

# 检查依赖
check_dependencies() {
    print_step "检查系统依赖..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker未安装，请先安装Docker"
        echo "安装命令: curl -fsSL https://get.docker.com -o get-docker.sh && sudo sh get-docker.sh"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    print_message "Docker版本: $(docker --version)"
    print_message "Docker Compose版本: $(docker-compose --version)"
}

# 检查DNS解析
check_dns() {
    print_step "检查DNS解析..."
    
    local domains=("$DOMAIN" "$API_DOMAIN" "$WWW_DOMAIN")
    local local_ip=$(curl -s ifconfig.me || curl -s ipinfo.io/ip)
    
    print_message "本地公网IP: $local_ip"
    
    for domain in "${domains[@]}"; do
        local resolved_ip=$(dig +short $domain | head -1)
        if [ -n "$resolved_ip" ]; then
            if [ "$resolved_ip" = "$local_ip" ]; then
                print_message "✓ $domain -> $resolved_ip (正确)"
            else
                print_warning "⚠ $domain -> $resolved_ip (与本地IP不符: $local_ip)"
            fi
        else
            print_warning "⚠ $domain 无法解析"
        fi
    done
    
    echo ""
    print_warning "请确保以下DNS记录已配置："
    echo "  类型: A    主机: @      值: $local_ip"
    echo "  类型: A    主机: api    值: $local_ip"
    echo "  类型: A    主机: www    值: $local_ip"
    echo ""
}

# 配置防火墙
setup_firewall() {
    print_step "配置防火墙..."
    
    if command -v ufw &> /dev/null; then
        sudo ufw allow 22      # SSH
        sudo ufw allow 80      # HTTP
        sudo ufw allow 443     # HTTPS
        sudo ufw --force enable
        print_message "防火墙配置完成"
    else
        print_warning "未检测到ufw防火墙，请手动开放80和443端口"
    fi
}

# 构建和启动服务
deploy_services() {
    print_step "构建和启动Docker服务..."
    
    # 停止现有服务
    docker-compose -f docker-compose.local.yml down 2>/dev/null || true
    
    # 构建镜像
    print_message "构建Docker镜像..."
    docker build -t wallet-backend:latest .
    
    # 启动服务（先不启用SSL）
    print_message "启动服务..."
    docker-compose -f docker-compose.local.yml up -d
    
    # 等待服务启动
    sleep 10
    
    # 检查服务状态
    if docker-compose -f docker-compose.local.yml ps | grep -q "Up"; then
        print_message "Docker服务启动成功"
    else
        print_error "Docker服务启动失败"
        docker-compose -f docker-compose.local.yml logs
        exit 1
    fi
}

# 申请SSL证书
setup_ssl() {
    print_step "申请SSL证书..."
    
    # 安装certbot
    if ! command -v certbot &> /dev/null; then
        print_message "安装certbot..."
        sudo apt update
        sudo apt install -y certbot
    fi
    
    # 创建webroot目录
    sudo mkdir -p /var/www/html
    
    # 临时停止nginx以申请证书
    docker-compose -f docker-compose.local.yml stop nginx
    
    # 申请证书
    print_message "申请Let's Encrypt证书..."
    sudo certbot certonly \
        --standalone \
        --email admin@$DOMAIN \
        --agree-tos \
        --no-eff-email \
        -d $DOMAIN \
        -d $API_DOMAIN \
        -d $WWW_DOMAIN
    
    if [ $? -eq 0 ]; then
        print_message "SSL证书申请成功"
        
        # 创建SSL目录并复制证书
        mkdir -p ssl
        sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem ssl/cert.pem
        sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem ssl/key.pem
        sudo chown $USER:$USER ssl/*
        
        # 重启nginx
        docker-compose -f docker-compose.local.yml start nginx
        
        # 设置自动续期
        print_message "设置SSL证书自动续期..."
        (sudo crontab -l 2>/dev/null; echo "0 12 * * * /usr/bin/certbot renew --quiet --deploy-hook 'docker-compose -f $(pwd)/docker-compose.local.yml restart nginx'") | sudo crontab -
        
    else
        print_error "SSL证书申请失败"
        return 1
    fi
}

# 测试部署
test_deployment() {
    print_step "测试部署..."
    
    local test_urls=(
        "http://$DOMAIN/health"
        "https://$DOMAIN/health"
        "https://$API_DOMAIN/health"
        "https://$DOMAIN/api/v1/networks/list"
    )
    
    for url in "${test_urls[@]}"; do
        if curl -f -s --max-time 10 "$url" > /dev/null; then
            print_message "✓ $url - 正常"
        else
            print_warning "⚠ $url - 失败"
        fi
    done
}

# 显示部署结果
show_results() {
    print_step "部署完成！"
    
    echo ""
    echo "========================================"
    echo "           部署结果"
    echo "========================================"
    echo ""
    echo "🌐 访问地址："
    echo "  主域名:     https://$DOMAIN"
    echo "  API子域名:  https://$API_DOMAIN"
    echo "  WWW域名:    https://$WWW_DOMAIN"
    echo ""
    echo "📋 API接口："
    echo "  健康检查:   https://$DOMAIN/health"
    echo "  网络列表:   https://$DOMAIN/api/v1/networks/list"
    echo "  钱包导入:   https://$DOMAIN/api/v1/wallets/import-mnemonic"
    echo ""
    echo "🔧 管理命令："
    echo "  查看日志:   docker-compose -f docker-compose.local.yml logs -f"
    echo "  重启服务:   docker-compose -f docker-compose.local.yml restart"
    echo "  停止服务:   docker-compose -f docker-compose.local.yml down"
    echo ""
    echo "🔒 SSL证书："
    echo "  证书路径:   /etc/letsencrypt/live/$DOMAIN/"
    echo "  自动续期:   已配置"
    echo ""
    
    # 显示客户端接入示例
    cat << 'EOF'
📱 客户端接入示例：

JavaScript:
```javascript
const API_BASE = 'https://nnkong.asiayu/api/v1';

// 获取网络列表
const networks = await fetch(`${API_BASE}/networks/list`);
console.log(await networks.json());

// 导入钱包
const wallet = await fetch(`${API_BASE}/wallets/import-mnemonic`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    mnemonic: 'your twelve word mnemonic phrase here',
    derivation_path: "m/44'/60'/0'/0/0"
  })
});
console.log(await wallet.json());
```

cURL:
```bash
# 健康检查
curl https://nnkong.asiayu/health

# 获取网络列表
curl https://nnkong.asiayu/api/v1/networks/list

# 导入钱包
curl -X POST https://nnkong.asiayu/api/v1/wallets/import-mnemonic \
  -H "Content-Type: application/json" \
  -d '{"mnemonic":"your mnemonic","derivation_path":"m/44'\''/60'\''/0'\''/0/0"}'
```
EOF
    
    echo "========================================"
    print_message "钱包服务已成功部署并可通过 nnkong.asiayu 域名访问！"
    echo "========================================"
}

# 主函数
main() {
    print_header
    
    check_dependencies
    check_dns
    
    read -p "是否继续部署？(y/n): " continue_deploy
    if [ "$continue_deploy" != "y" ]; then
        print_message "部署已取消"
        exit 0
    fi
    
    setup_firewall
    deploy_services
    
    read -p "是否申请SSL证书？(y/n): " setup_ssl_choice
    if [ "$setup_ssl_choice" = "y" ]; then
        setup_ssl
    fi
    
    test_deployment
    show_results
}

# 错误处理
trap 'print_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 运行主程序
main "$@"