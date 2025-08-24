# 🚀 5分钟快速部署指南

这是最简单的部署方式，让你快速将钱包服务部署到云服务器。

## 🎯 部署目标

部署完成后，其他人可以通过以下方式访问：
- **API服务**: `http://your-server-ip:8087`
- **或者域名**: `https://your-domain.com` (配置域名后)

## 📋 准备工作

1. **购买云服务器**（任选一家）：
   - [阿里云ECS](https://ecs.console.aliyun.com/) - 国内推荐
   - [腾讯云CVM](https://console.cloud.tencent.com/cvm) - 国内推荐
   - [DigitalOcean](https://www.digitalocean.com/) - 海外推荐

2. **服务器配置**：
   - CPU: 2核心
   - 内存: 4GB
   - 系统: Ubuntu 20.04

## ⚡ 一键部署

### 方案一：使用部署脚本（推荐）

```bash
# 1. 连接到服务器
ssh root@your-server-ip

# 2. 下载并运行部署脚本
curl -fsSL https://raw.githubusercontent.com/your-username/wallet/main/deploy.sh -o deploy.sh
chmod +x deploy.sh
./deploy.sh

# 3. 选择部署选项
# 首次部署选择：4（云服务器初始化）
# 然后选择：2（Docker容器部署）
```

### 方案二：手动部署

```bash
# 1. 连接服务器
ssh root@your-server-ip

# 2. 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 3. 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 4. 克隆项目
git clone https://github.com/your-username/wallet.git
cd wallet

# 5. 启动服务
docker-compose up -d

# 6. 检查状态
docker-compose ps
```

## 🔧 配置防火墙

```bash
# 开放必要端口
sudo ufw allow 22      # SSH
sudo ufw allow 8087    # API服务
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw enable
```

## 🎉 完成！

服务现在已经运行，可以通过以下地址访问：

- **API测试**: `http://your-server-ip:8087/health`
- **获取网络信息**: `http://your-server-ip:8087/api/v1/networks`

## 🌐 配置域名（可选）

如果你有域名，可以配置域名访问：

### 快速配置（推荐）
```bash
# 使用自动化脚本
./setup-domain.sh
# 按照提示输入域名和服务器IP即可
```

### 手动配置
1. **添加DNS记录**：
   - 类型: A
   - 名称: @ 或 api
   - 值: your-server-ip

2. **申请SSL证书**：
```bash
# 安装certbot
sudo apt install certbot python3-certbot-nginx

# 申请免费证书
sudo certbot --nginx -d your-domain.com
```

3. **修改配置文件**：
```bash
# 编辑nginx.conf，将your-domain.com替换为你的域名
vim nginx.conf

# 重启服务
docker-compose restart
```

### 📖 详细指南
需要完整的域名配置指南？查看 [DOMAIN_SETUP.md](./DOMAIN_SETUP.md)

## 📊 检查服务状态

```bash
# 检查Docker容器
docker ps

# 查看服务日志
docker logs wallet-backend

# 测试API
curl http://your-server-ip:8087/health
```

## 🔒 安全提醒

**重要：请务必修改默认密钥！**

```bash
# 编辑配置文件
vim config/config.yaml

# 修改以下配置：
security:
  jwt_secret: "your-new-secret-key"
  encryption_key: "your-new-encryption-key"

# 重启服务
docker-compose restart
```

## 📱 客户端访问示例

部署完成后，客户端可以这样访问：

```javascript
// JavaScript示例
const API_BASE = 'http://your-server-ip:8087/api/v1';

// 获取网络信息
fetch(`${API_BASE}/networks`)
  .then(response => response.json())
  .then(data => console.log(data));

// 导入钱包
fetch(`${API_BASE}/wallets/import-mnemonic`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    mnemonic: 'your twelve word mnemonic phrase here',
    derivation_path: "m/44'/60'/0'/0/0"
  })
})
.then(response => response.json())
.then(data => console.log('钱包地址:', data.data.address));
```

## 🆘 遇到问题？

1. **服务无法启动**：
   ```bash
   docker logs wallet-backend
   ```

2. **无法访问**：
   ```bash
   # 检查端口
   sudo netstat -tlnp | grep 8087
   
   # 检查防火墙
   sudo ufw status
   ```

3. **查看完整文档**：
   - 阅读 `DEPLOYMENT.md` 获取详细部署指南
   - 查看 `README.md` 了解更多功能

---

**恭喜！你的以太坊钱包服务现在已经可以通过互联网访问了！** 🎉