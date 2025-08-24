# 🚀 以太坊钱包后端服务部署指南

本文档详细介绍如何将以太坊钱包后端服务部署到生产环境，让其他人能够通过互联网访问。

## 📋 目录

1. [部署方式对比](#部署方式对比)
2. [快速部署](#快速部署)
3. [云服务器部署](#云服务器部署)
4. [Docker容器部署](#docker容器部署)
5. [域名和SSL配置](#域名和ssl配置)
6. [安全配置](#安全配置)
7. [监控和维护](#监控和维护)
8. [故障排除](#故障排除)

## 🔄 部署方式对比

| 部署方式 | 难度 | 成本 | 推荐场景 | 优缺点 |
|---------|------|------|----------|---------|
| **云服务器直接部署** | ⭐⭐ | 💰💰 | 小型项目、快速上线 | 简单快速，但需要手动维护 |
| **Docker容器部署** | ⭐⭐⭐ | 💰💰 | 中型项目、需要扩展 | 标准化部署，易于管理 |
| **Kubernetes部署** | ⭐⭐⭐⭐⭐ | 💰💰💰 | 大型项目、高可用 | 复杂但功能强大 |
| **Serverless部署** | ⭐⭐⭐ | 💰 | 不定期使用 | 按需付费，冷启动延迟 |

## ⚡ 快速部署

使用提供的部署脚本快速部署：

```bash
# 1. 克隆项目
git clone <your-repository-url>
cd wallet

# 2. 运行部署脚本
./deploy.sh

# 3. 选择部署方式
# 选项1: 本地直接部署（测试用）
# 选项2: Docker容器部署（推荐）
# 选项3: 生产环境部署（含SSL）
# 选项4: 云服务器初始化
```

## ☁️ 云服务器部署

### 第一步：购买云服务器

#### 推荐云服务商

**国内用户：**
- [阿里云ECS](https://ecs.console.aliyun.com/)
- [腾讯云CVM](https://console.cloud.tencent.com/cvm)
- [华为云ECS](https://console.huaweicloud.com/ecm/)

**国外用户：**
- [AWS EC2](https://aws.amazon.com/ec2/)
- [DigitalOcean](https://www.digitalocean.com/)
- [Vultr](https://www.vultr.com/)

#### 服务器配置建议

**最低配置：**
- CPU: 1核心
- 内存: 2GB
- 存储: 20GB SSD
- 带宽: 1Mbps

**推荐配置：**
- CPU: 2核心
- 内存: 4GB
- 存储: 40GB SSD
- 带宽: 5Mbps
- 操作系统: Ubuntu 20.04 LTS

### 第二步：服务器初始化

```bash
# 连接到服务器
ssh root@your-server-ip

# 运行初始化脚本
curl -fsSL https://raw.githubusercontent.com/your-username/wallet/main/deploy.sh -o deploy.sh
chmod +x deploy.sh
./deploy.sh
# 选择选项4进行云服务器初始化
```

### 第三步：部署应用

```bash
# 1. 克隆代码
git clone https://github.com/your-username/wallet.git
cd wallet

# 2. 配置环境变量
cp .env.example .env
vim .env  # 修改配置

# 3. 运行部署
./deploy.sh
# 选择选项2进行Docker部署
```

## 🐳 Docker容器部署

### 本地构建部署

```bash
# 1. 构建镜像
docker build -t wallet-backend:latest .

# 2. 运行容器
docker run -d \\
  --name wallet-backend \\
  -p 8087:8087 \\
  -v $(pwd)/config:/root/config:ro \\
  -v $(pwd)/keystores:/root/keystores \\
  --restart unless-stopped \\
  wallet-backend:latest

# 3. 检查状态
docker ps
docker logs wallet-backend
```

### 使用Docker Compose

```bash
# 1. 启动服务
docker-compose up -d

# 2. 查看日志
docker-compose logs -f

# 3. 停止服务
docker-compose down
```

### 生产环境配置

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  wallet-backend:
    build: .
    ports:
      - "8087:8087"
    environment:
      - GIN_MODE=release
      - CONFIG_PATH=/root/config/config.prod.yaml
    volumes:
      - ./config:/root/config:ro
      - ./keystores:/root/keystores
      - /var/log/wallet:/var/log/wallet
    restart: unless-stopped
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
    networks:
      - wallet-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - /var/log/nginx:/var/log/nginx
    depends_on:
      - wallet-backend
    restart: unless-stopped
    networks:
      - wallet-network

networks:
  wallet-network:
    driver: bridge
```

## 🌐 域名和SSL配置

### 购买域名

推荐域名注册商：
- [阿里云域名](https://wanwang.aliyun.com/)
- [腾讯云域名](https://dnspod.cloud.tencent.com/)
- [Namecheap](https://www.namecheap.com/)
- [Cloudflare](https://www.cloudflare.com/)

### DNS配置

```bash
# 添加A记录指向服务器IP
# 例如：
# Type: A
# Name: api (或 @)
# Value: your-server-ip
# TTL: 300
```

### SSL证书申请

#### 方法1：使用Let's Encrypt（免费）

```bash
# 安装certbot
sudo apt install certbot python3-certbot-nginx

# 申请证书
sudo certbot --nginx -d your-domain.com

# 自动续期
sudo crontab -e
# 添加：0 12 * * * /usr/bin/certbot renew --quiet
```

#### 方法2：使用云服务商SSL证书

大多数云服务商提供免费的SSL证书，配置更简单。

### Nginx SSL配置

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location /api/ {
        proxy_pass http://localhost:8087;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔒 安全配置

### 基础安全配置

```bash
# 1. 更新系统
sudo apt update && sudo apt upgrade -y

# 2. 配置防火墙
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw enable

# 3. 禁用root登录
sudo vim /etc/ssh/sshd_config
# 设置: PermitRootLogin no
sudo systemctl restart sshd

# 4. 创建普通用户
sudo adduser deploy
sudo usermod -aG sudo deploy
sudo usermod -aG docker deploy
```

### 应用安全配置

```yaml
# config/config.prod.yaml
security:
  jwt_secret: "your-very-long-and-random-jwt-secret-key-at-least-32-characters"
  encryption_key: "exactly-32-characters-encryption-key"
  rate_limit:
    general: 1000
    transaction: 50
    auth: 20
  cors:
    allowed_origins: 
      - "https://your-frontend-domain.com"
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
```

### 环境变量安全

```bash
# 使用环境变量替代配置文件中的敏感信息
export JWT_SECRET="your-jwt-secret"
export ENCRYPTION_KEY="your-encryption-key"
export ETHEREUM_RPC_URL="https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
```

## 📊 监控和维护

### 系统监控

```bash
# 1. 安装监控工具
sudo apt install htop iotop nethogs

# 2. 查看系统状态
htop                    # CPU和内存使用
df -h                   # 磁盘使用
free -h                 # 内存使用
sudo netstat -tlnp      # 端口监听状态
```

### 应用监控

```bash
# 查看应用状态
docker ps
docker stats wallet-backend

# 查看应用日志
docker logs -f wallet-backend
tail -f /var/log/wallet/app.log

# 健康检查
curl http://localhost:8087/health
```

### 备份策略

```bash
#!/bin/bash
# backup.sh - 备份脚本

BACKUP_DIR="/opt/backup/wallet"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建备份目录
mkdir -p $BACKUP_DIR

# 备份配置文件
tar -czf $BACKUP_DIR/config_$DATE.tar.gz config/

# 备份keystore文件
tar -czf $BACKUP_DIR/keystores_$DATE.tar.gz keystores/

# 删除30天前的备份
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "备份完成: $DATE"
```

### 定期维护

```bash
# 添加到crontab
sudo crontab -e

# 每天凌晨2点备份
0 2 * * * /opt/wallet/backup.sh

# 每周清理Docker
0 3 * * 0 docker system prune -f

# 每月重启服务
0 4 1 * * docker-compose restart
```

## 🔧 故障排除

### 常见问题

#### 1. 服务无法启动

```bash
# 检查端口占用
sudo netstat -tlnp | grep 8087

# 检查配置文件
cat config/config.yaml

# 查看详细错误
docker logs wallet-backend
```

#### 2. 无法连接区块链节点

```bash
# 测试RPC连接
curl -X POST \\
  -H "Content-Type: application/json" \\
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \\
  https://eth.llamarpc.com

# 更换RPC节点
# 修改 config/config.yaml 中的 rpc_url
```

#### 3. SSL证书问题

```bash
# 检查证书有效期
sudo certbot certificates

# 手动续期
sudo certbot renew

# 测试HTTPS
curl -I https://your-domain.com/health
```

#### 4. 性能问题

```bash
# 检查系统资源
htop
df -h
free -h

# 检查Docker资源
docker stats

# 优化配置
# 增加rate_limit限制
# 使用付费RPC节点
# 增加服务器配置
```

## 📝 部署检查清单

- [ ] 服务器配置充足（CPU、内存、存储）
- [ ] 域名解析正确指向服务器IP
- [ ] 防火墙配置正确（开放80、443端口）
- [ ] SSL证书申请并配置
- [ ] 修改默认的JWT密钥和加密密钥
- [ ] 配置生产级RPC节点（Infura/Alchemy）
- [ ] 设置适当的速率限制
- [ ] 配置日志记录和监控
- [ ] 设置自动备份
- [ ] 测试所有API接口
- [ ] 进行安全性测试

## 🎯 访问地址

部署完成后，你的服务将在以下地址可用：

- **HTTP**: `http://your-domain.com` (自动重定向到HTTPS)
- **HTTPS**: `https://your-domain.com`
- **API文档**: `https://your-domain.com/api/v1/docs`
- **健康检查**: `https://your-domain.com/health`

## 📞 技术支持

如果在部署过程中遇到问题，可以：

1. 查看本文档的故障排除章节
2. 检查项目的GitHub Issues
3. 查看服务日志获取详细错误信息
4. 联系技术支持团队

---

## 总结

本指南提供了完整的部署方案，从简单的云服务器部署到完整的生产环境配置。选择适合你项目规模和技术水平的部署方式，并严格遵循安全配置建议，确保服务的稳定运行和数据安全。