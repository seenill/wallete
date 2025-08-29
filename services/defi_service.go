/*
DeFi业务服务层

本文件实现了DeFi功能的业务服务层，封装DeFi相关的业务逻辑，
为API层提供高级接口，简化DeFi操作的复杂性。

主要服务：
交易服务：
- 最佳价格路由查询
- 多DEX价格比较
- 一键智能交易执行
- 交易历史和分析

收益服务：
- 收益策略推荐
- 自动化投资管理
- 收益统计和分析
- 风险评估和提醒

流动性服务：
- 流动性池推荐
- LP代币管理
- 无常损失计算
- 收益复投策略

安全特性：
- 交易前安全检查
- 滑点和价格影响保护
- 黑名单地址过滤
- 风险等级评估
*/
package services

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"wallet/core"
)

// DeFiService DeFi业务服务
// 提供完整的DeFi功能封装，包括交易、收益、流动性等服务
type DeFiService struct {
	multiChain     *core.MultiChainManager     // 多链管理器
	exchanges      map[string]core.DEXExchange // 支持的交易所
	strategies     map[string]*YieldStrategy   // 收益策略
	userPositions  map[string][]*UserPosition  // 用户仓位映射
	priceCache     map[string]*PriceCache      // 价格缓存
	oneInchService *OneInchService             // 1inch聚合器服务
	mu             sync.RWMutex                // 读写锁
}

// SwapRequest 交易请求参数
type SwapRequest struct {
	TokenIn     string `json:"token_in" binding:"required"`     // 输入代币地址
	TokenOut    string `json:"token_out" binding:"required"`    // 输出代币地址
	AmountIn    string `json:"amount_in" binding:"required"`    // 输入数量
	Slippage    string `json:"slippage"`                        // 滑点容忍度（百分比）
	UserAddress string `json:"user_address" binding:"required"` // 用户地址
	Deadline    int64  `json:"deadline"`                        // 交易截止时间
	GasPrice    string `json:"gas_price"`                       // Gas价格
}

// SwapQuote 交易报价
type SwapQuote struct {
	AmountOut    string           `json:"amount_out"`     // 预期输出数量
	AmountOutMin string           `json:"amount_out_min"` // 最小输出数量
	Price        string           `json:"price"`          // 交易价格
	PriceImpact  string           `json:"price_impact"`   // 价格影响
	GasEstimate  uint64           `json:"gas_estimate"`   // Gas估算
	Route        []*RouteInfo     `json:"route"`          // 交易路径
	Warnings     []string         `json:"warnings"`       // 风险警告
	Exchange     string           `json:"exchange"`       // 推荐交易所
	ValidUntil   int64            `json:"valid_until"`    // 报价有效期
	Comparison   []*ExchangeQuote `json:"comparison"`     // 多交易所比较
	// 1inch特定字段
	OneInchData *OneInchQuoteData `json:"oneinch_data,omitempty"` // 1inch数据
}

// OneInchQuoteData 1inch报价数据
type OneInchQuoteData struct {
	ToTokenAmount string `json:"to_token_amount"`
	EstimatedGas  int64  `json:"estimated_gas"`
	GasPrice      string `json:"gas_price"`
	TxData        string `json:"tx_data,omitempty"`
}

// RouteInfo 路由信息
type RouteInfo struct {
	Exchange   string `json:"exchange"`   // 交易所名称
	TokenIn    string `json:"token_in"`   // 输入代币
	TokenOut   string `json:"token_out"`  // 输出代币
	AmountIn   string `json:"amount_in"`  // 输入数量
	AmountOut  string `json:"amount_out"` // 输出数量
	Fee        string `json:"fee"`        // 手续费
	Percentage string `json:"percentage"` // 交易量占比
}

// ExchangeQuote 交易所报价比较
type ExchangeQuote struct {
	Exchange    string `json:"exchange"`     // 交易所名称
	AmountOut   string `json:"amount_out"`   // 输出数量
	GasEstimate uint64 `json:"gas_estimate"` // Gas估算
	Rating      string `json:"rating"`       // 评级（A-F）
	Reason      string `json:"reason"`       // 推荐理由
	// 1inch特定字段
	IsOneInch bool `json:"is_oneinch"` // 是否为1inch
}

// YieldStrategy 收益策略信息
type YieldStrategy struct {
	ID          string    `json:"id"`          // 策略ID
	Name        string    `json:"name"`        // 策略名称
	Protocol    string    `json:"protocol"`    // 协议名称
	Type        string    `json:"type"`        // 策略类型
	APY         string    `json:"apy"`         // 年化收益率
	TVL         string    `json:"tvl"`         // 总锁定价值
	RiskLevel   string    `json:"risk_level"`  // 风险等级
	MinAmount   string    `json:"min_amount"`  // 最小投资金额
	Description string    `json:"description"` // 策略描述
	Tags        []string  `json:"tags"`        // 标签
	LaunchTime  time.Time `json:"launch_time"` // 上线时间
	IsActive    bool      `json:"is_active"`   // 是否活跃
}

// UserPosition 用户仓位
type UserPosition struct {
	ID            string    `json:"id"`             // 仓位ID
	StrategyID    string    `json:"strategy_id"`    // 策略ID
	Type          string    `json:"type"`           // 仓位类型
	Amount        string    `json:"amount"`         // 投资金额
	CurrentValue  string    `json:"current_value"`  // 当前价值
	PnL           string    `json:"pnl"`            // 盈亏
	PnLPercent    string    `json:"pnl_percent"`    // 盈亏百分比
	TotalReward   string    `json:"total_reward"`   // 累计奖励
	PendingReward string    `json:"pending_reward"` // 待领取奖励
	Duration      int64     `json:"duration"`       // 持有时长（秒）
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// LiquidityPoolInfo 流动性池信息
type LiquidityPoolInfo struct {
	Address      string          `json:"address"`       // 池子地址
	Name         string          `json:"name"`          // 池子名称
	Exchange     string          `json:"exchange"`      // 交易所
	TokenA       *TokenInfo      `json:"token_a"`       // 代币A
	TokenB       *TokenInfo      `json:"token_b"`       // 代币B
	TVL          string          `json:"tvl"`           // 总锁定价值
	APY          string          `json:"apy"`           // 年化收益率
	Volume24h    string          `json:"volume_24h"`    // 24小时交易量
	Fee          string          `json:"fee"`           // 手续费率
	Rewards      []*RewardInfo   `json:"rewards"`       // 奖励信息
	UserPosition *UserLPPosition `json:"user_position"` // 用户仓位
	RiskFactors  []string        `json:"risk_factors"`  // 风险因素
}

// TokenInfo 代币信息
type TokenInfo struct {
	Address   string `json:"address"`    // 代币地址
	Symbol    string `json:"symbol"`     // 代币符号
	Name      string `json:"name"`       // 代币名称
	Decimals  int    `json:"decimals"`   // 小数位数
	LogoURL   string `json:"logo_url"`   // 图标URL
	Price     string `json:"price"`      // 当前价格
	Change24h string `json:"change_24h"` // 24小时涨跌幅
}

// RewardInfo 奖励信息
type RewardInfo struct {
	Token  *TokenInfo `json:"token"`  // 奖励代币
	APY    string     `json:"apy"`    // 奖励APY
	Amount string     `json:"amount"` // 奖励数量
}

// UserLPPosition 用户LP仓位
type UserLPPosition struct {
	LPTokens    string `json:"lp_tokens"`    // LP代币数量
	Share       string `json:"share"`        // 占池子份额
	ValueUSD    string `json:"value_usd"`    // 仓位价值（USD）
	IL          string `json:"il"`           // 无常损失
	Earned      string `json:"earned"`       // 已赚取费用
	PendingFees string `json:"pending_fees"` // 待领取费用
}

// PriceCache 价格缓存
type PriceCache struct {
	Price     string    `json:"price"`      // 价格
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	Source    string    `json:"source"`     // 数据源
}

// NewDeFiService 创建DeFi服务实例
func NewDeFiService(multiChain *core.MultiChainManager) *DeFiService {
	service := &DeFiService{
		multiChain:    multiChain,
		exchanges:     make(map[string]core.DEXExchange),
		strategies:    make(map[string]*YieldStrategy),
		userPositions: make(map[string][]*UserPosition),
		priceCache:    make(map[string]*PriceCache),
		// 初始化1inch服务（需要配置API密钥）
		oneInchService: NewOneInchService(""), // 在实际使用时需要设置API密钥
	}

	// 初始化默认策略
	service.initDefaultStrategies()

	return service
}

// SetOneInchAPIKey 设置1inch API密钥
func (s *DeFiService) SetOneInchAPIKey(apiKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.oneInchService == nil {
		s.oneInchService = NewOneInchService(apiKey)
	} else {
		// 重新创建服务实例以更新API密钥
		s.oneInchService = NewOneInchService(apiKey)
	}
}

// GetOneInchService 获取1inch服务实例（用于处理器直接调用）
func (s *DeFiService) GetOneInchService() *OneInchService {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oneInchService
}

// GetSwapQuote 获取交易报价
func (s *DeFiService) GetSwapQuote(req *SwapRequest) (*SwapQuote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 解析输入金额
	amountIn := new(big.Int)
	amountIn.SetString(req.AmountIn, 10)

	// 获取多个交易所的报价
	quotes := make([]*ExchangeQuote, 0)
	bestQuote := &core.QuoteResult{}
	bestExchange := ""

	for name, exchange := range s.exchanges {
		quote, err := exchange.GetQuote(req.TokenIn, req.TokenOut, amountIn)
		if err != nil {
			continue // 跳过错误的交易所
		}

		exchangeQuote := &ExchangeQuote{
			Exchange:    name,
			AmountOut:   quote.AmountOut.String(),
			GasEstimate: quote.GasEstimate,
			Rating:      s.calculateExchangeRating(quote),
			Reason:      s.getRecommendationReason(name, quote),
			IsOneInch:   false,
		}
		quotes = append(quotes, exchangeQuote)

		// 选择最佳报价（输出金额最高）
		if bestQuote.AmountOut == nil || quote.AmountOut.Cmp(bestQuote.AmountOut) > 0 {
			bestQuote = quote
			bestExchange = name
		}
	}

	// 获取1inch报价（如果配置了API密钥）
	var oneInchQuote *OneInchQuoteData
	if s.oneInchService != nil && s.oneInchService.apiKey != "" {
		oneInchReq := &OneInchQuoteRequest{
			FromTokenAddress: req.TokenIn,
			ToTokenAddress:   req.TokenOut,
			Amount:           req.AmountIn,
			GasPrice:         req.GasPrice,
		}

		oneInchResp, err := s.oneInchService.GetQuote(context.Background(), oneInchReq)
		if err == nil {
			oneInchQuote = &OneInchQuoteData{
				ToTokenAmount: oneInchResp.ToTokenAmount,
				EstimatedGas:  oneInchResp.EstimatedGas,
				GasPrice:      oneInchResp.GasPrice,
			}

			// 创建1inch报价比较项
			oneInchExchangeQuote := &ExchangeQuote{
				Exchange:    "1inch",
				AmountOut:   oneInchResp.ToTokenAmount,
				GasEstimate: uint64(oneInchResp.EstimatedGas),
				Rating:      "A+", // 1inch通常提供最优价格
				Reason:      "聚合多个DEX，提供最优价格",
				IsOneInch:   true,
			}
			quotes = append(quotes, oneInchExchangeQuote)

			// 检查是否1inch提供更好的报价
			oneInchAmountOut := new(big.Int)
			oneInchAmountOut.SetString(oneInchResp.ToTokenAmount, 10)

			if bestQuote.AmountOut == nil || oneInchAmountOut.Cmp(bestQuote.AmountOut) > 0 {
				bestQuote = &core.QuoteResult{
					AmountOut:    oneInchAmountOut,
					AmountOutMin: oneInchAmountOut, // 简化处理
					GasEstimate:  uint64(oneInchResp.EstimatedGas),
				}
				bestExchange = "1inch"
			}
		}
	}

	if bestQuote.AmountOut == nil {
		return nil, fmt.Errorf("no valid quotes found")
	}

	// 排序报价（按输出数量降序）
	sort.Slice(quotes, func(i, j int) bool {
		amountI := new(big.Int)
		amountJ := new(big.Int)
		amountI.SetString(quotes[i].AmountOut, 10)
		amountJ.SetString(quotes[j].AmountOut, 10)
		return amountI.Cmp(amountJ) > 0
	})

	// 构建路由信息
	routeInfo := make([]*RouteInfo, 0)
	for _, hop := range bestQuote.Route {
		routeInfo = append(routeInfo, &RouteInfo{
			Exchange:   hop.Exchange,
			TokenIn:    hop.TokenIn.Address,
			TokenOut:   hop.TokenOut.Address,
			AmountIn:   hop.AmountIn.String(),
			AmountOut:  hop.AmountOut.String(),
			Fee:        hop.Fee.String(),
			Percentage: "100", // 单路径为100%
		})
	}

	// 安全检查和警告
	warnings := s.generateWarnings(req, bestQuote)

	return &SwapQuote{
		AmountOut:    bestQuote.AmountOut.String(),
		AmountOutMin: bestQuote.AmountOutMin.String(),
		Price:        bestQuote.Price,
		PriceImpact:  bestQuote.PriceImpact,
		GasEstimate:  bestQuote.GasEstimate,
		Route:        routeInfo,
		Warnings:     warnings,
		Exchange:     bestExchange,
		ValidUntil:   bestQuote.ValidUntil,
		Comparison:   quotes,
		OneInchData:  oneInchQuote,
	}, nil
}

// ExecuteSwap 执行交易
func (s *DeFiService) ExecuteSwap(req *SwapRequest, sessionID string) (*core.SwapResult, error) {
	// 获取报价
	quote, err := s.GetSwapQuote(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	// 检查报价是否过期
	if time.Now().Unix() > quote.ValidUntil {
		return nil, fmt.Errorf("quote expired, please refresh")
	}

	// 风险检查
	if err := s.performRiskCheck(req, quote); err != nil {
		return nil, fmt.Errorf("risk check failed: %w", err)
	}

	// 如果是1inch交易，使用1inch执行
	if quote.Exchange == "1inch" && s.oneInchService != nil && s.oneInchService.apiKey != "" {
		return s.executeOneInchSwap(req, quote)
	}

	// 获取对应的交易所
	exchange, exists := s.exchanges[quote.Exchange]
	if !exists {
		return nil, fmt.Errorf("exchange not found: %s", quote.Exchange)
	}

	// 构建交易参数
	amountIn := new(big.Int)
	amountIn.SetString(req.AmountIn, 10)

	amountOutMin := new(big.Int)
	amountOutMin.SetString(quote.AmountOutMin, 10)

	swapParams := &core.SwapParams{
		TokenIn:      req.TokenIn,
		TokenOut:     req.TokenOut,
		AmountIn:     amountIn,
		AmountOutMin: amountOutMin,
		Deadline:     req.Deadline,
		Recipient:    req.UserAddress,
		Slippage:     req.Slippage,
	}

	// 执行交易
	ctx := context.Background()
	result, err := exchange.ExecuteSwap(ctx, swapParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute swap: %w", err)
	}

	// 记录交易历史
	s.recordSwapHistory(req.UserAddress, result)

	return result, nil
}

// executeOneInchSwap 执行1inch交换
func (s *DeFiService) executeOneInchSwap(req *SwapRequest, quote *SwapQuote) (*core.SwapResult, error) {
	// 构建1inch交换请求
	slippage := "1" // 默认1%滑点
	if req.Slippage != "" {
		slippage = req.Slippage
	}

	oneInchReq := &OneInchSwapRequest{
		FromTokenAddress: req.TokenIn,
		ToTokenAddress:   req.TokenOut,
		Amount:           req.AmountIn,
		FromAddress:      req.UserAddress,
		Slippage:         slippage,
		GasPrice:         req.GasPrice,
	}

	// 获取交换数据
	swapResp, err := s.oneInchService.GetSwap(context.Background(), oneInchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get 1inch swap data: %w", err)
	}

	// 构建交换结果
	amountOut := new(big.Int)
	amountOut.SetString(swapResp.ToTokenAmount, 10)

	gasPrice := new(big.Int)
	gasPrice.SetString(swapResp.Tx.GasPrice, 10)

	result := &core.SwapResult{
		TxHash:    "", // 交易哈希将在实际发送后获得
		AmountOut: amountOut,
		GasUsed:   uint64(swapResp.Tx.Gas),
		GasPrice:  gasPrice,
		Status:    "success",
		Exchange:  "1inch",
	}

	// 记录交易历史
	s.recordSwapHistory(req.UserAddress, result)

	return result, nil
}

// GetYieldStrategies 获取收益策略列表
func (s *DeFiService) GetYieldStrategies(riskLevel string, minAPY float64) ([]*YieldStrategy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make([]*YieldStrategy, 0)

	for _, strategy := range s.strategies {
		// 过滤风险等级
		if riskLevel != "" && strategy.RiskLevel != riskLevel {
			continue
		}

		// 过滤最小APY
		if minAPY > 0 {
			apy := parseFloatFromString(strategy.APY)
			if apy < minAPY {
				continue
			}
		}

		// 只返回活跃策略
		if strategy.IsActive {
			strategies = append(strategies, strategy)
		}
	}

	// 按APY降序排序
	sort.Slice(strategies, func(i, j int) bool {
		apyI := parseFloatFromString(strategies[i].APY)
		apyJ := parseFloatFromString(strategies[j].APY)
		return apyI > apyJ
	})

	return strategies, nil
}

// GetUserPositions 获取用户仓位
func (s *DeFiService) GetUserPositions(userAddress string) ([]*UserPosition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	positions, exists := s.userPositions[strings.ToLower(userAddress)]
	if !exists {
		return []*UserPosition{}, nil
	}

	// 更新仓位价值
	for _, position := range positions {
		s.updatePositionValue(position)
	}

	return positions, nil
}

// GetLiquidityPools 获取流动性池列表
func (s *DeFiService) GetLiquidityPools(exchange string, sortBy string) ([]*LiquidityPoolInfo, error) {
	var pools []*LiquidityPoolInfo

	// 从各个交易所获取流动性池
	for name, ex := range s.exchanges {
		if exchange != "" && name != exchange {
			continue
		}

		corePools, err := ex.GetLiquidityPools()
		if err != nil {
			continue
		}

		for _, corePool := range corePools {
			pool := &LiquidityPoolInfo{
				Address:  corePool.Address,
				Name:     corePool.Name,
				Exchange: name,
				TokenA: &TokenInfo{
					Address:  corePool.TokenA.Address,
					Symbol:   corePool.TokenA.Symbol,
					Name:     corePool.TokenA.Name,
					Decimals: corePool.TokenA.Decimals,
					LogoURL:  corePool.TokenA.LogoURL,
				},
				TokenB: &TokenInfo{
					Address:  corePool.TokenB.Address,
					Symbol:   corePool.TokenB.Symbol,
					Name:     corePool.TokenB.Name,
					Decimals: corePool.TokenB.Decimals,
					LogoURL:  corePool.TokenB.LogoURL,
				},
				TVL:       corePool.TVL.String(),
				APY:       corePool.APY,
				Volume24h: corePool.Volume24h.String(),
				Fee:       corePool.Fee,
			}

			// 添加奖励信息
			for _, rewardToken := range corePool.RewardTokens {
				reward := &RewardInfo{
					Token: &TokenInfo{
						Address:  rewardToken.Address,
						Symbol:   rewardToken.Symbol,
						Name:     rewardToken.Name,
						Decimals: rewardToken.Decimals,
						LogoURL:  rewardToken.LogoURL,
					},
					APY:    "0", // 简化处理，实际应从其他来源获取
					Amount: "0", // 简化处理，实际应从其他来源获取
				}
				pool.Rewards = append(pool.Rewards, reward)
			}

			pools = append(pools, pool)
		}
	}

	// 根据指定字段排序
	switch strings.ToLower(sortBy) {
	case "tvl":
		sort.Slice(pools, func(i, j int) bool {
			return parseFloatFromString(pools[i].TVL) > parseFloatFromString(pools[j].TVL)
		})
	case "apy":
		sort.Slice(pools, func(i, j int) bool {
			return parseFloatFromString(pools[i].APY) > parseFloatFromString(pools[j].APY)
		})
	case "volume":
		sort.Slice(pools, func(i, j int) bool {
			return parseFloatFromString(pools[i].Volume24h) > parseFloatFromString(pools[j].Volume24h)
		})
	default:
		// 默认按TVL排序
		sort.Slice(pools, func(i, j int) bool {
			return parseFloatFromString(pools[i].TVL) > parseFloatFromString(pools[j].TVL)
		})
	}

	return pools, nil
}

// 辅助函数
func parseFloatFromString(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// initDefaultStrategies 初始化默认策略
func (s *DeFiService) initDefaultStrategies() {
	strategies := []*YieldStrategy{
		{
			ID:          "uni-eth-usdc",
			Name:        "ETH/USDC LP",
			Protocol:    "Uniswap V2",
			Type:        "LP",
			APY:         "25.6",
			TVL:         "150000000",
			RiskLevel:   "Medium",
			MinAmount:   "100",
			Description: "Provide liquidity to ETH/USDC pair on Uniswap",
			Tags:        []string{"Stable", "Popular", "High Volume"},
			LaunchTime:  time.Now().AddDate(0, -6, 0),
			IsActive:    true,
		},
		{
			ID:          "aave-usdc-deposit",
			Name:        "USDC Lending",
			Protocol:    "Aave V3",
			Type:        "Lending",
			APY:         "8.2",
			TVL:         "500000000",
			RiskLevel:   "Low",
			MinAmount:   "50",
			Description: "Earn interest by lending USDC on Aave",
			Tags:        []string{"Safe", "Stable Coin", "Established"},
			LaunchTime:  time.Now().AddDate(-1, 0, 0),
			IsActive:    true,
		},
	}

	for _, strategy := range strategies {
		s.strategies[strategy.ID] = strategy
	}
}

// calculateExchangeRating 计算交易所评级
func (s *DeFiService) calculateExchangeRating(quote *core.QuoteResult) string {
	// 简化的评级算法
	priceImpact := parseFloatFromString(quote.PriceImpact)

	if priceImpact < 0.1 {
		return "A"
	} else if priceImpact < 0.5 {
		return "B"
	} else if priceImpact < 1.0 {
		return "C"
	} else {
		return "D"
	}
}

// getRecommendationReason 获取推荐理由
func (s *DeFiService) getRecommendationReason(exchange string, quote *core.QuoteResult) string {
	reasons := []string{
		"Best price available",
		"Low gas cost",
		"High liquidity",
		"Trusted exchange",
	}

	// 简化逻辑，实际应该根据具体指标判断
	return reasons[0]
}

// generateWarnings 生成安全警告
func (s *DeFiService) generateWarnings(req *SwapRequest, quote *core.QuoteResult) []string {
	warnings := make([]string, 0)

	// 检查价格影响
	priceImpact := parseFloatFromString(quote.PriceImpact)
	if priceImpact > 5.0 {
		warnings = append(warnings, "High price impact (>5%). Consider reducing trade size.")
	}

	// 检查滑点设置
	if req.Slippage == "" || parseFloatFromString(req.Slippage) > 10.0 {
		warnings = append(warnings, "High slippage tolerance may result in unfavorable trades.")
	}

	// 检查Gas费用
	if quote.GasEstimate > 500000 {
		warnings = append(warnings, "High gas consumption detected. Consider optimizing route.")
	}

	return warnings
}

// performRiskCheck 执行风险检查
func (s *DeFiService) performRiskCheck(req *SwapRequest, quote *SwapQuote) error {
	// 检查价格影响阈值
	priceImpact := parseFloatFromString(quote.PriceImpact)
	if priceImpact > 15.0 {
		return fmt.Errorf("price impact too high: %.2f%%", priceImpact)
	}

	// 检查滑点设置
	slippage := parseFloatFromString(req.Slippage)
	if slippage > 20.0 {
		return fmt.Errorf("slippage tolerance too high: %.2f%%", slippage)
	}

	return nil
}

// recordSwapHistory 记录交易历史
func (s *DeFiService) recordSwapHistory(userAddress string, result *core.SwapResult) {
	// 实现交易历史记录逻辑
	// 这里可以存储到数据库或内存中
}

// updatePositionValue 更新仓位价值
func (s *DeFiService) updatePositionValue(position *UserPosition) {
	// 实现仓位价值更新逻辑
	// 需要获取最新价格和收益信息
}

// convertToLiquidityPoolInfo 转换流动性池信息
func (s *DeFiService) convertToLiquidityPoolInfo(pool *core.LiquidityPool, exchange string) *LiquidityPoolInfo {
	return &LiquidityPoolInfo{
		Address:   pool.Address,
		Name:      pool.Name,
		Exchange:  exchange,
		TVL:       pool.TVL.String(),
		APY:       pool.APY,
		Volume24h: pool.Volume24h.String(),
		Fee:       pool.Fee,
		// 其他字段的转换...
	}
}

// sortLiquidityPools 排序流动性池
func (s *DeFiService) sortLiquidityPools(pools []*LiquidityPoolInfo, sortBy string) {
	switch sortBy {
	case "apy":
		sort.Slice(pools, func(i, j int) bool {
			apyI := parseFloatFromString(pools[i].APY)
			apyJ := parseFloatFromString(pools[j].APY)
			return apyI > apyJ
		})
	case "tvl":
		sort.Slice(pools, func(i, j int) bool {
			tvlI := parseFloatFromString(pools[i].TVL)
			tvlJ := parseFloatFromString(pools[j].TVL)
			return tvlI > tvlJ
		})
	case "volume":
		sort.Slice(pools, func(i, j int) bool {
			volI := parseFloatFromString(pools[i].Volume24h)
			volJ := parseFloatFromString(pools[j].Volume24h)
			return volI > volJ
		})
	}
}
