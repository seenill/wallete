# 🌐 公网部署完整检查清单

## 📋 部署前准备

### ☁️ 1. 云服务器配置
- [ ] **服务器选择**
  - [ ] 购买云服务器（推荐：2核4GB，40GB SSD）
  - [ ] 选择合适的地区（用户集中地区）
  - [ ] 配置安全组/防火墙规则
  - [ ] 记录服务器IP地址

- [ ] **操作系统配置**
  - [ ] 选择Ubuntu 20.04 LTS或CentOS 8
  - [ ] 创建非root用户
  - [ ] 配置SSH密钥登录
  - [ ] 禁用root直接登录

### 🌐 2. 域名配置
- [ ] **域名购买**
  - [ ] 购买合适的域名
  - [ ] 完成实名认证
  - [ ] 配置DNS解析

- [ ] **DNS记录配置**
  ```
  类型: A    主机: @      值: your-server-ip
  类型: A    主机: api    值: your-server-ip
  类型: A    主机: www    值: your-server-ip
  ```
  - [ ] 等待DNS生效（最多24小时）
  - [ ] 验证解析：`nslookup your-domain.com`

### 🔗 3. 区块链RPC服务
- [ ] **选择RPC提供商**
  - [ ] 注册Infura/Alchemy账号
  - [ ] 创建项目获取API Key
  - [ ] 测试RPC连接可用性
  - [ ] 配置多个备用RPC（容错）

## 🔧 部署配置

### 🐳 4. Docker环境
- [ ] **Docker安装**
  ```bash
  # 安装Docker
  curl -fsSL https://get.docker.com -o get-docker.sh
  sudo sh get-docker.sh
  
  # 安装Docker Compose
  sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
  sudo chmod +x /usr/local/bin/docker-compose
  ```
  - [ ] 验证安装：`docker --version`
  - [ ] 验证Compose：`docker-compose --version`

### 📁 5. 项目部署
- [ ] **代码部署**
  ```bash
  # 克隆项目
  git clone https://github.com/your-username/wallet.git
  cd wallet
  
  # 构建并启动
  ./deploy.sh
  ```
  - [ ] 选择Docker容器部署
  - [ ] 验证服务启动：`docker ps`

## 🔐 安全配置

### 🛡️ 6. 防火墙配置
```bash
sudo ufw allow 22      # SSH
sudo ufw allow 80      # HTTP  
sudo ufw allow 443     # HTTPS
sudo ufw deny 8087     # 禁止直接访问应用端口
sudo ufw enable
```
- [ ] 开放必要端口
- [ ] 禁止不必要端口
- [ ] 验证规则：`sudo ufw status`

### 🔑 7. 应用安全配置
- [ ] **修改默认密钥**（重要！）
  ```yaml
  # config/config.yaml
  security:
    jwt_secret: "your-production-jwt-secret-at-least-32-chars"
    encryption_key: "your-production-encryption-key-32-chars"
  ```

- [ ] **配置RPC URLs**
  ```yaml
  networks:
    ethereum:
      rpc_url: "https://mainnet.infura.io/v3/YOUR_PROJECT_ID"
    polygon:  
      rpc_url: "https://polygon-mainnet.infura.io/v3/YOUR_PROJECT_ID"
  ```

- [ ] **生产环境变量**
  ```bash
  export GIN_MODE=release
  export SERVER_PORT=8087
  ```

### 🔒 8. SSL证书配置
- [ ] **自动配置SSL**
  ```bash
  ./setup-domain.sh
  # 选择SSL证书申请选项
  ```
  
- [ ] **手动配置SSL**
  ```bash
  sudo apt install certbot python3-certbot-nginx
  sudo certbot --nginx -d your-domain.com
  ```

- [ ] **验证HTTPS访问**
  - [ ] 测试：`curl https://your-domain.com/health`
  - [ ] 浏览器访问验证证书

## 🚀 服务启动

### 9. 启动服务
```bash
# 使用生产配置启动
docker-compose -f docker-compose.prod.yml up -d

# 或使用部署脚本
./deploy.sh
```
- [ ] 检查容器状态：`docker ps`
- [ ] 查看日志：`docker-compose logs -f`

### 10. 服务验证
- [ ] **基础连通性**
  ```bash
  curl http://your-domain.com/health
  curl https://your-domain.com/health
  ```

- [ ] **API接口测试**
  ```bash
  curl https://your-domain.com/api/v1/networks/list
  ```

- [ ] **完整功能测试**
  ```javascript
  // 浏览器控制台测试
  const API_BASE = 'https://your-domain.com/api/v1';
  
  // 测试网络接口
  fetch(`${API_BASE}/networks/list`)
    .then(r => r.json())
    .then(console.log);
  ```

## 📊 监控和维护

### 11. 监控配置
- [ ] **日志配置**
  ```bash
  # 创建日志目录
  sudo mkdir -p /var/log/wallet
  sudo chown deploy:deploy /var/log/wallet
  ```

- [ ] **健康检查**
  - [ ] 配置Docker健康检查
  - [ ] 设置监控告警（可选）

### 12. 备份策略
- [ ] **自动备份**
  ```bash
  # 设置定时备份
  sudo crontab -e
  # 添加：0 2 * * * /opt/wallet/backup.sh
  ```

- [ ] **SSL证书自动续期**
  ```bash
  # 添加到crontab
  0 12 * * * /usr/bin/certbot renew --quiet
  ```

## ✅ 最终验证

### 13. 完整功能测试
- [ ] **访问地址验证**
  - [ ] `https://your-domain.com/health` ✅
  - [ ] `https://your-domain.com/api/v1/networks/list` ✅
  - [ ] `https://api.your-domain.com` ✅ (如果配置了子域名)

- [ ] **API功能测试**
  - [ ] 网络切换功能
  - [ ] 钱包导入功能
  - [ ] 余额查询功能
  - [ ] Gas估算功能

- [ ] **性能测试**
  - [ ] 响应时间 < 1秒
  - [ ] 并发请求处理
  - [ ] 错误处理机制

### 14. 文档和支持
- [ ] **用户文档**
  - [ ] API文档可访问
  - [ ] 使用示例完整
  - [ ] 错误码说明

- [ ] **运维文档**
  - [ ] 部署流程记录
  - [ ] 故障排除手册
  - [ ] 联系方式配置

## 🎯 部署成本估算

### 基础成本（月费用）
- **云服务器**: ¥45-200/月
- **域名**: ¥5-15/月（年付）
- **RPC服务**: $0-50/月（根据用量）
- **SSL证书**: 免费（Let's Encrypt）

### 总计：约 ¥50-300/月

## 📞 紧急联系

部署过程中如遇问题：

1. **检查日志**: `docker-compose logs -f`
2. **验证配置**: `./check-dns.sh`
3. **重启服务**: `docker-compose restart`
4. **查看文档**: [DEPLOYMENT.md](./DEPLOYMENT.md)

---

## 🎉 部署完成标志

当以下所有项目都完成时，你的钱包服务就成功部署到公网了：

- ✅ 通过域名可以访问服务
- ✅ HTTPS证书有效
- ✅ API接口正常响应
- ✅ 监控和备份配置完成
- ✅ 安全配置已加固

**恭喜！你的以太坊钱包服务现在可以为全世界用户提供服务了！** 🌍✨