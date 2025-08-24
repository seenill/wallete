#!/bin/bash

# nnkong.asiayu DNS配置检查脚本

DOMAIN="nnkong.asiayu"
SUBDOMAINS=("api.nnkong.asiayu" "www.nnkong.asiayu")

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}    nnkong.asiayu DNS配置检查${NC}"
    echo -e "${BLUE}======================================${NC}"
}

check_domain() {
    local domain=$1
    local local_ip=$(curl -s ifconfig.me 2>/dev/null || curl -s ipinfo.io/ip 2>/dev/null || echo "无法获取")
    
    echo -e "\n${BLUE}[检查]${NC} 域名: $domain"
    echo "----------------------------------------"
    
    # 检查A记录
    local resolved_ip=$(dig +short $domain 2>/dev/null | head -1)
    if [ -n "$resolved_ip" ]; then
        if [ "$resolved_ip" = "$local_ip" ]; then
            echo -e "${GREEN}✓${NC} A记录正确: $resolved_ip"
        else
            echo -e "${RED}✗${NC} A记录错误: $resolved_ip (本地IP: $local_ip)"
        fi
    else
        echo -e "${RED}✗${NC} 无法解析域名"
    fi
    
    # 检查HTTP连通性
    if curl -s --max-time 5 http://$domain/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} HTTP连接正常"
    else
        echo -e "${YELLOW}⚠${NC} HTTP连接失败"
    fi
    
    # 检查HTTPS连通性
    if curl -s --max-time 5 https://$domain/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} HTTPS连接正常"
    else
        echo -e "${YELLOW}⚠${NC} HTTPS连接失败（可能未配置SSL）"
    fi
    
    # 检查SSL证书
    local ssl_info=$(echo | openssl s_client -servername $domain -connect $domain:443 2>/dev/null | openssl x509 -noout -dates 2>/dev/null)
    if [ -n "$ssl_info" ]; then
        echo -e "${GREEN}✓${NC} SSL证书存在"
        echo "   $ssl_info"
    else
        echo -e "${YELLOW}⚠${NC} 未检测到SSL证书"
    fi
}

main() {
    print_header
    
    local local_ip=$(curl -s ifconfig.me 2>/dev/null || curl -s ipinfo.io/ip 2>/dev/null)
    echo -e "${BLUE}本地公网IP:${NC} $local_ip"
    
    # 检查主域名
    check_domain "$DOMAIN"
    
    # 检查子域名
    for subdomain in "${SUBDOMAINS[@]}"; do
        check_domain "$subdomain"
    done
    
    echo -e "\n${BLUE}======================================${NC}"
    echo -e "${BLUE}           检查完成${NC}"
    echo -e "${BLUE}======================================${NC}"
    
    echo -e "\n${YELLOW}DNS配置建议:${NC}"
    echo "如果DNS记录不正确，请在域名管理后台配置："
    echo "  类型: A    主机: @      值: $local_ip"
    echo "  类型: A    主机: api    值: $local_ip" 
    echo "  类型: A    主机: www    值: $local_ip"
    
    echo -e "\n${GREEN}部署命令:${NC}"
    echo "DNS配置正确后，运行以下命令部署："
    echo "  ./deploy-local.sh"
    
    echo -e "\n${GREEN}访问地址:${NC}"
    echo "部署完成后可通过以下地址访问："
    echo "  https://nnkong.asiayu/health"
    echo "  https://nnkong.asiayu/api/v1/networks/list"
    echo "  https://api.nnkong.asiayu/health"
}

# 运行主程序
main