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
	"wallet/api/handlers"
	"wallet/api/middleware"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建并配置新的Gin HTTP路由器
// 参数: walletService - 钱包服务实例，用于处理业务逻辑
// 返回: 完整配置的Gin引擎实例，包含所有路由和中间件
// 功能: 自动注册所有API端点、中间件和安全策略
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
	walletHandler := handlers.NewWalletHandler(walletService)                                            // 钱包相关操作处理器
	authHandler := handlers.NewAuthHandler()                                                             // 认证相关操作处理器
	networkHandler := handlers.NewNetworkHandler(walletService)                                          // 网络相关操作处理器
	defiHandler := handlers.NewDeFiHandler(walletService.GetDeFiService())                               // DeFi功能处理器
	nftHandler := handlers.NewNFTHandler(walletService.GetNFTService())                                  // NFT功能处理器
	dappBrowserHandler := handlers.NewDAppBrowserHandler(walletService.GetDAppBrowserService())          // DApp浏览器处理器
	socialHandler := handlers.NewSocialHandler(walletService.GetSocialService())                         // 社交功能处理器
	securityHandler := handlers.NewSecurityHandler(walletService.GetSecurityService())                   // 安全功能处理器
	nftMarketplaceHandler := handlers.NewNFTMarketplaceHandler(walletService.GetNFTMarketplaceService()) // NFT市场处理器

	// 认证相关路由组（公开接口，无需预先认证）
	// 包括用户登录、Token管理、API密钥生成等功能
	auth := r.Group("/api/v1/auth")
	{
		// 为认证接口应用特殊的速率限制（防暴力破解）
		auth.Use(middleware.AuthRateLimit())

		// 公开接口（无需认证）
		auth.POST("/login", authHandler.Login) // 用户登录，获取JWT Token

		// 需要JWT认证的接口
		auth.POST("/api-key", middleware.JWTAuth(), authHandler.GenerateAPIKey) // 生成API密钥
		auth.POST("/refresh", middleware.JWTAuth(), authHandler.RefreshToken)   // 刷新JWT Token
		auth.GET("/profile", middleware.JWTAuth(), authHandler.GetProfile)      // 获取用户资料
	}

	// API v1 主路由组（支持可选认证）
	// 使用OptionalAuth中间件，允许既可以通过JWT认证也可以通过API密钥认证
	v1 := r.Group("/api/v1")
	v1.Use(middleware.OptionalAuth()) // 灵活的认证机制，支持多种认证方式
	{
		// 钱包管理相关路由组（支持HD钱包功能）
		// 包括钱包创建、导入、余额查询和交易历史等核心功能
		walletGroup := v1.Group("/wallets")
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
		networkGroup := v1.Group("/networks")
		{
			networkGroup.GET("/current", networkHandler.GetCurrentNetwork)                                                                         // 获取当前活跃网络信息
			networkGroup.POST("/switch", networkHandler.SwitchNetwork)                                                                             // 切换到指定网络
			networkGroup.GET("/list", networkHandler.ListNetworks)                                                                                 // 列出所有可用网络
			networkGroup.GET("/:networkId", networkHandler.GetNetworkInfo)                                                                         // 获取特定网络详细信息
			networkGroup.GET("/:networkId/addresses/:address/balance", networkHandler.GetBalanceOnNetwork)                                         // 获取指定网络上的余额
			networkGroup.GET("/addresses/:address/cross-chain-balance", networkHandler.GetCrossChainBalance)                                       // 跨链余额查询（聚合所有网络）
			networkGroup.GET("/addresses/:address/tokens/:tokenAddress/cross-chain-balance", networkHandler.GetCrossChainTokenBalance)             // 跨链代币余额查询
			networkGroup.POST("/send-eth", middleware.TransactionRateLimit(), middleware.TransactionValidation(), networkHandler.SendETHOnNetwork) // 在指定网络发送ETH
		}

		// Gas价格建议接口（全局可用）
		v1.GET("/gas-suggestion", walletHandler.GetGasSuggestion) // 获取当前网络的Gas价格建议

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

			// NFT详情和操作接口
			nftGroup.GET("/details/:contract/:tokenId", nftHandler.GetNFTDetails) // 获取NFT详细信息
			nftGroup.GET("/search", nftHandler.SearchNFTs)                        // 搜索NFT
			nftGroup.GET("/activities", nftHandler.GetNFTActivities)              // 获取NFT活动记录
			nftGroup.POST("/estimate-value", nftHandler.EstimateNFTValue)         // 估算NFT价值

			// NFT集合相关接口
			collectionGroup := nftGroup.Group("/collections")
			{
				collectionGroup.GET("/:address", nftHandler.GetCollectionInfo) // 获取集合信息
			}

			// NFT市场数据接口
			marketGroup := nftGroup.Group("/market")
			{
				marketGroup.GET("/hot-collections", nftHandler.GetHotCollections) // 获取热门集合
				marketGroup.GET("/trends", nftHandler.GetMarketTrends)            // 获取市场趋势
			}

			// NFT转账相关接口（应用额外安全中间件）
			transferGroup := nftGroup.Group("/transfer")
			transferGroup.Use(middleware.TransactionRateLimit())  // 交易频率限制
			transferGroup.Use(middleware.TransactionValidation()) // 交易参数验证
			{
				transferGroup.POST("", nftHandler.TransferNFT) // 转NFT
			}

			// NFT投资组合接口
			portfolioGroup := nftGroup.Group("/portfolio")
			{
				portfolioGroup.GET("/:address", nftHandler.GetUserPortfolio) // 获取用户投资组合
			}

			// NFT市场相关接口
			marketplaceGroup := nftGroup.Group("/marketplace")
			{
				// 市场数据接口
				marketplaceGroup.GET("/listings", nftMarketplaceHandler.GetMarketListings)         // 获取市场挂单
				marketplaceGroup.GET("/transactions", nftMarketplaceHandler.GetMarketTransactions) // 获取交易记录
				marketplaceGroup.GET("/stats/:contract", nftMarketplaceHandler.GetMarketStats)     // 获取市场统计
				marketplaceGroup.POST("/analyze", nftMarketplaceHandler.AnalyzeMarket)             // 市场分析

				// 用户偏好接口
				marketplaceGroup.GET("/preferences", nftMarketplaceHandler.GetUserPreferences)  // 获取用户偏好
				marketplaceGroup.POST("/preferences", nftMarketplaceHandler.SetUserPreferences) // 设置用户偏好

				// 关注列表接口
				marketplaceGroup.POST("/watchlist", nftMarketplaceHandler.AddToWatchlist)        // 添加关注
				marketplaceGroup.GET("/watchlist/:listName", nftMarketplaceHandler.GetWatchlist) // 获取关注列表

				// 价格提醒接口
				marketplaceGroup.POST("/price-alert", nftMarketplaceHandler.CreatePriceAlert) // 创建价格提醒
				marketplaceGroup.GET("/price-alerts", nftMarketplaceHandler.GetPriceAlerts)   // 获取价格提醒
			}
		}

		// DApp浏览器功能相关路由组
		// 提供Web3应用集成、会话管理、安全控制等服务
		dappGroup := v1.Group("/dapp")
		{
			// DApp连接管理接口
			connectGroup := dappGroup.Group("/connect")
			{
				connectGroup.POST("", dappBrowserHandler.ConnectDApp)                 // 连接DApp应用
				connectGroup.GET("/:sessionId", dappBrowserHandler.GetSessionInfo)    // 获取会话信息
				connectGroup.DELETE("/:sessionId", dappBrowserHandler.DisconnectDApp) // 断开DApp连接
			}

			// Web3请求处理接口
			web3Group := dappGroup.Group("/web3")
			{
				web3Group.POST("/request", dappBrowserHandler.ProcessWeb3Request)                                    // 处理Web3请求
				web3Group.POST("/confirm", middleware.TransactionRateLimit(), dappBrowserHandler.ConfirmWeb3Request) // 确认Web3请求
				web3Group.GET("/pending/:address", dappBrowserHandler.GetPendingRequests)                            // 获取待处理请求
			}

			// DApp发现和浏览接口
			discoveryGroup := dappGroup.Group("/discovery")
			{
				discoveryGroup.GET("/list", dappBrowserHandler.GetDAppList)          // 获取DApp列表
				discoveryGroup.GET("/featured", dappBrowserHandler.GetFeaturedDApps) // 获取推荐DApp
				discoveryGroup.GET("/search", dappBrowserHandler.SearchDApps)        // 搜索DApp
				discoveryGroup.GET("/categories", dappBrowserHandler.GetCategories)  // 获取DApp分类
			}

			// 用户DApp活动接口
			userGroup := dappGroup.Group("/user")
			{
				userGroup.GET("/:address/activity", dappBrowserHandler.GetUserActivity) // 获取用户活动
				userGroup.POST("/favorite", dappBrowserHandler.ManageFavorite)          // 管理收藏
			}
		}

		// 社交功能相关路由组
		// 提供地址簿管理、转账记录分享、社交网络等服务
		socialGroup := v1.Group("/social")
		{
			// 联系人管理接口
			contactsGroup := socialGroup.Group("/contacts")
			{
				contactsGroup.GET("", socialHandler.GetContactList)              // 获取联系人列表
				contactsGroup.POST("", socialHandler.AddContact)                 // 添加联系人
				contactsGroup.GET("/:contactId", socialHandler.GetContact)       // 获取联系人详情
				contactsGroup.PUT("/:contactId", socialHandler.UpdateContact)    // 更新联系人
				contactsGroup.DELETE("/:contactId", socialHandler.DeleteContact) // 删除联系人
			}

			// 分享功能接口
			shareGroup := socialGroup.Group("/share")
			{
				shareGroup.POST("/transaction", socialHandler.ShareTransaction) // 分享交易
				shareGroup.GET("/my", socialHandler.GetMyShares)                // 获取我的分享
				shareGroup.GET("/:shareId", socialHandler.GetShareRecord)       // 获取分享记录
			}

			// 社交网络接口
			networkGroup := socialGroup.Group("/network")
			{
				networkGroup.POST("/action", socialHandler.SocialNetworkAction)     // 社交网络操作
				networkGroup.GET("/:address/followers", socialHandler.GetFollowers) // 获取粉丝列表
				networkGroup.GET("/:address/following", socialHandler.GetFollowing) // 获取关注列表
			}

			// 用户社交资料接口
			userGroup := socialGroup.Group("/user")
			{
				userGroup.GET("/:address/profile", socialHandler.GetUserSocialProfile) // 获取用户社交资料
				userGroup.PUT("/profile", socialHandler.UpdateUserSocialProfile)       // 更新用户社交资料
			}

			// 搜索功能接口
			searchGroup := socialGroup.Group("/search")
			{
				searchGroup.GET("/users", socialHandler.SearchUsers) // 搜索用户
			}
		}

		// 安全功能相关路由组
		// 提供硬件钱包集成、多重签名、MFA认证等安全服务
		securityGroup := v1.Group("/security")
		{
			// 硬件钱包相关接口
			hardwareGroup := securityGroup.Group("/hardware")
			{
				hardwareGroup.GET("/detect", securityHandler.DetectHardwareWallets)          // 检测硬件钱包
				hardwareGroup.POST("/request", securityHandler.ProcessHardwareWalletRequest) // 处理硬件钱包请求
			}

			// 多重签名钱包相关接口
			multisigGroup := securityGroup.Group("/multisig")
			{
				multisigGroup.POST("/create", securityHandler.CreateMultiSigWallet) // 创建多重签名钱包

				// 多重签名交易相关接口
				transactionGroup := multisigGroup.Group("/transaction")
				{
					transactionGroup.POST("/create", securityHandler.CreateMultiSigTransaction)                                // 创建多重签名交易
					transactionGroup.POST("/sign", middleware.TransactionRateLimit(), securityHandler.SignMultiSigTransaction) // 签名多重签名交易
				}
			}

			// 多因素认证相关接口
			mfaGroup := securityGroup.Group("/mfa")
			{
				mfaGroup.POST("/setup", securityHandler.SetupMFA)   // 设置MFA
				mfaGroup.POST("/verify", securityHandler.VerifyMFA) // 验MFA
			}

			// 安全审计相关接口
			auditGroup := securityGroup.Group("/audit")
			{
				auditGroup.GET("/logs", securityHandler.GetSecurityAuditLogs)         // 获取安全审计日志
				auditGroup.GET("/report/:address", securityHandler.GetSecurityReport) // 获取安全报告
			}

			// 生物识别认证相关接口
			biometricGroup := securityGroup.Group("/biometric")
			{
				biometricGroup.POST("/enable", securityHandler.EnableBiometric) // 启用生物识别
				biometricGroup.POST("/verify", securityHandler.VerifyBiometric) // 验证生物识别
			}

			// 安全状态接口
			securityGroup.GET("/status/:address", securityHandler.GetSecurityStatus) // 获取安全状态概览
		}

		// 交易操作相关路由组
		// 所有交易相关操作都应用额外的安全中间件
		txGroup := v1.Group("/transactions")
		txGroup.Use(middleware.TransactionRateLimit())  // 交易频率限制（更严格）
		txGroup.Use(middleware.TransactionValidation()) // 交易参数验证和安全检查
		{
			// 基础交易接口
			txGroup.POST("/send", walletHandler.SendTransaction) // 发送原生代币交易
			txGroup.POST("/send-erc20", walletHandler.SendERC20) // 发送ERC20代币交易

			// 高级交易接口（支持自定义Gas和Nonce）
			txGroup.POST("/send-advanced", walletHandler.SendTransactionAdvanced) // 高级ETH转账（自定义参数）
			txGroup.POST("/send-erc20-advanced", walletHandler.SendERC20Advanced) // 高级ERC20转账（自定义参数）

			// 交易工具接口
			txGroup.POST("/estimate", walletHandler.EstimateTransaction)      // 交易费估算
			txGroup.POST("/broadcast", walletHandler.BroadcastRawTransaction) // 广播签名后的原始交易

			// 交易查询接口
			txGroup.GET("/:hash/receipt", walletHandler.GetTxReceipt) // 获取交易收据和状态
		}

		// ERC20代币相关接口
		v1.GET("/tokens/:token/metadata", walletHandler.GetTokenMetadata) // 获取代币元数据（名称、符号、精度）

		// 代币授权管理接口
		v1.POST("/tokens/:token/approve", middleware.TransactionRateLimit(), middleware.TransactionValidation(), walletHandler.ApproveToken) // 授权代币使用
		v1.GET("/tokens/:token/allowance", walletHandler.GetAllowance)                                                                       // 查询授权额度

		// 消息签名相关接口
		// 用于登录验证、数据确认等场景
		sign := v1.Group("/sign")
		{
			sign.POST("/message", walletHandler.PersonalSign)  // Personal Sign消息签名
			sign.POST("/typed", walletHandler.SignTypedDataV4) // EIP-712结构化数据签名
		}
	}

	// 服务健康检查接口
	// 用于负载均衡器、监控系统和运维工具检查服务状态
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "钱包服务运行正常"})
	})

	return r
}
