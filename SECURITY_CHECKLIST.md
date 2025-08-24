# 🔐 公网部署安全配置清单

## 服务器安全配置

### 1. 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw deny 8087     # 禁止直接访问应用端口
sudo ufw enable

# 检查状态
sudo ufw status
```

### 2. SSH安全配置
```bash
# 禁用root登录
sudo vim /etc/ssh/sshd_config
# 设置：PermitRootLogin no

# 修改SSH端口（可选）
# Port 2022

# 重启SSH服务
sudo systemctl restart sshd
```

### 3. 创建非root用户
```bash
# 创建部署用户
sudo adduser deploy
sudo usermod -aG sudo deploy
sudo usermod -aG docker deploy

# 配置SSH密钥登录
ssh-copy-id deploy@your-server-ip
```

## 应用安全配置

### 1. 修改默认密钥（重要！）
```yaml
# config/config.yaml
security:
  jwt_secret: "your-super-secret-jwt-key-at-least-32-characters-long"
  encryption_key: "your-encryption-key-exactly-32-characters"
  rate_limit:
    general: 1000      # 提高生产环境限制
    transaction: 50    # 交易频率限制
    auth: 20          # 认证频率限制
```

### 2. 环境变量配置
```bash
# 创建 .env 文件
cat > .env << EOF
# 安全配置
JWT_SECRET=your-production-jwt-secret-key
ENCRYPTION_KEY=your-production-encryption-key

# RPC配置
ETHEREUM_RPC_URL=https://mainnet.infura.io/v3/YOUR_PROJECT_ID
POLYGON_RPC_URL=https://polygon-mainnet.infura.io/v3/YOUR_PROJECT_ID

# 服务器配置
GIN_MODE=release
SERVER_PORT=8087
EOF
```

### 3. Nginx安全配置
```nginx
# 在nginx.conf中添加安全头
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

# 隐藏服务器信息
server_tokens off;

# 限制请求大小
client_max_body_size 1M;

# 超时配置
client_body_timeout 12;
client_header_timeout 12;
send_timeout 10;
```

## 监控和日志

### 1. 日志配置
```bash
# 创建日志目录
sudo mkdir -p /var/log/wallet
sudo chown deploy:deploy /var/log/wallet

# Docker日志配置
# 在docker-compose.yml中添加
services:
  wallet-backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 2. 系统监控
```bash
# 安装监控工具
sudo apt install htop iotop nethogs

# 检查系统资源
htop                    # CPU和内存
df -h                   # 磁盘空间
free -h                 # 内存使用
sudo netstat -tlnp      # 端口监听
```

## 备份策略

### 1. 配置文件备份
```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/opt/backup/wallet"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份配置文件
tar -czf $BACKUP_DIR/config_$DATE.tar.gz config/
tar -czf $BACKUP_DIR/ssl_$DATE.tar.gz /etc/letsencrypt/

# 删除30天前的备份
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "备份完成: $DATE"
```

### 2. 定期任务
```bash
# 添加到crontab
sudo crontab -e

# 每天凌晨2点备份
0 2 * * * /opt/wallet/backup.sh

# 每周重启服务
0 4 * * 0 docker-compose restart

# SSL证书自动续期
0 12 * * * /usr/bin/certbot renew --quiet
```