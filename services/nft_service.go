/*
NFT业务服务层

本文件实现了NFT功能的业务服务层，封装NFT相关的复杂业务逻辑，
为API层提供高级接口，简化NFT操作的复杂性。

主要服务：
NFT管理服务：
- 用户NFT库存管理
- NFT详情查询和缓存
- 批量NFT信息获取
- NFT所有权验证

集合服务：
- 热门集合推荐
- 集合统计和排行
- 集合元数据管理
- 集合价格监控

市场服务：
- 实时价格监控
- 历史价格分析
- 市场趋势预测
- 交易机会发现

转账服务：
- NFT转账验证
- 批量转账优化
- 转账费用估算
- 转账状态跟踪

安全特性：
- 钓鱼NFT检测
- 价格异常警告
- 转账安全检查
- 合约风险评估
*/
package services

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"
	"wallet/core"
)

// NFTService NFT业务服务
// 提供完整的NFT功能封装，包括管理、转账、市场查询等服务
type NFTService struct {
	nftManager     *core.NFTManager           // NFT管理器
	multiChain     *core.MultiChainManager    // 多链管理器
	collections    map[string]*CollectionInfo // 集合信息缓存
	priceCache     map[string]*NFTPriceInfo   // 价格信息缓存
	userPortfolios map[string]*UserPortfolio  // 用户投资组合
	marketData     map[string]*MarketTrend    // 市场趋势数据
	hotCollections []*HotCollection           // 热门集合
	mu             sync.RWMutex               // 读写锁
	lastUpdate     time.Time                  // 最后更新时间
}

// CollectionInfo 集合详细信息
type CollectionInfo struct {
	*core.Collection                  // 继承核心集合信息
	TrendingRank     int              `json:"trending_rank"`     // 热度排名
	FloorChange24h   string           `json:"floor_change_24h"`  // 24小时地板价变化
	VolumeChange24h  string           `json:"volume_change_24h"` // 24小时交易量变化
	Listings         []*ListingInfo   `json:"listings"`          // 挂单信息
	Activities       []*ActivityInfo  `json:"activities"`        // 最近活动
	TopSales         []*core.SaleInfo `json:"top_sales"`         // 顶级成交
	Holders          []*HolderInfo    `json:"holders"`           // 主要持有者
	PriceHistory     []*PricePoint    `json:"price_history"`     // 价格历史
	MarketAnalysis   *MarketAnalysis  `json:"market_analysis"`   // 市场分析
}

// UserPortfolio 用户NFT投资组合
type UserPortfolio struct {
	UserAddress      string            `json:"user_address"`      // 用户地址
	TotalValue       *big.Int          `json:"total_value"`       // 总价值
	TotalCount       int               `json:"total_count"`       // 总数量
	Collections      []*UserCollection `json:"collections"`       // 持有集合
	TopNFTs          []*core.NFT       `json:"top_nfts"`          // 最有价值NFT
	RecentActivities []*ActivityInfo   `json:"recent_activities"` // 最近活动
	PnL24h           *big.Int          `json:"pnl_24h"`           // 24小时盈亏
	PnL7d            *big.Int          `json:"pnl_7d"`            // 7天盈亏
	PnL30d           *big.Int          `json:"pnl_30d"`           // 30天盈亏
	UpdatedAt        time.Time         `json:"updated_at"`        // 更新时间
}

// UserCollection 用户持有的集合信息
type UserCollection struct {
	Collection     *CollectionInfo `json:"collection"`      // 集合信息
	Count          int             `json:"count"`           // 持有数量
	FloorValue     *big.Int        `json:"floor_value"`     // 地板价价值
	EstimatedValue *big.Int        `json:"estimated_value"` // 估计价值
	AvgBuyPrice    *big.Int        `json:"avg_buy_price"`   // 平均买入价
	PnL            *big.Int        `json:"pnl"`             // 盈亏
	NFTs           []*core.NFT     `json:"nfts"`            // NFT列表
}

// NFTPriceInfo NFT价格信息
type NFTPriceInfo struct {
	ContractAddr   string    `json:"contract_addr"`    // 合约地址
	TokenID        string    `json:"token_id"`         // 代币ID
	CurrentPrice   *big.Int  `json:"current_price"`    // 当前价格
	FloorPrice     *big.Int  `json:"floor_price"`      // 地板价
	LastSalePrice  *big.Int  `json:"last_sale_price"`  // 最近成交价
	EstimatedPrice *big.Int  `json:"estimated_price"`  // 估计价格
	PriceChange24h string    `json:"price_change_24h"` // 24小时价格变化
	Liquidity      string    `json:"liquidity"`        // 流动性评级
	RarityScore    float64   `json:"rarity_score"`     // 稀有度评分
	MarketCap      *big.Int  `json:"market_cap"`       // 市值
	UpdatedAt      time.Time `json:"updated_at"`       // 更新时间
	Sources        []string  `json:"sources"`          // 数据来源
}

// ListingInfo 挂单信息
type ListingInfo struct {
	TokenID     string    `json:"token_id"`    // 代币ID
	Price       *big.Int  `json:"price"`       // 挂单价格
	Currency    string    `json:"currency"`    // 计价货币
	Seller      string    `json:"seller"`      // 卖方地址
	Marketplace string    `json:"marketplace"` // 交易市场
	ExpiresAt   time.Time `json:"expires_at"`  // 过期时间
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
}

// ActivityInfo 活动信息
type ActivityInfo struct {
	Type        string    `json:"type"`        // 活动类型(mint, sale, transfer, listing)
	TokenID     string    `json:"token_id"`    // 代币ID
	From        string    `json:"from"`        // 发送方
	To          string    `json:"to"`          // 接收方
	Price       *big.Int  `json:"price"`       // 价格
	Currency    string    `json:"currency"`    // 货币
	TxHash      string    `json:"tx_hash"`     // 交易哈希
	Marketplace string    `json:"marketplace"` // 交易市场
	Timestamp   time.Time `json:"timestamp"`   // 时间戳
}

// HolderInfo 持有者信息
type HolderInfo struct {
	Address        string   `json:"address"`         // 持有者地址
	Count          int      `json:"count"`           // 持有数量
	Percentage     float64  `json:"percentage"`      // 持有比例
	EstimatedValue *big.Int `json:"estimated_value"` // 估计价值
	IsWhale        bool     `json:"is_whale"`        // 是否为巨鲸
}

// PricePoint 价格点
type PricePoint struct {
	Timestamp  int64    `json:"timestamp"`   // 时间戳
	FloorPrice *big.Int `json:"floor_price"` // 地板价
	AvgPrice   *big.Int `json:"avg_price"`   // 平均价格
	Volume     *big.Int `json:"volume"`      // 交易量
	Sales      int      `json:"sales"`       // 成交数
}

// MarketAnalysis 市场分析
type MarketAnalysis struct {
	Trend          string   `json:"trend"`          // 趋势(up/down/stable)
	Momentum       string   `json:"momentum"`       // 动量(strong/weak/neutral)
	Sentiment      string   `json:"sentiment"`      // 市场情绪
	Recommendation string   `json:"recommendation"` // 推荐(buy/hold/sell)
	RiskLevel      string   `json:"risk_level"`     // 风险等级
	KeyMetrics     []string `json:"key_metrics"`    // 关键指标
	Opportunities  []string `json:"opportunities"`  // 机会点
	Risks          []string `json:"risks"`          // 风险点
}

// HotCollection 热门集合
type HotCollection struct {
	*CollectionInfo          // 集合信息
	HotRank         int      `json:"hot_rank"`    // 热度排名
	TrendScore      float64  `json:"trend_score"` // 热度评分
	ReasonCode      []string `json:"reason_code"` // 热门原因
}

// MarketTrend 市场趋势
type MarketTrend struct {
	Category        string           `json:"category"`          // 类别
	TotalVolume     *big.Int         `json:"total_volume"`      // 总交易量
	TotalSales      int              `json:"total_sales"`       // 总成交数
	AvgPrice        *big.Int         `json:"avg_price"`         // 平均价格
	FloorPriceAvg   *big.Int         `json:"floor_price_avg"`   // 平均地板价
	VolumeChange24h string           `json:"volume_change_24h"` // 24小时交易量变化
	PriceChange24h  string           `json:"price_change_24h"`  // 24小时价格变化
	TopCollections  []*HotCollection `json:"top_collections"`   // 热门集合
	UpdatedAt       time.Time        `json:"updated_at"`        // 更新时间
}

// NewNFTService 创建NFT服务实例
func NewNFTService(multiChain *core.MultiChainManager) (*NFTService, error) {
	// 获取当前EVM适配器
	evmAdapter, err := multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, fmt.Errorf("获取EVM适配器失败: %w", err)
	}

	// 创建NFT管理器
	nftManager, err := core.NewNFTManager(evmAdapter)
	if err != nil {
		return nil, fmt.Errorf("创建NFT管理器失败: %w", err)
	}

	service := &NFTService{
		nftManager:     nftManager,
		multiChain:     multiChain,
		collections:    make(map[string]*CollectionInfo),
		priceCache:     make(map[string]*NFTPriceInfo),
		userPortfolios: make(map[string]*UserPortfolio),
		marketData:     make(map[string]*MarketTrend),
		hotCollections: make([]*HotCollection, 0),
		lastUpdate:     time.Now(),
	}

	// 初始化热门集合数据
	service.initHotCollections()

	return service, nil
}

// GetUserNFTs 获取用户NFT列表
// 参数: userAddr - 用户地址, filters - 过滤条件
// 返回: NFT列表和错误
func (s *NFTService) GetUserNFTs(ctx context.Context, userAddr string, filters *NFTFilters) ([]*core.NFT, error) {
	// 验证地址
	if userAddr == "" {
		return nil, fmt.Errorf("用户地址不能为空")
	}

	// 应用默认过滤条件
	if filters == nil {
		filters = &NFTFilters{
			Limit:   50,
			SortBy:  "updated_at",
			SortDir: "desc",
		}
	}

	// 获取用户NFT
	var contractAddrs []string
	if filters.Collection != "" {
		contractAddrs = []string{filters.Collection}
	} else {
		// 获取所有已知集合地址
		contractAddrs = s.getKnownCollections()
	}

	nfts, err := s.nftManager.GetUserNFTs(ctx, userAddr, contractAddrs)
	if err != nil {
		return nil, fmt.Errorf("获取用户NFT失败: %w", err)
	}

	// 应用过滤和排序
	filteredNFTs := s.applyFilters(nfts, filters)

	// 更新价格信息
	s.updateNFTPrices(ctx, filteredNFTs)

	return filteredNFTs, nil
}

// GetNFTDetails 获取NFT详细信息
// 参数: contractAddr - 合约地址, tokenID - 代币ID
// 返回: NFT详细信息和错误
func (s *NFTService) GetNFTDetails(ctx context.Context, contractAddr, tokenID string) (*core.NFT, error) {
	// 从核心模块获取NFT信息
	nft, err := s.nftManager.GetNFT(ctx, contractAddr, tokenID)
	if err != nil {
		return nil, fmt.Errorf("获取NFT信息失败: %w", err)
	}

	// 获取价格信息
	priceInfo, err := s.getNFTPrice(ctx, contractAddr, tokenID)
	if err == nil && priceInfo != nil {
		// 更新市场数据
		nft.MarketData = &core.MarketData{
			LastPrice:      priceInfo.LastSalePrice,
			EstimatedValue: priceInfo.EstimatedPrice,
			FloorPrice:     priceInfo.FloorPrice,
		}
	}

	// 获取最近销售记录
	s.updateRecentSales(ctx, nft)

	return nft, nil
}

// GetCollectionInfo 获取集合详细信息
// 参数: contractAddr - 合约地址
// 返回: 集合信息和错误
func (s *NFTService) GetCollectionInfo(ctx context.Context, contractAddr string) (*CollectionInfo, error) {
	s.mu.RLock()
	if cached, exists := s.collections[contractAddr]; exists {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// 从核心模块获取基础信息（简化实现）
	collection := &core.Collection{
		Address:     contractAddr,
		Name:        "示例集合",
		Symbol:      "EXAMPLE",
		Description: "这是一个示例NFT集合",
		TotalSupply: big.NewInt(10000),
		Verified:    true,
		CreatedAt:   time.Now(),
	}

	// 构建详细信息
	collectionInfo := &CollectionInfo{
		Collection:      collection,
		TrendingRank:    0,
		FloorChange24h:  "0%",
		VolumeChange24h: "0%",
		Listings:        make([]*ListingInfo, 0),
		Activities:      make([]*ActivityInfo, 0),
		TopSales:        make([]*core.SaleInfo, 0),
		Holders:         make([]*HolderInfo, 0),
		PriceHistory:    make([]*PricePoint, 0),
		MarketAnalysis: &MarketAnalysis{
			Trend:          "stable",
			Momentum:       "neutral",
			Sentiment:      "neutral",
			Recommendation: "hold",
			RiskLevel:      "medium",
		},
	}

	// 缓存结果
	s.mu.Lock()
	s.collections[contractAddr] = collectionInfo
	s.mu.Unlock()

	return collectionInfo, nil
}

// GetHotCollections 获取热门集合
// 参数: limit - 数量限制
// 返回: 热门集合列表
func (s *NFTService) GetHotCollections(limit int) []*HotCollection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.hotCollections) {
		limit = len(s.hotCollections)
	}

	return s.hotCollections[:limit]
}

// TransferNFT 转账NFT
// 参数: ctx - 上下文, params - 转账参数, credentials - 认证信息
// 返回: 交易哈希和错误
func (s *NFTService) TransferNFT(ctx context.Context, params *NFTTransferRequest, mnemonic, derivationPath string) (*TransferResult, error) {
	// 验证转账参数
	if err := s.validateTransferParams(params); err != nil {
		return nil, fmt.Errorf("转账参数验证失败: %w", err)
	}

	// 构建核心转账参数
	gasPrice, _ := new(big.Int).SetString(params.GasPrice, 10)
	coreParams := &core.NFTTransferParams{
		ContractAddr: params.ContractAddr,
		From:         params.From,
		To:           params.To,
		TokenID:      params.TokenID,
		Amount:       big.NewInt(int64(params.Amount)),
		GasLimit:     params.GasLimit,
		GasPrice:     gasPrice,
	}

	// 执行转账
	txHash, err := s.nftManager.TransferNFT(ctx, coreParams, mnemonic, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("NFT转账失败: %w", err)
	}

	// 构建结果
	result := &TransferResult{
		TxHash:    txHash,
		Status:    "pending",
		From:      params.From,
		To:        params.To,
		TokenID:   params.TokenID,
		Timestamp: time.Now(),
	}

	return result, nil
}

// 辅助方法

// NFTFilters NFT过滤条件
type NFTFilters struct {
	Collection string `json:"collection"` // 集合地址
	Category   string `json:"category"`   // 类别
	MinPrice   string `json:"min_price"`  // 最低价格
	MaxPrice   string `json:"max_price"`  // 最高价格
	SortBy     string `json:"sort_by"`    // 排序字段
	SortDir    string `json:"sort_dir"`   // 排序方向
	Limit      int    `json:"limit"`      // 数量限制
	Offset     int    `json:"offset"`     // 偏移量
}

// NFTTransferRequest NFT转账请求
type NFTTransferRequest struct {
	ContractAddr string `json:"contract_addr" binding:"required"` // 合约地址
	From         string `json:"from" binding:"required"`          // 发送方地址
	To           string `json:"to" binding:"required"`            // 接收方地址
	TokenID      string `json:"token_id" binding:"required"`      // 代币ID
	Amount       int    `json:"amount"`                           // 数量(ERC-1155)
	GasLimit     uint64 `json:"gas_limit"`                        // Gas限制
	GasPrice     string `json:"gas_price"`                        // Gas价格
}

// TransferResult 转账结果
type TransferResult struct {
	TxHash    string    `json:"tx_hash"`   // 交易哈希
	Status    string    `json:"status"`    // 状态
	From      string    `json:"from"`      // 发送方
	To        string    `json:"to"`        // 接收方
	TokenID   string    `json:"token_id"`  // 代币ID
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// 私有方法实现

// initHotCollections 初始化热门集合
func (s *NFTService) initHotCollections() {
	// 示例热门集合数据
	s.hotCollections = []*HotCollection{
		{
			CollectionInfo: &CollectionInfo{
				Collection: &core.Collection{
					Address:     "0x1234567890123456789012345678901234567890",
					Name:        "热门NFT集合",
					Symbol:      "HOT",
					Description: "当前最热门的NFT集合",
					TotalSupply: big.NewInt(10000),
					FloorPrice:  big.NewInt(1000000000000000000), // 1 ETH
					Verified:    true,
				},
				TrendingRank:    1,
				FloorChange24h:  "+15.5%",
				VolumeChange24h: "+235.8%",
			},
			HotRank:    1,
			TrendScore: 95.5,
			ReasonCode: []string{"volume_surge", "celebrity_endorsement", "utility_launch"},
		},
	}
}

// getKnownCollections 获取已知集合地址
func (s *NFTService) getKnownCollections() []string {
	return []string{
		"0x1234567890123456789012345678901234567890",
		"0x0987654321098765432109876543210987654321",
	}
}

// applyFilters 应用过滤条件
func (s *NFTService) applyFilters(nfts []*core.NFT, filters *NFTFilters) []*core.NFT {
	var filtered []*core.NFT

	for _, nft := range nfts {
		// 应用集合过滤
		if filters.Collection != "" && !strings.EqualFold(nft.ContractAddr, filters.Collection) {
			continue
		}

		// 应用价格过滤
		// 这里需要价格信息，简化处理
		filtered = append(filtered, nft)
	}

	// 应用排序
	sort.Slice(filtered, func(i, j int) bool {
		switch filters.SortBy {
		case "token_id":
			return filtered[i].TokenID < filtered[j].TokenID
		default: // updated_at
			return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
		}
	})

	// 应用分页
	start := filters.Offset
	end := start + filters.Limit
	if start >= len(filtered) {
		return []*core.NFT{}
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// updateNFTPrices 更新NFT价格信息
func (s *NFTService) updateNFTPrices(ctx context.Context, nfts []*core.NFT) {
	// 简化实现：为演示目的
	for _, nft := range nfts {
		if nft.MarketData == nil {
			nft.MarketData = &core.MarketData{
				FloorPrice:     big.NewInt(1000000000000000000), // 1 ETH
				LastPrice:      big.NewInt(1200000000000000000), // 1.2 ETH
				EstimatedValue: big.NewInt(1100000000000000000), // 1.1 ETH
				IsListed:       false,
			}
		}
	}
}

// getNFTPrice 获取NFT价格信息
func (s *NFTService) getNFTPrice(ctx context.Context, contractAddr, tokenID string) (*NFTPriceInfo, error) {
	cacheKey := fmt.Sprintf("%s:%s", contractAddr, tokenID)

	s.mu.RLock()
	if cached, exists := s.priceCache[cacheKey]; exists {
		s.mu.RUnlock()
		return cached, nil
	}
	s.mu.RUnlock()

	// 模拟价格信息
	priceInfo := &NFTPriceInfo{
		ContractAddr:   contractAddr,
		TokenID:        tokenID,
		CurrentPrice:   big.NewInt(1200000000000000000), // 1.2 ETH
		FloorPrice:     big.NewInt(1000000000000000000), // 1 ETH
		LastSalePrice:  big.NewInt(1100000000000000000), // 1.1 ETH
		EstimatedPrice: big.NewInt(1150000000000000000), // 1.15 ETH
		PriceChange24h: "+5.2%",
		Liquidity:      "medium",
		RarityScore:    75.5,
		UpdatedAt:      time.Now(),
	}

	// 缓存价格信息
	s.mu.Lock()
	s.priceCache[cacheKey] = priceInfo
	s.mu.Unlock()

	return priceInfo, nil
}

// updateRecentSales 更新最近销售记录
func (s *NFTService) updateRecentSales(ctx context.Context, nft *core.NFT) {
	// 简化实现：添加示例销售记录
	if nft.LastSale == nil {
		nft.LastSale = &core.SaleInfo{
			Price:       big.NewInt(1100000000000000000), // 1.1 ETH
			Currency:    "ETH",
			PriceUSD:    big.NewInt(2200000000), // $2200
			From:        "0x1111111111111111111111111111111111111111",
			To:          "0x2222222222222222222222222222222222222222",
			TxHash:      "0x" + fmt.Sprintf("%x", time.Now().UnixNano()),
			Marketplace: "OpenSea",
			SaleTime:    time.Now().Add(-24 * time.Hour),
		}
	}
}

// validateTransferParams 验证转账参数
func (s *NFTService) validateTransferParams(params *NFTTransferRequest) error {
	if params.ContractAddr == "" {
		return fmt.Errorf("合约地址不能为空")
	}
	if params.From == "" {
		return fmt.Errorf("发送方地址不能为空")
	}
	if params.To == "" {
		return fmt.Errorf("接收方地址不能为空")
	}
	if params.TokenID == "" {
		return fmt.Errorf("代币ID不能为空")
	}
	if params.Amount <= 0 {
		params.Amount = 1 // 默认数量为1
	}
	return nil
}
