# 🚀 nnkong.asiayu 本地Docker部署指南

本指南将帮你在本地服务器上部署钱包服务，并通过 `nnkong.asiayu` 域名向外暴露。

## 📋 部署前准备

### 1. 确保DNS配置正确

在域名管理后台添加以下DNS记录：

```
类型: A    主机: @      值: [你的公网IP]
类型: A    主机: api    值: [你的公网IP]  
类型: A    主机: www    值: [你的公网IP]
```

### 2. 检查DNS是否生效

```bash
# 运行DNS检查脚本
./check-nnkong-dns.sh
```

### 3. 确保端口开放

确保你的路由器/防火墙开放了以下端口：
- **80** (HTTP)
- **443** (HTTPS)
- **22** (SSH，如果需要远程管理)

## 🐳 一键部署

### 快速部署命令

```bash
# 1. 运行部署脚本
./deploy-local.sh

# 脚本会自动完成：
# - 检查系统依赖
# - 验证DNS配置
# - 构建Docker镜像
# - 启动服务
# - 申请SSL证书
# - 配置HTTPS访问
```

### 手动分步部署

如果你想分步骤操作：

```bash
# 1. 检查DNS配置
./check-nnkong-dns.sh

# 2. 构建并启动服务
docker-compose -f docker-compose.local.yml build
docker-compose -f docker-compose.local.yml up -d

# 3. 申请SSL证书（可选）
sudo certbot certonly --standalone -d nnkong.asiayu -d api.nnkong.asiayu -d www.nnkong.asiayu

# 4. 重启Nginx
docker-compose -f docker-compose.local.yml restart nginx
```

## 🌐 访问地址

部署完成后，你的钱包服务将在以下地址可用：

### HTTP访问（会自动重定向到HTTPS）
- http://nnkong.asiayu
- http://api.nnkong.asiayu
- http://www.nnkong.asiayu

### HTTPS访问（推荐）
- **主域名**: https://nnkong.asiayu
- **API子域名**: https://api.nnkong.asiayu
- **WWW域名**: https://www.nnkong.asiayu

### API接口示例
- **健康检查**: https://nnkong.asiayu/health
- **网络列表**: https://nnkong.asiayu/api/v1/networks/list
- **钱包导入**: https://nnkong.asiayu/api/v1/wallets/import-mnemonic

## 📱 客户端接入示例

### JavaScript
```javascript
const API_BASE = 'https://nnkong.asiayu/api/v1';

// 获取支持的网络
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
console.log('钱包地址:', (await wallet.json()).data.address);

// 查询余额
const balance = await fetch(`${API_BASE}/wallets/0x.../balance`);
console.log(await balance.json());
```

### cURL
```bash
# 健康检查
curl https://nnkong.asiayu/health

# 获取网络列表
curl https://nnkong.asiayu/api/v1/networks/list

# 导入钱包
curl -X POST https://nnkong.asiayu/api/v1/wallets/import-mnemonic \
  -H "Content-Type: application/json" \
  -d '{
    "mnemonic": "your twelve word mnemonic phrase",
    "derivation_path": "m/44'\'''/60'\'''/0'\'''/0/0"
  }'

# 查询余额
curl "https://nnkong.asiayu/api/v1/wallets/0x.../balance"
```

### Python
```python
import requests

API_BASE = 'https://nnkong.asiayu/api/v1'

# 获取网络列表
response = requests.get(f'{API_BASE}/networks/list')
print(response.json())

# 导入钱包
wallet_data = {
    'mnemonic': 'your twelve word mnemonic phrase',
    'derivation_path': "m/44'/60'/0'/0/0"
}
response = requests.post(f'{API_BASE}/wallets/import-mnemonic', json=wallet_data)
print('钱包地址:', response.json())
```

## 🔧 管理命令

### 查看服务状态
```bash
# 查看容器状态
docker-compose -f docker-compose.local.yml ps

# 查看服务日志
docker-compose -f docker-compose.local.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.local.yml logs -f wallet-backend
docker-compose -f docker-compose.local.yml logs -f nginx
```

### 重启服务
```bash
# 重启所有服务
docker-compose -f docker-compose.local.yml restart

# 重启特定服务
docker-compose -f docker-compose.local.yml restart wallet-backend
docker-compose -f docker-compose.local.yml restart nginx
```

### 停止服务
```bash
# 停止所有服务
docker-compose -f docker-compose.local.yml down

# 停止并删除数据卷
docker-compose -f docker-compose.local.yml down -v
```

### 更新服务
```bash
# 重新构建并启动
docker-compose -f docker-compose.local.yml down
docker-compose -f docker-compose.local.yml build --no-cache
docker-compose -f docker-compose.local.yml up -d
```

## 🔒 安全注意事项

### 1. 修改默认密钥
部署前请修改 `config/config.nnkong.yaml` 中的安全配置：

```yaml
security:
  jwt_secret: "your-secure-random-jwt-secret-32-characters-long"
  encryption_key: "your-secure-random-encryption-key-32ch"
```

### 2. 防火墙配置
```bash
# 开放必要端口
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP
sudo ufw allow 443     # HTTPS
sudo ufw enable
```

### 3. SSL证书自动续期
SSL证书每90天需要续期，部署脚本已自动配置：

```bash
# 检查自动续期设置
sudo crontab -l | grep certbot

# 手动测试续期
sudo certbot renew --dry-run
```

## 🔍 故障排除

### 1. 域名无法访问
```bash
# 检查DNS解析
nslookup nnkong.asiayu
dig nnkong.asiayu

# 检查本地IP
curl ifconfig.me
```

### 2. 服务无法启动
```bash
# 查看错误日志
docker-compose -f docker-compose.local.yml logs

# 检查端口占用
sudo netstat -tlnp | grep :80
sudo netstat -tlnp | grep :443
```

### 3. SSL证书问题
```bash
# 检查证书状态
sudo certbot certificates

# 重新申请证书
sudo certbot delete --cert-name nnkong.asiayu
sudo certbot certonly --standalone -d nnkong.asiayu
```

### 4. API接口404错误
```bash
# 检查后端服务
curl http://localhost:8087/health

# 检查Nginx配置
docker-compose -f docker-compose.local.yml exec nginx nginx -t
```

## 📊 监控和维护

### 健康检查
```bash
# 检查所有服务健康状态
curl https://nnkong.asiayu/health

# 检查API接口
curl https://nnkong.asiayu/api/v1/networks/list
```

### 日志管理
```bash
# 查看实时日志
docker-compose -f docker-compose.local.yml logs -f --tail=100

# 清理旧日志
sudo truncate -s 0 /var/log/wallet/*.log
```

### 备份
```bash
# 备份配置文件
tar -czf backup-$(date +%Y%m%d).tar.gz config/ ssl/ docker-compose.local.yml

# 备份SSL证书
sudo tar -czf ssl-backup-$(date +%Y%m%d).tar.gz /etc/letsencrypt/
```

## 🎉 部署完成检查清单

- [ ] DNS记录配置正确
- [ ] Docker服务运行正常
- [ ] HTTP访问成功重定向到HTTPS
- [ ] HTTPS访问正常
- [ ] API接口返回正确数据
- [ ] SSL证书有效且自动续期已配置
- [ ] 防火墙规则正确配置
- [ ] 安全密钥已修改

## 📞 获取帮助

如果遇到问题：

1. 查看服务日志：`docker-compose -f docker-compose.local.yml logs -f`
2. 检查DNS配置：`./check-nnkong-dns.sh`
3. 验证SSL证书：`curl -I https://nnkong.asiayu`
4. 测试API接口：`curl https://nnkong.asiayu/api/v1/networks/list`

---

**恭喜！你的以太坊钱包服务现在可以通过 nnkong.asiayu 域名访问了！** 🎉