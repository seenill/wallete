# 🌐 域名配置详细指南

本指南将手把手教你如何为以太坊钱包服务配置域名，让其他人可以通过 `https://your-domain.com` 访问你的服务。

## 📋 配置流程

1. [购买域名](#1-购买域名)
2. [DNS配置](#2-dns配置)
3. [修改服务器配置](#3-修改服务器配置)
4. [申请SSL证书](#4-申请ssl证书)
5. [测试访问](#5-测试访问)

---

## 1. 购买域名

### 🛒 推荐域名注册商

#### 国内用户推荐
- **[阿里云域名](https://wanwang.aliyun.com/)**
  - 优势：与阿里云服务器集成好，管理方便
  - 价格：.com域名 约55元/年
  
- **[腾讯云域名](https://dnspod.cloud.tencent.com/)**
  - 优势：DNS解析速度快，与腾讯云集成
  - 价格：.com域名 约55元/年

#### 国外用户推荐
- **[Namecheap](https://www.namecheap.com/)**
  - 优势：价格便宜，界面友好
  - 价格：.com域名 约$8-12/年
  
- **[Cloudflare](https://www.cloudflare.com/)**
  - 优势：免费CDN和DDoS防护
  - 价格：.com域名 约$8.03/年（按成本价）

### 🎯 域名选择建议

```bash
# ✅ 推荐格式
my-wallet-api.com
blockchain-wallet.net  
crypto-service.org
eth-wallet.io

# ❌ 避免格式
123wallet.com          # 数字开头不专业
wallet_api.com         # 包含下划线
walletapi.cn          # 太短可能被抢注
```

---

## 2. DNS配置

### 📍 添加DNS记录

登录你的域名管理后台，按以下方式配置：

#### 基础配置（根域名）
```
记录类型: A
主机记录: @
记录值: your-server-ip-address
TTL: 600
```

#### API子域名配置（推荐）
```
记录类型: A
主机记录: api
记录值: your-server-ip-address  
TTL: 600
```

#### 配置示例

假设你的服务器IP是 `123.456.789.0`，域名是 `my-wallet.com`：

| 类型 | 主机记录 | 记录值 | TTL |
|------|----------|--------|-----|
| A | @ | 123.456.789.0 | 600 |
| A | api | 123.456.789.0 | 600 |
| A | www | 123.456.789.0 | 600 |

配置后的访问地址：
- `https://my-wallet.com`
- `https://api.my-wallet.com` 
- `https://www.my-wallet.com`

### 🕐 等待DNS生效

DNS记录通常需要 10分钟-24小时 生效，可以用以下方式检查：

```bash
# 方法1：使用dig命令
dig your-domain.com

# 方法2：使用nslookup
nslookup your-domain.com

# 方法3：使用ping
ping your-domain.com
```

**在线检查工具：**
- [DNSChecker](https://dnschecker.org/)
- [WhatsMyDNS](https://www.whatsmydns.net/)

---

## 3. 修改服务器配置

### 🔧 方法一：修改Nginx配置文件

编辑 `nginx.conf` 文件，将 `your-domain.com` 替换为你的实际域名：

```bash
# 连接到服务器
ssh root@your-server-ip

# 进入项目目录
cd /path/to/your/wallet

# 编辑Nginx配置
vim nginx.conf
```

修改以下行：
```nginx
# 将这一行
server_name your-domain.com;

# 改为你的实际域名
server_name api.my-wallet.com my-wallet.com;
```

### 🔧 方法二：使用环境变量

编辑 `.env` 文件：
```bash
# 设置域名
DOMAIN=my-wallet.com
API_DOMAIN=api.my-wallet.com
```

### 🔄 重启服务

```bash
# 如果使用Docker Compose
docker-compose down
docker-compose up -d

# 或者重启Nginx
docker-compose restart nginx
```

---

## 4. 申请SSL证书

### 🔒 方法一：Let's Encrypt免费证书（推荐）

#### 安装Certbot
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install certbot python3-certbot-nginx
```

#### 申请证书
```bash
# 为你的域名申请证书
sudo certbot --nginx -d your-domain.com -d api.your-domain.com

# 示例
sudo certbot --nginx -d my-wallet.com -d api.my-wallet.com
```

#### 自动续期
```bash
# 测试自动续期
sudo certbot renew --dry-run

# 设置定时任务
sudo crontab -e
# 添加以下行（每天检查续期）
0 12 * * * /usr/bin/certbot renew --quiet
```

### 🔒 方法二：云服务商SSL证书

#### 阿里云SSL证书
1. 登录阿里云控制台
2. 搜索"SSL证书"
3. 申请免费证书（1年有效期）
4. 下载证书文件
5. 上传到服务器

#### 腾讯云SSL证书
1. 登录腾讯云控制台  
2. 进入"SSL证书管理"
3. 申请免费证书
4. 按照指引完成验证

### 🔧 手动配置SSL证书

如果使用云服务商证书，需要手动配置：

```bash
# 创建SSL目录
sudo mkdir -p /etc/nginx/ssl

# 上传证书文件
# cert.pem - 证书文件
# key.pem - 私钥文件

# 修改nginx.conf中的证书路径
ssl_certificate /etc/nginx/ssl/cert.pem;
ssl_certificate_key /etc/nginx/ssl/key.pem;
```

---

## 5. 测试访问

### 🧪 基础连通性测试

```bash
# 测试HTTP访问（会自动跳转到HTTPS）
curl -I http://your-domain.com

# 测试HTTPS访问
curl -I https://your-domain.com

# 测试API接口
curl https://your-domain.com/health
curl https://your-domain.com/api/v1/networks/list
```

### 🧪 浏览器测试

在浏览器中访问以下地址：

1. **健康检查**: `https://your-domain.com/health`
   - 应该返回：`{"status":"ok","message":"...","timestamp":...}`

2. **API文档**: `https://your-domain.com/api/v1/networks/list`
   - 应该返回支持的网络列表

3. **检查SSL证书**: 点击浏览器地址栏的锁图标
   - 确认证书有效且安全

### 🧪 完整API测试

```javascript
// 在浏览器控制台中运行
const API_BASE = 'https://your-domain.com/api/v1';

// 测试网络接口
fetch(`${API_BASE}/networks/list`)
  .then(r => r.json())
  .then(data => console.log('网络列表:', data));

// 测试钱包导入（需要有效助记词）
fetch(`${API_BASE}/wallets/import-mnemonic`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    mnemonic: 'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about',
    derivation_path: "m/44'/60'/0'/0/0"
  })
})
.then(r => r.json())
.then(data => console.log('钱包地址:', data));
```

---

## 📝 完整配置示例

假设你的域名是 `my-wallet.com`，服务器IP是 `123.456.789.0`：

### DNS配置
```
A    @     123.456.789.0    600
A    api   123.456.789.0    600  
A    www   123.456.789.0    600
```

### Nginx配置
```nginx
server {
    listen 80;
    server_name my-wallet.com api.my-wallet.com www.my-wallet.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name my-wallet.com api.my-wallet.com www.my-wallet.com;
    
    ssl_certificate /etc/letsencrypt/live/my-wallet.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/my-wallet.com/privkey.pem;
    
    # ... 其他配置
}
```

### 访问地址
```
https://my-wallet.com/health
https://my-wallet.com/api/v1/networks/list
https://api.my-wallet.com/api/v1/networks/list
```

---

## 🔧 故障排除

### 问题1：域名无法访问
```bash
# 检查DNS解析
nslookup your-domain.com

# 检查服务器防火墙
sudo ufw status
sudo ufw allow 80
sudo ufw allow 443

# 检查Nginx状态
docker-compose ps
docker logs nginx_container_name
```

### 问题2：SSL证书错误
```bash
# 检查证书文件
sudo ls -la /etc/letsencrypt/live/your-domain.com/

# 测试Nginx配置
sudo nginx -t

# 重新申请证书
sudo certbot delete --cert-name your-domain.com
sudo certbot --nginx -d your-domain.com
```

### 问题3：API无法访问
```bash
# 检查后端服务
docker ps
docker logs wallet-backend

# 检查端口监听
sudo netstat -tlnp | grep 8087

# 测试内部连接
curl http://localhost:8087/health
```

---

## 💡 高级配置

### CDN加速（可选）
使用Cloudflare为你的域名提供CDN加速：

1. 注册Cloudflare账号
2. 添加你的域名
3. 将域名DNS服务器改为Cloudflare提供的
4. 启用SSL/TLS加密模式

### 域名邮箱（可选）
配置 `admin@your-domain.com` 邮箱：

1. 在域名管理中添加MX记录
2. 使用腾讯企业邮箱或阿里云邮箱
3. 用于接收SSL证书通知和系统告警

### 子域名规划
```
api.your-domain.com    - API服务
app.your-domain.com    - 前端应用  
admin.your-domain.com  - 管理后台
docs.your-domain.com   - API文档
```

---

## ✅ 配置完成检查清单

- [ ] 域名已购买并实名认证
- [ ] DNS A记录配置正确
- [ ] 服务器防火墙开放80和443端口
- [ ] Nginx配置文件更新域名
- [ ] SSL证书申请成功
- [ ] HTTPS访问正常
- [ ] API接口测试通过
- [ ] 设置SSL证书自动续期

完成以上步骤后，你的钱包服务就可以通过专业的域名访问了！🎉

**示例最终访问地址：**
- `https://your-domain.com/health` - 健康检查
- `https://your-domain.com/api/v1/networks/list` - 获取网络列表
- `https://your-domain.com/api/v1/wallets/import-mnemonic` - 导入钱包