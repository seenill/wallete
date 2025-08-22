package main

import (
	"fmt"
	"wallet/api/router"
	"wallet/config"
	"wallet/storage"
)

func main() {
	// 1. 加载配置
	config.LoadConfig()

	// 2. 初始化数据库
	storage.InitDB()

	// 3. 设置路由
	r := router.NewRouter()

	// 4. 启动服务器
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	fmt.Printf("Server is running at %s\n", addr)
	if err := r.Run(addr); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}