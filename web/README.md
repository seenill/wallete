# Wallet Frontend

这是一个基于React + TypeScript + Vite构建的以太坊钱包前端应用。

## 功能特性

- 🔒 **安全可靠**: 助记词不会被存储到服务器
- 💰 **余额查询**: 实时查看ETH和ERC20代币余额  
- 📤 **便捷转账**: 简单快捷的ETH转账功能
- 📊 **交易历史**: 查看详细的交易记录
- 🎨 **现代化UI**: 响应式设计，支持移动端

## 技术栈

- **前端框架**: React 18 + TypeScript
- **构建工具**: Vite
- **路由**: React Router v6
- **HTTP客户端**: Axios  
- **以太坊交互**: Ethers.js v6
- **样式**: 纯CSS + CSS Variables

## 快速开始

### 1. 安装依赖

```bash
cd /Users/xaviernong/GolandProjects/wallet/web
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端将在 `http://localhost:3000` 启动

### 3. 启动后端服务

确保后端服务在 `http://localhost:8080` 运行:

```bash
cd /Users/xaviernong/GolandProjects/wallet
go run main.go
```

## 项目结构

```
src/
├── components/          # 可复用组件
│   └── Layout/         # 布局组件
├── contexts/           # React上下文
│   └── WalletContext.tsx  # 钱包状态管理
├── pages/              # 页面组件
│   ├── Home.tsx        # 首页
│   ├── Wallet.tsx      # 钱包概览
│   ├── Send.tsx        # 发送交易
│   ├── Receive.tsx     # 接收地址
│   ├── History.tsx     # 交易历史
│   └── Settings.tsx    # 设置页面
├── services/           # API服务
│   └── api.ts          # 后端API封装
├── App.tsx             # 主应用组件
└── main.tsx            # 应用入口
```

## 主要功能

### 钱包管理
- 导入助记词创建钱包
- 查看钱包地址和余额
- 断开钱包连接

### 交易功能  
- 发送ETH到指定地址
- Gas费用估算和自定义
- 交易状态跟踪

### 安全特性
- 助记词本地存储（不发送到服务器）
- 交易签名在前端完成
- 安全的私钥管理

## API集成

前端通过以下API与后端通信:

- `POST /api/v1/wallets/import-mnemonic` - 导入助记词
- `GET /api/v1/wallets/{address}/balance` - 查询余额
- `GET /api/v1/networks/gas-suggestion` - 获取Gas建议
- `POST /api/v1/transactions/send` - 发送交易

更多API详情请参考 `openapi.yaml`

## 构建部署

### 构建生产版本

```bash
npm run build
```

构建产物将生成在 `dist/` 目录

### 预览构建结果

```bash
npm run preview
```

## 开发注意事项

1. **代理配置**: Vite已配置代理，将API请求转发到后端
2. **类型安全**: 使用TypeScript确保类型安全
3. **响应式设计**: 支持桌面端和移动端
4. **错误处理**: 完善的错误处理和用户提示

## 浏览器支持

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## 许可证

MIT License