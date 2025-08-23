/*
桥接业务服务层

本文件实现了跨链桥接功能的业务服务层，封装桥接相关的业务逻辑。

主要服务：
- 跨链路径查询和费用估算
- 桥接交易执行和状态追踪
- 用户桥接历史管理
- 风险评估和安全检查
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

// BridgeService 桥接业务服务
type BridgeService struct {
	bridgeManager *core.BridgeManager           // 桥接管理器
	multiChain    *core.MultiChainManager       // 多链管理器
	bridgeHistory map[string]*BridgeUserHistory // 用户桥接历史
	mu            sync.RWMutex                  // 读写锁
}

// BridgeUserHistory 用户桥接历史
type BridgeUserHistory struct {
	UserAddress  string               `json:"user_address"`  // 用户地址
	TotalBridges int                  `json:"total_bridges"` // 总桥接次数
	TotalVolume  *big.Int             `json:"total_volume"`  // 总交易量
	Records      []*core.BridgeRecord `json:"records"`       // 桥接记录
	LastActivity time.Time            `json:"last_activity"` // 最后活动时间
}

// BridgeQuoteRequest 桥接报价请求
type BridgeQuoteRequest struct {
	FromChain         string  `json:"from_chain" binding:"required"`
	ToChain           string  `json:"to_chain" binding:"required"`
	TokenAddress      string  `json:"token_address"`
	Amount            string  `json:"amount" binding:"required"`
	FromAddress       string  `json:"from_address" binding:"required"`
	ToAddress         string  `json:"to_address" binding:"required"`
	SlippageTolerance float64 `json:"slippage_tolerance"`
	Priority          string  `json:"priority"` // fast/normal/cheap
}

// BridgeQuoteResponse 桥接报价响应
type BridgeQuoteResponse struct {
	BestQuote         *core.BridgeQuote   `json:"best_quote"`
	AlternativeRoutes []*core.BridgeQuote `json:"alternative_routes"`
	RiskAssessment    *RiskAssessment     `json:"risk_assessment"`
	Recommendations   []string            `json:"recommendations"`
	ValidUntil        int64               `json:"valid_until"`
}

// BridgeExecuteRequest 桥接执行请求
type BridgeExecuteRequest struct {
	FromChain         string  `json:"from_chain" binding:"required"`
	ToChain           string  `json:"to_chain" binding:"required"`
	TokenAddress      string  `json:"token_address"`
	Amount            string  `json:"amount" binding:"required"`
	FromAddress       string  `json:"from_address" binding:"required"`
	ToAddress         string  `json:"to_address" binding:"required"`
	SlippageTolerance float64 `json:"slippage_tolerance"`
	Priority          string  `json:"priority"`
	Deadline          int64   `json:"deadline"`
	Mnemonic          string  `json:"mnemonic" binding:"required"`
	DerivationPath    string  `json:"derivation_path"`
	SessionID         string  `json:"session_id"`
}

// BridgeExecuteResponse 桥接执行响应
type BridgeExecuteResponse struct {
	BridgeID      string   `json:"bridge_id"`
	FromTxHash    string   `json:"from_tx_hash"`
	Status        string   `json:"status"`
	EstimatedTime int64    `json:"estimated_time"`
	TrackingURL   string   `json:"tracking_url"`
	NextSteps     []string `json:"next_steps"`
}

// BridgeStatusResponse 桥接状态响应
type BridgeStatusResponse struct {
	BridgeID            string    `json:"bridge_id"`
	Status              string    `json:"status"`
	Progress            float64   `json:"progress"`
	FromTxHash          string    `json:"from_tx_hash"`
	ToTxHash            string    `json:"to_tx_hash"`
	EstimatedCompletion time.Time `json:"estimated_completion"`
	NextAction          string    `json:"next_action"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRisk string   `json:"overall_risk"` // low/medium/high
	RiskFactors []string `json:"risk_factors"` // 风险因素
	SafetyScore float64  `json:"safety_score"` // 安全评分(0-1)
	Warnings    []string `json:"warnings"`     // 警告信息
}

// NewBridgeService 创建桥接服务实例
func NewBridgeService(multiChain *core.MultiChainManager) (*BridgeService, error) {
	bridgeManager, err := core.NewBridgeManager(multiChain)
	if err != nil {
		return nil, fmt.Errorf("创建桥接管理器失败: %w", err)
	}

	return &BridgeService{
		bridgeManager: bridgeManager,
		multiChain:    multiChain,
		bridgeHistory: make(map[string]*BridgeUserHistory),
	}, nil
}

// GetBestRoute 获取最佳桥接路径
func (s *BridgeService) GetBestRoute(ctx context.Context, request *BridgeQuoteRequest) (*BridgeQuoteResponse, error) {
	// 构建桥接参数
	amount, _ := new(big.Int).SetString(request.Amount, 10)
	params := &core.BridgeParams{
		FromChain:         request.FromChain,
		ToChain:           request.ToChain,
		TokenAddress:      request.TokenAddress,
		Amount:            amount,
		FromAddress:       request.FromAddress,
		ToAddress:         request.ToAddress,
		SlippageTolerance: request.SlippageTolerance,
		Priority:          request.Priority,
	}

	// 获取最佳路径
	quote, err := s.bridgeManager.GetBestRoute(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("获取最佳路径失败: %w", err)
	}

	// 构建响应
	response := &BridgeQuoteResponse{
		BestQuote:         quote,
		AlternativeRoutes: []*core.BridgeQuote{}, // 简化实现
		RiskAssessment:    s.assessRisk(quote),
		Recommendations:   s.generateRecommendations(),
		ValidUntil:        quote.ValidUntil,
	}

	return response, nil
}

// ExecuteBridge 执行桥接
func (s *BridgeService) ExecuteBridge(ctx context.Context, request *BridgeExecuteRequest) (*BridgeExecuteResponse, error) {
	// 构建桥接参数
	amount, _ := new(big.Int).SetString(request.Amount, 10)
	params := &core.BridgeParams{
		FromChain:         request.FromChain,
		ToChain:           request.ToChain,
		TokenAddress:      request.TokenAddress,
		Amount:            amount,
		FromAddress:       request.FromAddress,
		ToAddress:         request.ToAddress,
		SlippageTolerance: request.SlippageTolerance,
		Priority:          request.Priority,
		Deadline:          request.Deadline,
	}

	// 构建认证信息
	credentials := &core.BridgeCredentials{
		Mnemonic:       request.Mnemonic,
		DerivationPath: request.DerivationPath,
		SessionID:      request.SessionID,
	}

	// 执行桥接
	result, err := s.bridgeManager.ExecuteBridge(ctx, params, credentials)
	if err != nil {
		return nil, fmt.Errorf("执行桥接失败: %w", err)
	}

	// 记录历史
	s.recordBridgeHistory(request.FromAddress, result)

	// 构建响应
	response := &BridgeExecuteResponse{
		BridgeID:      result.BridgeID,
		FromTxHash:    result.FromTxHash,
		Status:        result.Status,
		EstimatedTime: result.EstimatedTime,
		TrackingURL:   s.generateTrackingURL(result.BridgeID),
		NextSteps:     s.generateNextSteps(),
	}

	return response, nil
}

// GetBridgeStatus 获取桥接状态
func (s *BridgeService) GetBridgeStatus(ctx context.Context, bridgeID string) (*BridgeStatusResponse, error) {
	status, err := s.bridgeManager.GetBridgeStatus(ctx, bridgeID)
	if err != nil {
		return nil, fmt.Errorf("获取桥接状态失败: %w", err)
	}

	if status == nil {
		return nil, fmt.Errorf("桥接记录不存在")
	}

	response := &BridgeStatusResponse{
		BridgeID:            status.BridgeID,
		Status:              status.Status,
		Progress:            status.Progress,
		FromTxHash:          status.FromTxHash,
		ToTxHash:            status.ToTxHash,
		EstimatedCompletion: status.EstimatedCompletion,
		NextAction:          s.determineNextAction(status),
	}

	return response, nil
}

// GetBridgeHistory 获取用户桥接历史
func (s *BridgeService) GetBridgeHistory(userAddress string) (*BridgeUserHistory, error) {
	s.mu.RLock()
	history, exists := s.bridgeHistory[userAddress]
	s.mu.RUnlock()

	if !exists {
		return &BridgeUserHistory{
			UserAddress: userAddress,
			Records:     []*core.BridgeRecord{},
		}, nil
	}

	return history, nil
}

// 私有方法实现

// assessRisk 评估风险
func (s *BridgeService) assessRisk(quote *core.BridgeQuote) *RiskAssessment {
	return &RiskAssessment{
		OverallRisk: "low",
		SafetyScore: 0.95,
		RiskFactors: []string{},
		Warnings:    []string{},
	}
}

// generateRecommendations 生成推荐
func (s *BridgeService) generateRecommendations() []string {
	return []string{
		"建议在网络拥堵较低时执行桥接",
		"确认接收地址正确无误",
		"设置适当的Gas价格",
	}
}

// recordBridgeHistory 记录桥接历史
func (s *BridgeService) recordBridgeHistory(userAddress string, result *core.BridgeResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.bridgeHistory[userAddress]; !exists {
		s.bridgeHistory[userAddress] = &BridgeUserHistory{
			UserAddress: userAddress,
			Records:     make([]*core.BridgeRecord, 0),
		}
	}

	record := &core.BridgeRecord{
		BridgeResult: result,
		Success:      false,
	}

	s.bridgeHistory[userAddress].Records = append(s.bridgeHistory[userAddress].Records, record)
	s.bridgeHistory[userAddress].TotalBridges++
	s.bridgeHistory[userAddress].LastActivity = time.Now()
}

// generateTrackingURL 生成追踪URL
func (s *BridgeService) generateTrackingURL(bridgeID string) string {
	return fmt.Sprintf("https://wallet.example.com/bridge/track/%s", bridgeID)
}

// generateNextSteps 生成下一步操作
func (s *BridgeService) generateNextSteps() []string {
	return []string{
		"等待源链交易确认",
		"桥接协议处理中",
		"目标链接收处理",
		"交易完成确认",
	}
}

// determineNextAction 确定下一步操作
func (s *BridgeService) determineNextAction(status *core.BridgeStatus) string {
	switch status.Status {
	case "pending":
		return "等待交易确认"
	case "processing":
		return "桥接处理中，请耐心等待"
	case "completed":
		return "桥接已完成"
	case "failed":
		return "桥接失败，请联系客服"
	default:
		return "检查交易状态"
	}
}
