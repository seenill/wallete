#!/bin/bash

# DNS配置检查脚本
# 用于验证域名DNS设置是否正确

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}       DNS配置检查工具${NC}"
    echo -e "${BLUE}======================================${NC}"
}

check_domain() {
    local domain=$1
    local expected_ip=$2
    
    echo -e "\n${BLUE}[检查]${NC} 域名: $domain"
    echo "----------------------------------------"
    
    # 检查A记录
    local resolved_ip=$(dig +short $domain | head -1)
    if [ -n "$resolved_ip" ]; then
        if [ "$resolved_ip" = "$expected_ip" ]; then
            echo -e "${GREEN}✓${NC} A记录正确: $resolved_ip"
        else
            echo -e "${RED}✗${NC} A记录错误: $resolved_ip (期望: $expected_ip)"
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
    
    # 获取用户输入
    read -p "请输入要检查的域名: " domain
    read -p "请输入预期的服务器IP: " server_ip
    
    if [ -z "$domain" ] || [ -z "$server_ip" ]; then
        echo -e "${RED}错误: 域名和IP地址不能为空${NC}"
        exit 1
    fi
    
    # 检查主域名
    check_domain "$domain" "$server_ip"
    
    # 检查www子域名
    check_domain "www.$domain" "$server_ip"
    
    # 检查api子域名
    check_domain "api.$domain" "$server_ip"
    
    echo -e "\n${BLUE}======================================${NC}"
    echo -e "${BLUE}           检查完成${NC}"
    echo -e "${BLUE}======================================${NC}"
    
    echo -e "\n${YELLOW}注意事项:${NC}"
    echo "1. DNS记录修改后通常需要10分钟-24小时生效"
    echo "2. 如果A记录正确但连接失败，请检查防火墙设置"
    echo "3. SSL证书需要单独配置，可使用 ./setup-domain.sh"
    
    echo -e "\n${GREEN}推荐配置:${NC}"
    echo "DNS记录配置："
    echo "  类型: A    主机: @      值: $server_ip"
    echo "  类型: A    主机: www    值: $server_ip"
    echo "  类型: A    主机: api    值: $server_ip"
}

# 运行主程序
main