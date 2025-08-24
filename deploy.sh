#!/bin/bash

# 部署脚本 - 以太坊钱包后端服务
# 支持多种部署方式：直接部署、Docker部署、云服务器部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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
    echo "    以太坊钱包后端服务部署脚本"
    echo "========================================"
    echo "请选择部署方式："
    echo "1) 本地直接部署"
    echo "2) Docker容器部署"
    echo "3) 生产环境部署（含SSL）"
    echo "4) 云服务器初始化"
    echo "5) 退出"
    echo "========================================"
}

# 检查依赖
check_dependencies() {
    print_step "检查系统依赖..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go 1.18+"
        exit 1
    fi
    
    print_message "Go 版本: $(go version)"
}

# 检查Docker依赖
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    print_message "Docker 版本: $(docker --version)"
    print_message "Docker Compose 版本: $(docker-compose --version)"
}

# 本地直接部署
deploy_local() {
    print_step "开始本地直接部署..."
    
    check_dependencies
    
    # 安装依赖
    print_step "安装 Go 依赖..."
    go mod tidy
    
    # 构建应用
    print_step "构建应用..."
    go build -o wallet-server main.go
    
    # 检查配置文件
    if [ ! -f "config/config.yaml" ]; then
        print_warning "配置文件不存在，请确保 config/config.yaml 配置正确"
        return 1
    fi
    
    # 启动服务
    print_step "启动服务..."
    print_message "服务将在 http://localhost:8087 运行"
    print_message "按 Ctrl+C 停止服务"
    ./wallet-server
}

# Docker部署
deploy_docker() {
    print_step "开始 Docker 容器部署..."
    
    check_docker
    
    # 构建镜像
    print_step "构建 Docker 镜像..."
    docker build -t wallet-backend:latest .
    
    # 启动容器
    print_step "启动 Docker 容器..."
    docker-compose up -d
    
    # 检查服务状态
    sleep 5
    if docker-compose ps | grep -q "Up"; then
        print_message "Docker 部署成功！"
        print_message "服务地址: http://localhost:8087"
        print_message "查看日志: docker-compose logs -f"
        print_message "停止服务: docker-compose down"
    else
        print_error "Docker 部署失败，请检查日志"
        docker-compose logs
    fi
}

# 生产环境部署
deploy_production() {
    print_step "开始生产环境部署..."
    
    check_docker
    
    # 检查SSL证书
    if [ ! -d "ssl" ]; then
        print_warning "SSL证书目录不存在，创建示例目录..."
        mkdir -p ssl
        print_warning "请将SSL证书放置在 ssl/ 目录下："
        print_warning "  - ssl/cert.pem (证书文件)"
        print_warning "  - ssl/key.pem (私钥文件)"
        return 1
    fi
    
    # 检查域名配置
    print_warning "请确保已在 nginx.conf 中配置正确的域名"
    
    # 设置生产环境变量
    export GIN_MODE=release
    
    # 启动生产环境
    print_step "启动生产环境..."
    docker-compose -f docker-compose.yml up -d
    
    print_message "生产环境部署完成！"
    print_message "HTTP将自动重定向到HTTPS"
    print_message "请确保防火墙开放80和443端口"
}

# 云服务器初始化
init_cloud_server() {
    print_step "云服务器环境初始化..."
    
    # 更新系统
    print_step "更新系统包..."
    sudo apt update && sudo apt upgrade -y
    
    # 安装基础工具
    print_step "安装基础工具..."
    sudo apt install -y curl wget git vim htop
    
    # 安装Docker
    print_step "安装 Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    
    # 安装Docker Compose
    print_step "安装 Docker Compose..."
    sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    
    # 安装Go
    print_step "安装 Go..."
    wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    
    # 配置防火墙
    print_step "配置防火墙..."
    sudo ufw allow 22
    sudo ufw allow 80
    sudo ufw allow 443
    sudo ufw allow 8087
    sudo ufw --force enable
    
    print_message "云服务器初始化完成！"
    print_message "请重新登录以生效环境变量"
    print_message "或运行: source ~/.bashrc"
}

# 显示服务状态
show_status() {
    print_step "检查服务状态..."
    
    if command -v docker &> /dev/null; then
        if docker-compose ps 2>/dev/null | grep -q "Up"; then
            print_message "Docker 服务运行中"
            docker-compose ps
        else
            print_warning "Docker 服务未运行"
        fi
    fi
    
    if pgrep -f "wallet-server" > /dev/null; then
        print_message "本地服务运行中"
    else
        print_warning "本地服务未运行"
    fi
}

# 主菜单循环
main() {
    while true; do
        show_menu
        read -p "请输入选项 (1-5): " choice
        
        case $choice in
            1)
                deploy_local
                ;;
            2)
                deploy_docker
                ;;
            3)
                deploy_production
                ;;
            4)
                init_cloud_server
                ;;
            5)
                print_message "退出部署脚本"
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