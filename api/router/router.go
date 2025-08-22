package router

import (
	"github.com/gin-gonic/gin"
	"wallet/api/handlers"
	"wallet/api/middleware"
	"wallet/services"
)

// NewRouter 创建并配置一个新的 Gin 路由器
func NewRouter(walletService *services.WalletService) *gin.Engine {
	r := gin.Default()

	// 使用全局中间件
	r.Use(middleware.ErrorHandler())
	// 可以添加更多中间件，例如日志、CORS等

	// 创建处理器实例
	walletHandler := handlers.NewWalletHandler(walletService)

	// API V1 路由组
	v1 := r.Group("/api/v1")
	{
		// 钱包相关路由
		walletGroup := v1.Group("/wallets")
		{
			walletGroup.POST("/", walletHandler.CreateWallet)           // 创建钱包
			walletGroup.GET("/:address", walletHandler.GetWalletInfo)   // 获取钱包信息
		}

		// 交易相关路由（预留）
		txGroup := v1.Group("/transactions")
		{
			// txGroup.POST("/send", handlers.SendTransaction)
			_ = txGroup // 暂时避免未使用变量的警告
		}
	}

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "钱包服务运行正常",
		})
	})

	return r
}