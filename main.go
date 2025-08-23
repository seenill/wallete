/*
区块链钱包服务应用程序入口文件

本文件是区块链钱包服务的主启动程序，负责：
1. 加载应用配置
2. 初始化认证和安全中间件
3. 初始化钱包服务
4. 配置路由并启动HTTP服务器

支持的功能：
- HD钱包创建和管理
- 多链支持（以太坊、Polygon、BSC等）
- 交易历史查询
- JWT认证和API密钥管理
- 速率限制和安全防护
- ERC20代币支持
*/
package main

import (
	"fmt"
	"wallet/api/middleware"
	"wallet/api/router"
	"wallet/config"
	"wallet/services"
)

// main 应用程序主入口函数
// 按顺序初始化各个组件并启动HTTP服务器
func main() {
	// 1. 加载配置文件和环境变量
	// 从config.yaml加载服务器、数据库、网络等配置
	config.LoadConfig()

	// 2. 初始化认证和安全中间件
	// 设置JWT密钥和速率限制器，为API接口提供安全保护
	middleware.InitAuth(config.AppConfig.Security.JWTSecret)
	middleware.InitRateLimiters()

	// 3. 初始化钱包服务，并注入到路由（无需数据库）
	// 创建钱包服务实例，包含多链管理器和加密管理器
	walletService := services.NewWalletService()
	r := router.NewRouter(walletService)

	// 4. 启动HTTP服务器
	// 在配置的端口上启动Gin HTTP服务器
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	fmt.Printf("Server is running at %s\n", addr)
	if err := r.Run(addr); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}
