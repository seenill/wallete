/*
NFT市场业务服务层

本文件实现了NFT市场功能的业务服务层，提供市场数据查询、价格分析、交易推荐等服务。
*/
package services

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"
	"wallet/core"
)

// NFTMarketplaceService NFT市场服务
type NFTMarketplaceService struct {
	marketplace     *core.NFTMarketplace        // NFT市场管理器
	nftService      *NFTService                 // NFT服务
	userPreferences map[string]*UserMarketPrefs // 用户市场偏好
	watchlists      map[string]*Watchlist       // 用户关注列表
	priceAlerts     map[string]*PriceAlert      // 价格提醒
	mu              sync.RWMutex                // 读写锁
}

// UserMarketPrefs 用户市场偏好
type UserMarketPrefs struct {
	UserAddress          string                `json:"user_address"`          // 用户地址
	PreferredPlatforms   []string              `json:"preferred_platforms"`   // 偏好平台
	PreferredCurrency    string                `json:"preferred_currency"`    // 偏好货币
	PriceRange           *PriceRange           `json:"price_range"`           // 价格范围偏好
	Categories           []string              `json:"categories"`            // 关注分类
	AutoBidEnabled       bool                  `json:"auto_bid_enabled"`      // 自动出价
	MaxAutoBidAmount     *big.Int              `json:"max_auto_bid_amount"`   // 最大自动出价
	NotificationSettings *NotificationSettings `json:"notification_settings"` // 通知设置
}

// PriceRange 价格范围
type PriceRange struct {
	MinPrice *big.Int `json:"min_price"` // 最低价
	MaxPrice *big.Int `json:"max_price"` // 最高价
	Currency string   `json:"currency"`  // 货币类型
}

// NotificationSettings 通知设置
type NotificationSettings struct {
	PriceDrops    bool `json:"price_drops"`    // 价格下跌通知
	NewListings   bool `json:"new_listings"`   // 新挂单通知
	AuctionEnding bool `json:"auction_ending"` // 拍卖结束通知
	Outbid        bool `json:"outbid"`         // 被超越通知
}

// Watchlist 关注列表
type Watchlist struct {
	ID          string           `json:"id"`           // 关注列表ID
	UserAddress string           `json:"user_address"` // 用户地址
	Name        string           `json:"name"`         // 列表名称
	Items       []*WatchlistItem `json:"items"`        // 关注项目
	CreatedAt   time.Time        `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time        `json:"updated_at"`   // 更新时间
}

// WatchlistItem 关注项目
type WatchlistItem struct {
	Type     string    `json:"type"`      // 类型（collection/nft）
	Contract string    `json:"contract"`  // 合约地址
	TokenID  string    `json:"token_id"`  // Token ID（NFT特定时使用）
	Name     string    `json:"name"`      // 名称
	ImageURL string    `json:"image_url"` // 图片URL
	AddedAt  time.Time `json:"added_at"`  // 添加时间
}

// PriceAlert 价格提醒
type PriceAlert struct {
	ID          string            `json:"id"`           // 提醒ID
	UserAddress string            `json:"user_address"` // 用户地址
	Contract    string            `json:"contract"`     // 合约地址
	TokenID     string            `json:"token_id"`     // Token ID（可选）
	AlertType   string            `json:"alert_type"`   // 提醒类型（above/below/change）
	TargetPrice *core.MarketPrice `json:"target_price"` // 目标价格
	IsActive    bool              `json:"is_active"`    // 是否激活
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	TriggeredAt *time.Time        `json:"triggered_at"` // 触发时间
}

// MarketAnalysisRequest 市场分析请求
type MarketAnalysisRequest struct {
	Contract     string `json:"contract"`      // 合约地址
	TokenID      string `json:"token_id"`      // Token ID（可选）
	AnalysisType string `json:"analysis_type"` // 分析类型
	TimeRange    string `json:"time_range"`    // 时间范围
	Platform     string `json:"platform"`      // 平台过滤
}

// MarketAnalysisResponse 市场分析响应
type MarketAnalysisResponse struct {
	Analysis        *CustomMarketAnalysis `json:"analysis"`         // 市场分析
	PriceProjection *PriceProjection      `json:"price_projection"` // 价格预测
	TradingAdvice   *TradingAdvice        `json:"trading_advice"`   // 交易建议
	RiskAssessment  *CustomRiskAssessment `json:"risk_assessment"`  // 风险评估
	ComparableItems []*ComparableItem     `json:"comparable_items"` // 可比较项目
}

// CustomMarketAnalysis 自定义市场分析（避免与core包冲突）
type CustomMarketAnalysis struct {
	Contract        string             `json:"contract"`         // 合约地址
	TokenID         string             `json:"token_id"`         // Token ID
	CurrentPrice    *core.MarketPrice  `json:"current_price"`    // 当前价格
	FloorPrice      *core.MarketPrice  `json:"floor_price"`      // 地板价
	PriceHistory    *core.PriceHistory `json:"price_history"`    // 价格历史
	Volatility      float64            `json:"volatility"`       // 波动率
	Liquidity       string             `json:"liquidity"`        // 流动性评级
	MarketCap       *core.MarketPrice  `json:"market_cap"`       // 市值
	TradingVolume   *core.MarketPrice  `json:"trading_volume"`   // 交易量
	HolderAnalysis  *HolderAnalysis    `json:"holder_analysis"`  // 持有者分析
	TrendIndicators *TrendIndicators   `json:"trend_indicators"` // 趋势指标
}

// HolderAnalysis 持有者分析
type HolderAnalysis struct {
	TotalHolders      int     `json:"total_holders"`      // 总持有者数
	UniqueHolders     int     `json:"unique_holders"`     // 唯一持有者数
	WhaleHolders      int     `json:"whale_holders"`      // 巨鲸持有者数
	DistributionScore float64 `json:"distribution_score"` // 分布评分
	ConcentrationRisk string  `json:"concentration_risk"` // 集中度风险
}

// TrendIndicators 趋势指标
type TrendIndicators struct {
	SMA7           *core.MarketPrice `json:"sma_7"`           // 7日简单移动平均
	SMA30          *core.MarketPrice `json:"sma_30"`          // 30日简单移动平均
	RSI            float64           `json:"rsi"`             // 相对强弱指数
	MACD           float64           `json:"macd"`            // MACD指标
	BollingerBands *BollingerBands   `json:"bollinger_bands"` // 布林带
	TrendDirection string            `json:"trend_direction"` // 趋势方向
	TrendStrength  float64           `json:"trend_strength"`  // 趋势强度
}

// BollingerBands 布林带
type BollingerBands struct {
	UpperBand  *core.MarketPrice `json:"upper_band"`  // 上轨
	MiddleBand *core.MarketPrice `json:"middle_band"` // 中轨
	LowerBand  *core.MarketPrice `json:"lower_band"`  // 下轨
}

// PriceProjection 价格预测
type PriceProjection struct {
	TimeRange   string                  `json:"time_range"`  // 预测时间范围
	Projections []*PriceProjectionPoint `json:"projections"` // 预测点
	Confidence  float64                 `json:"confidence"`  // 置信度
	Method      string                  `json:"method"`      // 预测方法
	Factors     []string                `json:"factors"`     // 影响因素
}

// PriceProjectionPoint 价格预测点
type PriceProjectionPoint struct {
	Date           time.Time         `json:"date"`            // 日期
	PredictedPrice *core.MarketPrice `json:"predicted_price"` // 预测价格
	LowerBound     *core.MarketPrice `json:"lower_bound"`     // 下界
	UpperBound     *core.MarketPrice `json:"upper_bound"`     // 上界
	Probability    float64           `json:"probability"`     // 概率
}

// TradingAdvice 交易建议
type TradingAdvice struct {
	Recommendation string            `json:"recommendation"` // 建议（buy/sell/hold）
	Confidence     float64           `json:"confidence"`     // 信心度
	Reasoning      []string          `json:"reasoning"`      // 推理原因
	OptimalPrice   *core.MarketPrice `json:"optimal_price"`  // 最优价格
	TimeHorizon    string            `json:"time_horizon"`   // 时间范围
	RiskLevel      string            `json:"risk_level"`     // 风险等级
	StopLoss       *core.MarketPrice `json:"stop_loss"`      // 止损价
	TakeProfit     *core.MarketPrice `json:"take_profit"`    // 止盈价
}

// CustomRiskAssessment 自定义风险评估（避免与core包冲突）
type CustomRiskAssessment struct {
	OverallRisk       string        `json:"overall_risk"`       // 总体风险
	RiskScore         float64       `json:"risk_score"`         // 风险评分（0-100）
	RiskFactors       []*RiskFactor `json:"risk_factors"`       // 风险因素
	LiquidityRisk     string        `json:"liquidity_risk"`     // 流动性风险
	VolatilityRisk    string        `json:"volatility_risk"`    // 波动性风险
	ConcentrationRisk string        `json:"concentration_risk"` // 集中度风险
	TechnicalRisk     string        `json:"technical_risk"`     // 技术风险
}

// RiskFactor 风险因素
type RiskFactor struct {
	Type        string  `json:"type"`        // 风险类型
	Level       string  `json:"level"`       // 风险等级
	Impact      float64 `json:"impact"`      // 影响度
	Description string  `json:"description"` // 描述
}

// ComparableItem 可比较项目
type ComparableItem struct {
	Contract     string            `json:"contract"`      // 合约地址
	TokenID      string            `json:"token_id"`      // Token ID
	Name         string            `json:"name"`          // 名称
	ImageURL     string            `json:"image_url"`     // 图片URL
	CurrentPrice *core.MarketPrice `json:"current_price"` // 当前价格
	Similarity   float64           `json:"similarity"`    // 相似度
	PriceDiff    float64           `json:"price_diff"`    // 价格差异
	Reasons      []string          `json:"reasons"`       // 相似原因
}

// NewNFTMarketplaceService 创建NFT市场服务
func NewNFTMarketplaceService(nftService *NFTService) *NFTMarketplaceService {
	return &NFTMarketplaceService{
		marketplace:     core.NewNFTMarketplace(),
		nftService:      nftService,
		userPreferences: make(map[string]*UserMarketPrefs),
		watchlists:      make(map[string]*Watchlist),
		priceAlerts:     make(map[string]*PriceAlert),
	}
}

// GetMarketListings 获取市场挂单
func (nms *NFTMarketplaceService) GetMarketListings(ctx context.Context, userAddress string, request *core.MarketListingRequest) ([]*core.MarketListing, error) {
	// 应用用户偏好
	if prefs := nms.getUserPreferences(userAddress); prefs != nil {
		nms.applyUserPreferencesToListingRequest(request, prefs)
	}

	return nms.marketplace.GetMarketListings(ctx, request)
}

// GetMarketTransactions 获取市场交易记录
func (nms *NFTMarketplaceService) GetMarketTransactions(ctx context.Context, userAddress string, request *core.MarketTransactionRequest) ([]*core.MarketTransaction, error) {
	return nms.marketplace.GetMarketTransactions(ctx, request)
}

// GetMarketStats 获取市场统计数据
func (nms *NFTMarketplaceService) GetMarketStats(ctx context.Context, contract, platform string) (*core.MarketStats, error) {
	return nms.marketplace.GetMarketStats(ctx, contract, platform)
}

// GetPriceHistory 获取价格历史数据
func (nms *NFTMarketplaceService) GetPriceHistory(ctx context.Context, contract, tokenID, platform, timeRange string) (*core.PriceHistory, error) {
	return nms.marketplace.GetPriceHistory(ctx, contract, tokenID, platform, timeRange)
}

// AnalyzeMarket 市场分析
func (nms *NFTMarketplaceService) AnalyzeMarket(ctx context.Context, userAddress string, request *MarketAnalysisRequest) (*MarketAnalysisResponse, error) {
	// 获取基础数据
	stats, err := nms.marketplace.GetMarketStats(ctx, request.Contract, request.Platform)
	if err != nil {
		return nil, fmt.Errorf("获取市场统计失败: %w", err)
	}

	priceHistory, err := nms.marketplace.GetPriceHistory(ctx, request.Contract, request.TokenID, request.Platform, request.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("获取价格历史失败: %w", err)
	}

	// 执行市场分析
	analysis := nms.performMarketAnalysis(stats, priceHistory)

	// 生成价格预测
	projection := nms.generatePriceProjection(priceHistory, request.TimeRange)

	// 提供交易建议
	advice := nms.generateTradingAdvice(analysis, projection, userAddress)

	// 评估风险
	riskAssessment := nms.assessRisk(analysis, stats)

	// 查找可比较项目
	comparableItems := nms.findComparableItems(ctx, request.Contract, request.TokenID)

	response := &MarketAnalysisResponse{
		Analysis:        analysis,
		PriceProjection: projection,
		TradingAdvice:   advice,
		RiskAssessment:  riskAssessment,
		ComparableItems: comparableItems,
	}

	return response, nil
}

// SetUserPreferences 设置用户市场偏好
func (nms *NFTMarketplaceService) SetUserPreferences(userAddress string, prefs *UserMarketPrefs) {
	nms.mu.Lock()
	defer nms.mu.Unlock()

	prefs.UserAddress = userAddress
	nms.userPreferences[userAddress] = prefs
}

// GetUserPreferences 获取用户市场偏好
func (nms *NFTMarketplaceService) GetUserPreferences(userAddress string) *UserMarketPrefs {
	return nms.getUserPreferences(userAddress)
}

// AddToWatchlist 添加到关注列表
func (nms *NFTMarketplaceService) AddToWatchlist(userAddress, listName, itemType, contract, tokenID string) error {
	nms.mu.Lock()
	defer nms.mu.Unlock()

	listID := fmt.Sprintf("%s_%s", userAddress, listName)

	watchlist, exists := nms.watchlists[listID]
	if !exists {
		watchlist = &Watchlist{
			ID:          listID,
			UserAddress: userAddress,
			Name:        listName,
			Items:       make([]*WatchlistItem, 0),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		nms.watchlists[listID] = watchlist
	}

	// 检查是否已存在
	for _, item := range watchlist.Items {
		if item.Contract == contract && item.TokenID == tokenID {
			return fmt.Errorf("项目已在关注列表中")
		}
	}

	// 添加新项目
	item := &WatchlistItem{
		Type:     itemType,
		Contract: contract,
		TokenID:  tokenID,
		AddedAt:  time.Now(),
	}

	// 获取名称和图片（简化实现）
	item.Name = fmt.Sprintf("NFT %s #%s", contract[:10], tokenID)
	item.ImageURL = "https://example.com/placeholder.png"

	watchlist.Items = append(watchlist.Items, item)
	watchlist.UpdatedAt = time.Now()

	return nil
}

// GetWatchlist 获取关注列表
func (nms *NFTMarketplaceService) GetWatchlist(userAddress, listName string) (*Watchlist, error) {
	nms.mu.RLock()
	defer nms.mu.RUnlock()

	listID := fmt.Sprintf("%s_%s", userAddress, listName)
	if watchlist, exists := nms.watchlists[listID]; exists {
		return watchlist, nil
	}

	return nil, fmt.Errorf("关注列表不存在")
}

// CreatePriceAlert 创建价格提醒
func (nms *NFTMarketplaceService) CreatePriceAlert(userAddress, contract, tokenID, alertType string, targetPrice *core.MarketPrice) (*PriceAlert, error) {
	nms.mu.Lock()
	defer nms.mu.Unlock()

	alertID := fmt.Sprintf("alert_%d", time.Now().UnixNano())

	alert := &PriceAlert{
		ID:          alertID,
		UserAddress: userAddress,
		Contract:    contract,
		TokenID:     tokenID,
		AlertType:   alertType,
		TargetPrice: targetPrice,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	nms.priceAlerts[alertID] = alert

	return alert, nil
}

// GetPriceAlerts 获取价格提醒
func (nms *NFTMarketplaceService) GetPriceAlerts(userAddress string) ([]*PriceAlert, error) {
	nms.mu.RLock()
	defer nms.mu.RUnlock()

	var alerts []*PriceAlert
	for _, alert := range nms.priceAlerts {
		if alert.UserAddress == userAddress {
			alerts = append(alerts, alert)
		}
	}

	// 按创建时间排序
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})

	return alerts, nil
}

// 私有方法

// getUserPreferences 获取用户偏好
func (nms *NFTMarketplaceService) getUserPreferences(userAddress string) *UserMarketPrefs {
	nms.mu.RLock()
	defer nms.mu.RUnlock()

	return nms.userPreferences[userAddress]
}

// applyUserPreferencesToListingRequest 应用用户偏好到挂单请求
func (nms *NFTMarketplaceService) applyUserPreferencesToListingRequest(request *core.MarketListingRequest, prefs *UserMarketPrefs) {
	// 应用偏好平台
	if len(prefs.PreferredPlatforms) > 0 && request.Platform == "" {
		request.Platform = prefs.PreferredPlatforms[0]
	}

	// 应用价格范围
	if prefs.PriceRange != nil {
		if request.MinPrice == nil {
			request.MinPrice = prefs.PriceRange.MinPrice
		}
		if request.MaxPrice == nil {
			request.MaxPrice = prefs.PriceRange.MaxPrice
		}
	}

	// 应用货币偏好
	if request.Currency == "" && prefs.PreferredCurrency != "" {
		request.Currency = prefs.PreferredCurrency
	}
}

// performMarketAnalysis 执行市场分析
func (nms *NFTMarketplaceService) performMarketAnalysis(stats *core.MarketStats, history *core.PriceHistory) *CustomMarketAnalysis {
	// 简化实现：基础市场分析
	analysis := &CustomMarketAnalysis{
		Contract:      stats.Contract,
		CurrentPrice:  stats.AveragePrice,
		FloorPrice:    stats.FloorPrice,
		PriceHistory:  history,
		Volatility:    nms.calculateVolatility(history),
		Liquidity:     nms.assessLiquidity(stats),
		MarketCap:     stats.Volume30d,
		TradingVolume: stats.Volume24h,
		HolderAnalysis: &HolderAnalysis{
			TotalHolders:      stats.OwnersCount,
			UniqueHolders:     stats.OwnersCount,
			WhaleHolders:      stats.OwnersCount / 10, // 假设10%是巨鲸
			DistributionScore: 0.7,
			ConcentrationRisk: "medium",
		},
		TrendIndicators: nms.calculateTrendIndicators(history),
	}

	return analysis
}

// calculateVolatility 计算波动率
func (nms *NFTMarketplaceService) calculateVolatility(history *core.PriceHistory) float64 {
	// 简化实现：返回固定值
	return 0.25 // 25%波动率
}

// assessLiquidity 评估流动性
func (nms *NFTMarketplaceService) assessLiquidity(stats *core.MarketStats) string {
	if stats.Sales24h > 10 {
		return "high"
	} else if stats.Sales24h > 3 {
		return "medium"
	}
	return "low"
}

// calculateTrendIndicators 计算趋势指标
func (nms *NFTMarketplaceService) calculateTrendIndicators(history *core.PriceHistory) *TrendIndicators {
	// 简化实现：返回示例数据
	return &TrendIndicators{
		SMA7: &core.MarketPrice{
			Amount:   big.NewInt(800000000000000000), // 0.8 ETH
			Currency: "ETH",
			USDValue: 1600.0,
		},
		SMA30: &core.MarketPrice{
			Amount:   big.NewInt(750000000000000000), // 0.75 ETH
			Currency: "ETH",
			USDValue: 1500.0,
		},
		RSI:            65.0,
		MACD:           0.05,
		TrendDirection: "up",
		TrendStrength:  0.7,
	}
}

// generatePriceProjection 生成价格预测
func (nms *NFTMarketplaceService) generatePriceProjection(history *core.PriceHistory, timeRange string) *PriceProjection {
	// 简化实现：返回示例预测
	return &PriceProjection{
		TimeRange: timeRange,
		Projections: []*PriceProjectionPoint{
			{
				Date: time.Now().Add(7 * 24 * time.Hour),
				PredictedPrice: &core.MarketPrice{
					Amount:   big.NewInt(1100000000000000000), // 1.1 ETH
					Currency: "ETH",
					USDValue: 2200.0,
				},
				LowerBound: &core.MarketPrice{
					Amount:   big.NewInt(950000000000000000), // 0.95 ETH
					Currency: "ETH",
					USDValue: 1900.0,
				},
				UpperBound: &core.MarketPrice{
					Amount:   big.NewInt(1250000000000000000), // 1.25 ETH
					Currency: "ETH",
					USDValue: 2500.0,
				},
				Probability: 0.8,
			},
		},
		Confidence: 0.75,
		Method:     "machine_learning",
		Factors:    []string{"历史趋势", "市场情绪", "交易量变化"},
	}
}

// generateTradingAdvice 生成交易建议
func (nms *NFTMarketplaceService) generateTradingAdvice(analysis *CustomMarketAnalysis, projection *PriceProjection, userAddress string) *TradingAdvice {
	// 简化实现：基于分析生成建议
	advice := &TradingAdvice{
		Recommendation: "hold",
		Confidence:     0.7,
		Reasoning:      []string{"价格趋势向上", "流动性良好", "技术指标积极"},
		TimeHorizon:    "short_term",
		RiskLevel:      "medium",
	}

	// 根据用户偏好调整建议
	if prefs := nms.getUserPreferences(userAddress); prefs != nil {
		// 考虑用户的风险偏好等
		advice.Confidence *= 0.9 // 调整信心度
	}

	return advice
}

// assessRisk 评估风险
func (nms *NFTMarketplaceService) assessRisk(analysis *CustomMarketAnalysis, stats *core.MarketStats) *CustomRiskAssessment {
	// 简化实现：基础风险评估
	return &CustomRiskAssessment{
		OverallRisk: "medium",
		RiskScore:   45.0,
		RiskFactors: []*RiskFactor{
			{
				Type:        "liquidity",
				Level:       "low",
				Impact:      0.3,
				Description: "流动性风险较低",
			},
			{
				Type:        "volatility",
				Level:       "medium",
				Impact:      0.5,
				Description: "价格波动中等",
			},
		},
		LiquidityRisk:     "low",
		VolatilityRisk:    "medium",
		ConcentrationRisk: "medium",
		TechnicalRisk:     "low",
	}
}

// findComparableItems 查找可比较项目
func (nms *NFTMarketplaceService) findComparableItems(ctx context.Context, contract, tokenID string) []*ComparableItem {
	// 简化实现：返回示例可比较项目
	return []*ComparableItem{
		{
			Contract: "0x9876543210987654321098765432109876543210",
			TokenID:  "42",
			Name:     "Similar NFT #42",
			ImageURL: "https://example.com/similar-nft.png",
			CurrentPrice: &core.MarketPrice{
				Amount:   big.NewInt(950000000000000000), // 0.95 ETH
				Currency: "ETH",
				USDValue: 1900.0,
			},
			Similarity: 0.85,
			PriceDiff:  -5.0, // 比当前NFT便宜5%
			Reasons:    []string{"相同艺术家", "相似风格", "相同稀有度"},
		},
	}
}
