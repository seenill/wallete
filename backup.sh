#!/bin/bash

# 钱包服务自动备份脚本
# 用于备份重要配置文件和数据

set -e

# 配置
BACKUP_DIR="/opt/backup/wallet"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_message() {
    echo -e "${GREEN}[BACKUP]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 创建备份目录
create_backup_dir() {
    print_message "创建备份目录: $BACKUP_DIR"
    mkdir -p $BACKUP_DIR
}

# 备份配置文件
backup_config() {
    print_message "备份配置文件..."
    
    if [ -d "config" ]; then
        tar -czf $BACKUP_DIR/config_$DATE.tar.gz config/
        print_message "配置文件备份完成"
    else
        print_warning "配置目录不存在"
    fi
}

# 备份SSL证书
backup_ssl() {
    print_message "备份SSL证书..."
    
    if [ -d "/etc/letsencrypt" ]; then
        sudo tar -czf $BACKUP_DIR/ssl_$DATE.tar.gz /etc/letsencrypt/
        print_message "SSL证书备份完成"
    else
        print_warning "SSL证书目录不存在"
    fi
}

# 备份Docker配置
backup_docker() {
    print_message "备份Docker配置..."
    
    local files=(
        "docker-compose.yml"
        "docker-compose.prod.yml"
        "Dockerfile"
        "nginx.conf"
        ".env"
    )
    
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "$BACKUP_DIR/${file}_$DATE"
        fi
    done
    
    print_message "Docker配置备份完成"
}

# 备份日志文件（最近7天）
backup_logs() {
    print_message "备份日志文件..."
    
    if [ -d "/var/log/wallet" ]; then
        find /var/log/wallet -name "*.log" -mtime -7 -exec tar -czf $BACKUP_DIR/logs_$DATE.tar.gz {} +
        print_message "日志文件备份完成"
    else
        print_warning "日志目录不存在"
    fi
}

# 清理旧备份
cleanup_old_backups() {
    print_message "清理${RETENTION_DAYS}天前的备份..."
    
    find $BACKUP_DIR -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete
    find $BACKUP_DIR -name "*_*" -mtime +$RETENTION_DAYS -delete
    
    print_message "旧备份清理完成"
}

# 生成备份报告
generate_report() {
    local report_file="$BACKUP_DIR/backup_report_$DATE.txt"
    
    cat > $report_file << EOF
钱包服务备份报告
================
备份时间: $(date)
备份位置: $BACKUP_DIR

备份文件:
$(ls -la $BACKUP_DIR/*$DATE* 2>/dev/null || echo "无备份文件")

磁盘使用情况:
$(df -h $BACKUP_DIR)

系统状态:
$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "Docker未运行")
EOF

    print_message "备份报告生成: $report_file"
}

# 发送备份通知（可选）
send_notification() {
    local status=$1
    local message=$2
    
    # 可以集成邮件、钉钉、企业微信等通知
    # 这里只是记录到系统日志
    logger "钱包服务备份: $status - $message"
}

# 主备份流程
main() {
    print_message "开始钱包服务备份..."
    
    # 检查权限
    if [ "$EUID" -ne 0 ] && [ ! -w "$BACKUP_DIR" ]; then
        print_error "权限不足，请使用sudo运行或确保有写入权限"
        exit 1
    fi
    
    # 执行备份
    create_backup_dir
    backup_config
    backup_ssl
    backup_docker
    backup_logs
    cleanup_old_backups
    generate_report
    
    print_message "备份完成！备份位置: $BACKUP_DIR"
    send_notification "SUCCESS" "备份完成"
}

# 错误处理
trap 'print_error "备份过程中发生错误"; send_notification "FAILED" "备份失败"; exit 1' ERR

# 运行备份
main "$@"