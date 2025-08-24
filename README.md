# 🦄 Ethereum Wallet - 以太坊钱包

一个基于 Go + React 构建的完整以太坊钱包解决方案，提供安全、高效的数字资产管理服务。

## 🎆 项目亮点

- **🔒 安全可靠**: 助记词不存储，私钥本地管理
- **💰 全面功能**: 支持ETH/ERC20转账、余额查询、交易历史
- **🎨 现代化UI**: 响应式设计，支持桌面端和移动端
- **⚡ 高性能**: Go后端 + React前端，快速响应
- **🔧 开发友好**: 完整的API文档和类型定义

## 🏠 项目架构

```
wallet/
├── api/                    # Go 后端 API
│   ├── handlers/          # HTTP 请求处理器
│   ├── middleware/        # 中间件
│   └── router/            # 路由配置
├── core/                   # 核心业务逻辑
├── services/               # 服务层
├── web/                    # React 前端
│   ├── src/
│   │   ├── components/    # 组件
│   │   ├── contexts/      # 状态管理
│   │   ├── pages/         # 页面
│   │   └── services/      # API封装
│   └── README.md           # 前端文档
├── config/                 # 配置文件
├── main.go                 # 后端入口
├── start.sh                # 快速启动脚本
└── openapi.yaml            # API 文档
```

## ✨ 功能特性

### 钱包管理
- **助记词导入**: 通过助记词派生并返回钱包地址，但不存储助记词以确保安全。
- **余额查询**:
  - 查询指定地址的 ETH 余额。
  - 查询指定地址的 ERC20 代币余额。
- **Nonce 查询**: 获取地址的 `latest` 和 `pending` nonce。

### 交易处理
- **ETH 转账**:
  - **基础发送**: 快速发送 ETH。
  - **高级发送**: 自定义 `gas` 和 `nonce` 进行发送。
- **ERC20 转账**:
  - **基础发送**: 快速发送 ERC20 代币。
  - **高级发送**: 自定义 `gas` 和 `nonce` 进行发送。
- **Gas 估算**: 估算交易所需的 `gas limit`。
- **Gas 建议**: 获取 EIP-1559 和 legacy 两种模式下的 `gas` 价格建议。
- **交易广播**: 广播已签名的原始交易。
- **交易回执**: 查询交易回执，并在交易失败时尝试返回 revert reason。

### ERC20 代币
- **元数据查询**: 获取 ERC20 代币的名称、符号和小数位数。
- **授权 (Approve)**: 对 `spender` 地址进行 ERC20 代币授权。
- **授权额度查询 (Allowance)**: 查询 `owner` 对 `spender` 的授权额度。

### 签名服务
- **文本消息签名**: 使用 `personal_sign` 标准对文本消息进行签名。
- **EIP-712 签名**: 对符合 EIP-712 标准的 typed data 进行签名。

## 🏛️ 项目架构

项目采用分层架构，确保代码的模块化和可维护性。

- **/api**: API 层，负责处理 HTTP 请求和响应。
  - **/handlers**: 包含所有 HTTP `handler`，负责解析请求、调用 `service` 并返回响应。
  - **/router**: 定义所有 API 路由，并将它们与 `handler` 关联。
  - **/middleware**: 包含中间件，如错误处理、日志记录等。
- **/core**: 核心逻辑层，封装了与以太坊节点的交互。
  - **evm_adapter.go**: 封装了与以太坊 JSON-RPC 的所有交互，例如查询余额、发送交易等。
  - **hd.go**: 实现了 HD 钱包的助记词派生逻辑。
- **/services**: 服务层，封装了核心业务逻辑。
  - **wallet_service.go**: 组合 `core` 层的功能，为 `api` 层提供统一的业务接口。
- **/config**: 配置管理。
- **/pkg**: 工具包，提供项目范围内的通用功能。
- **main.go**: 项目入口文件。
- **openapi.yaml**: API 规范文档。
## 🚀 快速开始

### 本地开发

#### 使用一键启动脚本

```bash
# 克隆项目
git clone <repository-url>
cd wallet

# 一键启动（仅后端）
./start.sh
```

### 生产部署

🌐 **想让别人通过互联网访问你的钱包服务？**

- **5分钟快速部署**: 查看 [`QUICK_DEPLOY.md`](./QUICK_DEPLOY.md)
- **完整部署指南**: 查看 [`DEPLOYMENT.md`](./DEPLOYMENT.md)
- **一键部署脚本**: 运行 `./deploy.sh`

```bash
# 快速部署到云服务器
./deploy.sh
# 选择相应的部署选项即可
```

### 手动启动

#### 1. 启动后端服务

```bash
# 安装依赖
go mod tidy

# 启动服务
go run main.go
```

后端服务将在 `http://localhost:8080` 运行

#### 2. 启动前端应用

```bash
# 进入前端目录
cd web

# 安装依赖
npm install
# 或使用 yarn
yarn install

# 启动开发服务器
npm run dev
# 或使用 yarn
yarn dev
```

前端应用将在 `http://localhost:3000` 运行

### 访问应用

- **前端界面**: http://localhost:3000
- **后端 API**: http://localhost:8080
- **API 文档**: 导入 `openapi.yaml` 到 Swagger Editor

## 🌍 部署到生产环境

如果你希望将这个钱包服务部署到互联网上让其他人访问，我们提供了多种部署方案：

### 📦 部署文档

- **🚀 [快速部署指南](./QUICK_DEPLOY.md)** - 5分钟上线，适合新手
- **📚 [完整部署指南](./DEPLOYMENT.md)** - 详细部署文档，包含安全配置

### 🌐 部署方式对比

| 部署方式 | 难度 | 适合场景 | 成本 |
|------------|------|----------|------|
| **云服务器** | ⭐⭐ | 小型项目、快速上线 | ¥45-200/月 |
| **Docker容器** | ⭐⭐⭐ | 中型项目、需要扩展 | ¥45-200/月 |
| **Serverless** | ⭐⭐⭐ | 不定期使用 | 按量付费 |

### 🛠️ 一键部署

使用提供的部署脚本：

```bash
# 运行部署脚本
./deploy.sh

# 菜单选项：
# 1) 本地直接部署
# 2) Docker容器部署 (推荐)
# 3) 生产环境部署(含SSL)
# 4) 云服务器初始化
```

### 🎆 部署成果

部署完成后，你的服务将可以通过以下地址访问：

- **API服务**: `https://your-domain.com/api/v1`
- **健康检查**: `https://your-domain.com/health`
- **网络信息**: `https://your-domain.com/api/v1/networks`

**客户端访问示例**：
```javascript
// 访问你的API服务
const response = await fetch('https://your-domain.com/api/v1/networks');
const networks = await response.json();
console.log('支持的网络:', networks.data);
```

## 未来功能与改进

为了使钱包服务更加完善，可以考虑增加以下功能：

### ⛓️ 多链支持
- **动态网络切换**: 支持在运行时切换不同的 EVM 兼容链（如 Polygon, BSC, Arbitrum）。
- **链配置管理**: 通过配置文件或 API 管理多条链的 RPC 端点和 `chain_id`。

### 🖼️ NFT (非同质化代币)
- **资产查询**: 查询并展示用户持有的 ERC-721 和 ERC-1155 代币。
- **NFT 转账**: 发送 NFT 代币。
- **元数据获取**: 获取 NFT 的元数据和图像。

### 🔐 安全与钱包管理
- **助记词加密存储**: 提供选项让用户加密并安全地存储助记词。
- **多账户管理**: 支持从单个助记词派生和管理多个地址。
- **地址簿**: 管理常用联系人地址。

### 📈 DeFi 与高级功能
- **交易历史**: 提供 API 来查询和过滤用户的交易历史。
- **代币价格**: 集成价格预言机（如 Chainlink）或第三方 API 来显示代币的法币价值。
- **Swap 集成**: 集成去中心化交易所（DEX）的路由协议，实现代币兑换功能。

### ⚙️ 开发者体验
- **WebSocket 支持**: 通过 WebSocket 提供实时的链上事件通知（如新区块、交易确认）。
- **插件系统**: 设计一个插件系统，允许开发者轻松扩展新功能。