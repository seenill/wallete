/*
资产管理业务服务层

本文件实现了资产管理功能的业务服务层，提供投资组合分析、收益统计等服务。
*/
package services

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"
	"wallet/core"
)

// AssetManagementService 资产管理服务
type AssetManagementService struct {
	assetManager   *core.AssetManager         // 资产管理器
	multiChain     *core.MultiChainManager    // 多链管理器
	portfolioCache map[string]*PortfolioCache // 投资组合缓存
	mu             sync.RWMutex               // 读写锁
}

// PortfolioCache 投资组合缓存
type PortfolioCache struct {
	Portfolio *core.Portfolio `json:"portfolio"`  // 投资组合数据
	CachedAt  time.Time       `json:"cached_at"`  // 缓存时间
	ExpiresAt time.Time       `json:"expires_at"` // 过期时间
}

// PortfolioRequest 投资组合请求
type PortfolioRequest struct {
	UserAddress   string   `json:"user_address" binding:"required"`
	IncludeChains []string `json:"include_chains"` // 包含的链
	ExcludeChains []string `json:"exclude_chains"` // 排除的链
	IncludeTypes  []string `json:"include_types"`  // 包含的资产类型
	MinValue      string   `json:"min_value"`      // 最小价值过滤
	ForceRefresh  bool     `json:"force_refresh"`  // 强制刷新
}

// PortfolioResponse 投资组合响应
type PortfolioResponse struct {
	Portfolio       *core.Portfolio    `json:"portfolio"`       // 投资组合
	Insights        *PortfolioInsights `json:"insights"`        // 洞察
	Recommendations []string           `json:"recommendations"` // 建议
	LastUpdated     time.Time          `json:"last_updated"`    // 最后更新
}

// PortfolioInsights 投资组合洞察
type PortfolioInsights struct {
	TopPerformers   []*AssetPerformance `json:"top_performers"`   // 最佳表现资产
	WorstPerformers []*AssetPerformance `json:"worst_performers"` // 最差表现资产
	RiskAnalysis    *core.RiskAnalysis  `json:"risk_analysis"`    // 风险分析
	Opportunities   []string            `json:"opportunities"`    // 投资机会
	Warnings        []string            `json:"warnings"`         // 风险警告
}

// AssetPerformance 资产表现
type AssetPerformance struct {
	Symbol      string   `json:"symbol"`      // 资产符号
	Name        string   `json:"name"`        // 资产名称
	Performance float64  `json:"performance"` // 表现百分比
	Value       *big.Int `json:"value"`       // 价值
	Weight      float64  `json:"weight"`      // 权重
}

// NewAssetManagementService 创建资产管理服务
func NewAssetManagementService(multiChain *core.MultiChainManager) *AssetManagementService {
	return &AssetManagementService{
		assetManager:   core.NewAssetManager(),
		multiChain:     multiChain,
		portfolioCache: make(map[string]*PortfolioCache),
	}
}

// GetPortfolio 获取投资组合
func (ams *AssetManagementService) GetPortfolio(ctx context.Context, request *PortfolioRequest) (*PortfolioResponse, error) {
	// 检查缓存
	if !request.ForceRefresh {
		if cached := ams.getCachedPortfolio(request.UserAddress); cached != nil {
			return ams.buildPortfolioResponse(cached), nil
		}
	}

	// 分析投资组合
	portfolio, err := ams.assetManager.AnalyzePortfolio(ctx, request.UserAddress)
	if err != nil {
		return nil, fmt.Errorf("分析投资组合失败: %w", err)
	}

	// 应用过滤器
	portfolio = ams.applyFilters(portfolio, request)

	// 缓存结果
	ams.cachePortfolio(request.UserAddress, portfolio)

	return ams.buildPortfolioResponse(portfolio), nil
}

// GetRiskAnalysis 获取风险分析
func (ams *AssetManagementService) GetRiskAnalysis(ctx context.Context, userAddress string) (*core.RiskAnalysis, error) {
	portfolio, err := ams.assetManager.AnalyzePortfolio(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("获取投资组合失败: %w", err)
	}

	return ams.assetManager.AssessRisk(portfolio)
}

// ComparePerformance 比较表现
func (ams *AssetManagementService) ComparePerformance(ctx context.Context, userAddress, benchmarkName string) (*core.PortfolioComparison, error) {
	portfolio, err := ams.assetManager.AnalyzePortfolio(ctx, userAddress)
	if err != nil {
		return nil, fmt.Errorf("获取投资组合失败: %w", err)
	}

	return ams.assetManager.CompareToBenchmark(portfolio, benchmarkName)
}

// 私有方法

// getCachedPortfolio 获取缓存的投资组合
func (ams *AssetManagementService) getCachedPortfolio(userAddress string) *core.Portfolio {
	ams.mu.RLock()
	defer ams.mu.RUnlock()

	cached, exists := ams.portfolioCache[userAddress]
	if !exists || time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached.Portfolio
}

// cachePortfolio 缓存投资组合
func (ams *AssetManagementService) cachePortfolio(userAddress string, portfolio *core.Portfolio) {
	ams.mu.Lock()
	defer ams.mu.Unlock()

	ams.portfolioCache[userAddress] = &PortfolioCache{
		Portfolio: portfolio,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
}

// applyFilters 应用过滤器
func (ams *AssetManagementService) applyFilters(portfolio *core.Portfolio, request *PortfolioRequest) *core.Portfolio {
	// 简化实现：直接返回原始组合
	return portfolio
}

// buildPortfolioResponse 构建投资组合响应
func (ams *AssetManagementService) buildPortfolioResponse(portfolio *core.Portfolio) *PortfolioResponse {
	insights := ams.generateInsights(portfolio)
	recommendations := ams.generateRecommendations(portfolio)

	return &PortfolioResponse{
		Portfolio:       portfolio,
		Insights:        insights,
		Recommendations: recommendations,
		LastUpdated:     portfolio.LastUpdated,
	}
}

// generateInsights 生成洞察
func (ams *AssetManagementService) generateInsights(portfolio *core.Portfolio) *PortfolioInsights {
	return &PortfolioInsights{
		TopPerformers:   ams.getTopPerformers(portfolio),
		WorstPerformers: ams.getWorstPerformers(portfolio),
		Opportunities:   []string{"考虑增加稳定币配置", "可以关注DeFi收益机会"},
		Warnings:        []string{"部分资产波动较大", "建议分散投资风险"},
	}
}

// getTopPerformers 获取最佳表现资产
func (ams *AssetManagementService) getTopPerformers(portfolio *core.Portfolio) []*AssetPerformance {
	performers := make([]*AssetPerformance, 0)

	for _, holding := range portfolio.Holdings {
		if holding.PnLPercent > 0 {
			performers = append(performers, &AssetPerformance{
				Symbol:      holding.Symbol,
				Name:        holding.Name,
				Performance: holding.PnLPercent,
				Value:       holding.Value,
				Weight:      holding.Weight,
			})
		}
	}

	return performers
}

// getWorstPerformers 获取最差表现资产
func (ams *AssetManagementService) getWorstPerformers(portfolio *core.Portfolio) []*AssetPerformance {
	performers := make([]*AssetPerformance, 0)

	for _, holding := range portfolio.Holdings {
		if holding.PnLPercent < 0 {
			performers = append(performers, &AssetPerformance{
				Symbol:      holding.Symbol,
				Name:        holding.Name,
				Performance: holding.PnLPercent,
				Value:       holding.Value,
				Weight:      holding.Weight,
			})
		}
	}

	return performers
}

// generateRecommendations 生成建议
func (ams *AssetManagementService) generateRecommendations(portfolio *core.Portfolio) []string {
	return []string{
		"建议定期重新平衡投资组合",
		"考虑增加资产多样化",
		"可以设置止盈止损策略",
		"关注市场趋势变化",
	}
}
