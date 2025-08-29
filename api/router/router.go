/*
API路由配置包

本包负责配置所有的HTTP路由和中间件，组织API的结构和访问权限。

路由组织结构：
- /api/v1/auth/* - 认证相关接口（登录、注册、Token管理）
- /api/v1/wallets/* - 钱包管理接口（创建、导入、余额查询）
- /api/v1/networks/* - 多链网络管理接口（切换、状态查询）
- /api/v1/transactions/* - 交易相关接口（发送、查询、广播）
- /api/v1/tokens/* - 代币相关接口（元数据、授权管理）
- /api/v1/sign/* - 消息签名接口（Personal Sign、EIP-712）
- /api/v1/defi/* - DeFi相关接口（1inch集成、流动性、收益等）
- /health - 服务健康检查接口

中间件应用：
- 全局中间件：错误处理、安全头、请求ID、速率限制
- 认证中间件：JWT认证、API密钥认证、可选认证
- 业务中间件：交易验证、特殊速率限制

安全特性：
- 分层的速率限制策略
- JWT和API密钥双重认证机制
- 交易操作的额外安全验证
- 超时和请求频率控制
*/
package router

import (
	"net/http"
	"wallet/api/handlers"
	"wallet/api/middleware"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置新的Gin HTTP路由器
func NewRouter(walletService *services.WalletService) *gin.Engine {
	// 创建默认Gin引擎（包含Logger和Recovery中间件）
	r := gin.Default()

	// 应用全局中间件（按顺序执行）
	r.Use(middleware.CORS())            // CORS跨域支持
	r.Use(middleware.ErrorHandler())    // 统一错误处理
	r.Use(middleware.SecurityHeaders()) // HTTP安全头设置
	r.Use(middleware.RequestID())       // 请求追踪ID生成
	r.Use(middleware.RateLimit())       // 通用速率限制
	// 可以添加更多中间件，例如日志、CORS等

	// 创建各个业务处理器实例
	walletHandler := handlers.NewWalletHandler(walletService) // 钱包相关操作处理器
	// 移除传统的认证处理器
	//authHandler := handlers.NewAuthHandler()                                                             // 认证相关操作处理器
	// 创建助记词认证处理器
	mnemonicAuthHandler := handlers.NewMnemonicAuthHandler(walletService)                                // 助记词认证处理器
	watchAddressHandler := handlers.NewWatchAddressHandler()                                             // 观察地址管理处理器
	userWalletHandler := handlers.NewUserWalletHandler()                                                 // 用户钱包记录处理器
	networkHandler := handlers.NewNetworkHandler(walletService.GetMultiChainManager(), walletService)    // 网络相关操作处理器
	defiHandler := handlers.NewDeFiHandler(walletService.GetDeFiService())                               // DeFi功能处理器
	nftHandler := handlers.NewNFTHandler(walletService.GetNFTService())                                  // NFT功能处理器
	dappBrowserHandler := handlers.NewDAppBrowserHandler(walletService.GetDAppBrowserService())          // DApp浏览器处理器
	socialHandler := handlers.NewSocialHandler(walletService.GetSocialService())                         // 社交功能处理器
	securityHandler := handlers.NewSecurityHandler(walletService.GetSecurityService())                   // 安全功能处理器
	nftMarketplaceHandler := handlers.NewNFTMarketplaceHandler(walletService.GetNFTMarketplaceService()) // NFT市场处理器
	// 创建1inch处理器
	oneInchHandler := handlers.NewOneInchHandler(walletService.GetDeFiService()) // 1inch聚合器处理器

	// 助记词认证相关路由组（替代传统的注册登录）
	// 包括钱包创建、助记词认证、会话管理等功能
	auth := r.Group("/api/v1/auth")
	{
		// 为认证接口应用特殊的速率限制（防暴力破解）
		auth.Use(middleware.AuthRateLimit())

		// 公开接口（无需认证）
		auth.POST("/mnemonic/auth", mnemonicAuthHandler.AuthenticateWithMnemonic) // 助记词认证
		auth.POST("/mnemonic/create", mnemonicAuthHandler.CreateWallet)           // 创建新钱包
	}

	// API v1 主路由组（需要用户认证）
	// 使用JWTAuth中间件，确保所有接口都需要有效的JWT令牌
	v1 := r.Group("/api/v1")
	v1.Use(middleware.JWTAuth()) // 统一的JWT认证机制
	{
		// 添加会话注销接口
		auth.POST("/logout", mnemonicAuthHandler.Logout) // 会话注销

		// 观察地址管理相关路由组
		// 提供用户观察地址的增删改查功能
		watchAddressGroup := v1.Group("/watch-addresses")
		{
			watchAddressGroup.POST("", watchAddressHandler.AddWatchAddress)          // 添加观察地址
			watchAddressGroup.GET("", watchAddressHandler.GetWatchAddresses)         // 获取观察地址列表
			watchAddressGroup.GET("/:id", watchAddressHandler.GetWatchAddress)       // 获取单个观察地址详情
			watchAddressGroup.PUT("/:id", watchAddressHandler.UpdateWatchAddress)    // 更新观察地址
			watchAddressGroup.DELETE("/:id", watchAddressHandler.DeleteWatchAddress) // 删除观察地址
		}

		// 用户钱包记录管理相关路由组
		// 管理用户导入/创建的钱包记录
		userWalletGroup := v1.Group("/user-wallets")
		{
			userWalletGroup.POST("", userWalletHandler.AddUserWallet)                    // 添加钱包记录
			userWalletGroup.GET("", userWalletHandler.GetUserWallets)                    // 获取钱包记录列表
			userWalletGroup.GET("/:id", userWalletHandler.GetUserWallet)                 // 获取单个钱包记录详情
			userWalletGroup.PUT("/:id", userWalletHandler.UpdateUserWallet)              // 更新钱包记录
			userWalletGroup.DELETE("/:id", userWalletHandler.DeleteUserWallet)           // 删除钱包记录
			userWalletGroup.POST("/:id/set-primary", userWalletHandler.SetPrimaryWallet) // 设置主钱包
		}
		// 钱包管理相关路由组（支持HD钱包功能）
		// 包括钱包创建、导入、余额查询和交易历史等核心功能
		// 使用可选认证，兼容现有功能
		walletGroup := r.Group("/api/v1/wallets")
		walletGroup.Use(middleware.OptionalAuth()) // 灵活的认证机制
		{
			walletGroup.POST("/new", walletHandler.CreateWallet)                                     // 创建新钱包（生成助记词）
			walletGroup.POST("/import-mnemonic", walletHandler.ImportMnemonic)                       // 通过助记词导入钱包
			walletGroup.GET("/:address/balance", walletHandler.GetBalance)                           // 获取原生代币余额（ETH/MATIC/BNB）
			walletGroup.GET("/:address/tokens/:tokenAddress/balance", walletHandler.GetERC20Balance) // 获取ERC20代币余额
			walletGroup.GET("/:address/nonce", walletHandler.GetNonces)                              // 获取地址的nonce值
			walletGroup.GET("/:address/history", walletHandler.GetTransactionHistory)                // 查询交易历史（支持分页和过滤）
		}

		// 多链网络管理路由组
		// 支持动态网络切换、状态查询和跨链操作
		networkGroup := r.Group("/api/v1/networks")
		// 注意：网络列表和当前网络信息不需要认证，但其他操作需要认证
		{
			networkGroup.GET("", networkHandler.ListNetworks)              // 获取所有可用网络
			networkGroup.GET("/current", networkHandler.GetCurrentNetwork) // 获取当前活跃网络信息
			networkGroup.GET("/list", networkHandler.ListNetworks)         // 列出所有可用网络
			networkGroup.GET("/:networkId", networkHandler.GetNetworkInfo) // 获取特定网络详细信息
		}

		// 需要认证的网络操作
		networkGroupAuth := v1.Group("/networks")
		networkGroupAuth.Use(middleware.OptionalAuth()) // 灵活的认证机制
		{
			networkGroupAuth.GET("/addresses/:address/balance", networkHandler.GetBalanceOnNetwork)                                                    // 获取指定网络上的余额
			networkGroupAuth.GET("/addresses/:address/cross-chain-balance", networkHandler.GetCrossChainBalance)                                       // 跨链余额查询（聚合所有网络）
			networkGroupAuth.GET("/addresses/:address/tokens/:tokenAddress/cross-chain-balance", networkHandler.GetCrossChainTokenBalance)             // 跨链代币余额查询
			networkGroupAuth.POST("/send-eth", middleware.TransactionRateLimit(), middleware.TransactionValidation(), networkHandler.SendETHOnNetwork) // 在指定网络发送ETH
			networkGroupAuth.POST("/switch", networkHandler.SwitchNetwork)                                                                             // 切换到指定网络
		}

		// Gas价格建议接口（全局可用）
		gasGroup := r.Group("/api/v1")
		gasGroup.Use(middleware.OptionalAuth())
		{
			gasGroup.GET("/gas-suggestion", walletHandler.GetGasSuggestion) // 获取当前网络的Gas价格建议
		}

		// DeFi功能相关路由组
		// 提供去中心化金融服务，包括DEX交易、流动性挖矿、收益农场等
		defiGroup := v1.Group("/defi")
		{
			// DEX交易聚合相关接口
			swapGroup := defiGroup.Group("/swap")
			{
				swapGroup.GET("/quote", defiHandler.GetSwapQuote)                                      // 获取最佳交易报价
				swapGroup.POST("/execute", middleware.TransactionRateLimit(), defiHandler.ExecuteSwap) // 执行Swap交易
			}

			// 1inch聚合器相关接口
			oneInchGroup := defiGroup.Group("/oneinch")
			{
				oneInchGroup.GET("/quote", oneInchHandler.GetQuote)                        // 获取1inch报价
				oneInchGroup.GET("/swap", oneInchHandler.GetSwap)                          // 获取1inch交换数据
				oneInchGroup.GET("/tokens", oneInchHandler.GetTokens)                      // 获取代币列表
				oneInchGroup.GET("/liquidity-sources", oneInchHandler.GetLiquiditySources) // 获取流动性源
			}

			// 流动性管理相关接口
			liquidityGroup := defiGroup.Group("/liquidity")
			{
				liquidityGroup.GET("/pools", defiHandler.GetLiquidityPools)                              // 获取流动性池列表
				liquidityGroup.POST("/add", middleware.TransactionRateLimit(), defiHandler.AddLiquidity) // 添加流动性
			}

			// 收益农场相关接口
			yieldGroup := defiGroup.Group("/yield")
			{
				yieldGroup.GET("/strategies", defiHandler.GetYieldStrategies) // 获取收益策略列表
			}

			// 价格查询相关接口
			priceGroup := defiGroup.Group("/price")
			{
				priceGroup.GET("/tokens", defiHandler.GetTokenPrices) // 获取代币价格信息
			}
		}

		// NFT功能相关路由组
		// 提供NFT相关服务，包括查询、转账、市场数据等
		nftGroup := v1.Group("/nft")
		{
			// 用户NFT相关接口
			userGroup := nftGroup.Group("/user")
			{
				userGroup.GET("/:address/nfts", nftHandler.GetUserNFTs) // 获取用户NFT列表
			}

			// NFT详情相关接口
			detailsGroup := nftGroup.Group("/details")
			{
				detailsGroup.GET("/:contract/:tokenId", nftHandler.GetNFTDetails) // 获取NFT详情
			}

			// NFT搜索相关接口
			nftGroup.GET("/search", nftHandler.SearchNFTs) // 搜索NFT

			// NFT活动相关接口
			nftGroup.GET("/activities", nftHandler.GetNFTActivities) // 获取NFT活动记录

			// NFT估值相关接口
			nftGroup.POST("/estimate-value", nftHandler.EstimateNFTValue) // 估算NFT价值

			// NFT集合相关接口
			collectionGroup := nftGroup.Group("/collections")
			{
				collectionGroup.GET("/:address", nftHandler.GetCollectionInfo) // 获取NFT集合信息
			}

			// NFT市场相关接口
			marketGroup := nftGroup.Group("/market")
			{
				marketGroup.GET("/hot-collections", nftHandler.GetHotCollections) // 获取热门NFT集合
				marketGroup.GET("/trends", nftHandler.GetMarketTrends)            // 获取市场趋势
			}

			// NFT转账相关接口
			nftGroup.POST("/transfer", middleware.TransactionRateLimit(), nftHandler.TransferNFT) // 转移NFT

			// NFT投资组合相关接口
			portfolioGroup := nftGroup.Group("/portfolio")
			{
				portfolioGroup.GET("/:address", nftHandler.GetUserPortfolio) // 获取用户NFT投资组合
			}
		}

		// NFT市场相关路由组
		// 提供NFT市场功能，包括交易、列表、统计等
		marketplaceGroup := v1.Group("/nft/marketplace")
		{
			marketplaceGroup.GET("/listings", nftMarketplaceHandler.GetMarketListings)         // 获取市场列表
			marketplaceGroup.GET("/transactions", nftMarketplaceHandler.GetMarketTransactions) // 获取市场交易记录
			marketplaceGroup.GET("/stats/:contract", nftMarketplaceHandler.GetMarketStats)     // 获取市场统计数据
			marketplaceGroup.POST("/analyze", nftMarketplaceHandler.AnalyzeMarket)             // 分析市场数据
			marketplaceGroup.GET("/preferences", nftMarketplaceHandler.GetUserPreferences)     // 获取用户偏好设置
			marketplaceGroup.POST("/preferences", nftMarketplaceHandler.SetUserPreferences)    // 设置用户偏好设置
			marketplaceGroup.POST("/watchlist", nftMarketplaceHandler.AddToWatchlist)          // 添加到关注列表
			marketplaceGroup.GET("/watchlist/:listName", nftMarketplaceHandler.GetWatchlist)   // 获取关注列表
			marketplaceGroup.POST("/price-alert", nftMarketplaceHandler.CreatePriceAlert)      // 创建价格提醒
			marketplaceGroup.GET("/price-alerts", nftMarketplaceHandler.GetPriceAlerts)        // 获取价格提醒列表
		}

		// DApp浏览器相关路由组
		// 提供DApp连接和交互功能
		dappGroup := v1.Group("/dapp")
		{
			dappGroup.POST("/connect", dappBrowserHandler.ConnectDApp)                                                // 连接DApp
			dappGroup.GET("/connect/:sessionId", dappBrowserHandler.GetSessionInfo)                                   // 获取会话信息
			dappGroup.DELETE("/connect/:sessionId", dappBrowserHandler.DisconnectDApp)                                // 断开DApp连接
			dappGroup.POST("/web3/request", dappBrowserHandler.ProcessWeb3Request)                                    // 处理Web3请求
			dappGroup.POST("/web3/confirm", middleware.TransactionRateLimit(), dappBrowserHandler.ConfirmWeb3Request) // 确认Web3请求
			dappGroup.GET("/web3/pending/:address", dappBrowserHandler.GetPendingRequests)                            // 获取待处理请求
			dappGroup.GET("/discovery/list", dappBrowserHandler.GetDAppList)                                          // 获取DApp列表
			dappGroup.GET("/discovery/featured", dappBrowserHandler.GetFeaturedDApps)                                 // 获取推荐DApp
			dappGroup.GET("/discovery/search", dappBrowserHandler.SearchDApps)                                        // 搜索DApp
			dappGroup.GET("/discovery/categories", dappBrowserHandler.GetCategories)                                  // 获取DApp分类
			dappGroup.GET("/user/:address/activity", dappBrowserHandler.GetUserActivity)                              // 获取用户活动记录
			dappGroup.POST("/user/favorite", dappBrowserHandler.ManageFavorite)                                       // 管理收藏DApp
		}

		// 社交功能相关路由组
		// 提供联系人管理、交易分享等社交功能
		socialGroup := v1.Group("/social")
		{
			socialGroup.GET("/contacts", socialHandler.GetContactList)                    // 获取联系人列表
			socialGroup.POST("/contacts", socialHandler.AddContact)                       // 添加联系人
			socialGroup.GET("/contacts/:contactId", socialHandler.GetContact)             // 获取联系人详情
			socialGroup.PUT("/contacts/:contactId", socialHandler.UpdateContact)          // 更新联系人
			socialGroup.DELETE("/contacts/:contactId", socialHandler.DeleteContact)       // 删除联系人
			socialGroup.POST("/share/transaction", socialHandler.ShareTransaction)        // 分享交易
			socialGroup.GET("/share/my", socialHandler.GetMyShares)                       // 获取我的分享
			socialGroup.GET("/share/:shareId", socialHandler.GetShareRecord)              // 获取分享记录
			socialGroup.POST("/network/action", socialHandler.SocialNetworkAction)        // 社交网络操作
			socialGroup.GET("/network/:address/followers", socialHandler.GetFollowers)    // 获取关注者列表
			socialGroup.GET("/network/:address/following", socialHandler.GetFollowing)    // 获取关注列表
			socialGroup.GET("/user/:address/profile", socialHandler.GetUserSocialProfile) // 获取用户社交资料
			socialGroup.PUT("/user/profile", socialHandler.UpdateUserSocialProfile)       // 更新用户社交资料
			socialGroup.GET("/search/users", socialHandler.SearchUsers)                   // 搜索用户
		}

		// 安全功能相关路由组
		// 提供硬件钱包检测、多签钱包、MFA等安全功能
		securityGroup := v1.Group("/security")
		{
			securityGroup.GET("/hardware/detect", securityHandler.DetectHardwareWallets)                                                 // 检测硬件钱包
			securityGroup.POST("/hardware/request", securityHandler.ProcessHardwareWalletRequest)                                        // 处理硬件钱包请求
			securityGroup.POST("/multisig/create", securityHandler.CreateMultiSigWallet)                                                 // 创建多签钱包
			securityGroup.POST("/multisig/transaction/create", securityHandler.CreateMultiSigTransaction)                                // 创建多签交易
			securityGroup.POST("/multisig/transaction/sign", middleware.TransactionRateLimit(), securityHandler.SignMultiSigTransaction) // 签名多签交易
			securityGroup.POST("/mfa/setup", securityHandler.SetupMFA)                                                                   // 设置MFA
			securityGroup.POST("/mfa/verify", securityHandler.VerifyMFA)                                                                 // 验证MFA
			securityGroup.GET("/audit/logs", securityHandler.GetSecurityAuditLogs)                                                       // 获取安全审计日志
			securityGroup.GET("/audit/report/:address", securityHandler.GetSecurityReport)                                               // 获取安全报告
			securityGroup.POST("/biometric/enable", securityHandler.EnableBiometric)                                                     // 启用生物识别
			securityGroup.POST("/biometric/verify", securityHandler.VerifyBiometric)                                                     // 验证生物识别
			securityGroup.GET("/status/:address", securityHandler.GetSecurityStatus)                                                     // 获取安全状态
		}

		// 交易相关路由组
		// 提供交易发送、估算、广播等核心功能
		transactionGroup := v1.Group("/transactions")
		transactionGroup.Use(middleware.TransactionRateLimit())  // 交易专用速率限制
		transactionGroup.Use(middleware.TransactionValidation()) // 交易验证中间件
		{
			transactionGroup.POST("/send", walletHandler.SendTransaction)                  // 发送交易
			transactionGroup.POST("/send-erc20", walletHandler.SendERC20)                  // 发送ERC20代币
			transactionGroup.POST("/send-advanced", walletHandler.SendTransactionAdvanced) // 发送高级交易
			transactionGroup.POST("/send-erc20-advanced", walletHandler.SendERC20Advanced) // 发送高级ERC20交易
			transactionGroup.POST("/estimate", walletHandler.EstimateTransaction)          // 估算交易
			transactionGroup.POST("/broadcast", walletHandler.BroadcastRawTransaction)     // 广播原始交易
			transactionGroup.GET("/:hash/receipt", walletHandler.GetTxReceipt)             // 获取交易回执
		}

		// 代币相关路由组
		// 提供代币元数据、授权管理等功能
		tokenGroup := v1.Group("/tokens")
		{
			tokenGroup.GET("/:token/metadata", walletHandler.GetTokenMetadata) // 获取代币元数据
			tokenGroup.POST("/:token/approve", walletHandler.ApproveToken)     // 授权代币
			tokenGroup.GET("/:token/allowance", walletHandler.GetAllowance)    // 获取授权额度
		}

		// 消息签名相关路由组
		// 提供个人签名和EIP-712签名功能
		signGroup := v1.Group("/sign")
		{
			signGroup.POST("/message", walletHandler.PersonalSign)  // 个人消息签名
			signGroup.POST("/typed", walletHandler.SignTypedDataV4) // EIP-712签名
		}
	}

	// 健康检查接口（无需认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Wallet service is running",
		})
	})

	return r
}
