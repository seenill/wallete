/*
高级交易业务服务层

本文件实现了高级交易功能的业务服务层，为API层提供便捷的接口。

主要功能：
- 批量交易管理和执行
- 定时交易调度和监控
- 条件交易触发和执行
- 交易策略创建和管理
- 交易分析和统计
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

// AdvancedTradingService 高级交易服务
type AdvancedTradingService struct {
	tradingManager   *core.AdvancedTradingManager // 高级交易管理器
	multiChain       *core.MultiChainManager      // 多链管理器
	activeSchedules  map[string]*ScheduleMonitor  // 活跃调度监控
	conditionChecker *ConditionChecker            // 条件检查器
	strategyEngine   *StrategyEngine              // 策略引擎
	mu               sync.RWMutex                 // 读写锁
}

// ScheduleMonitor 调度监控器
type ScheduleMonitor struct {
	ScheduledTxID    string             `json:"scheduled_tx_id"`   // 定时交易ID
	NextCheck        time.Time          `json:"next_check"`        // 下次检查时间
	IsActive         bool               `json:"is_active"`         // 是否活跃
	LastExecution    *time.Time         `json:"last_execution"`    // 最后执行时间
	ExecutionHistory []*ExecutionRecord `json:"execution_history"` // 执行历史
}

// ExecutionRecord 执行记录
type ExecutionRecord struct {
	ExecutedAt   time.Time `json:"executed_at"`   // 执行时间
	TxHash       string    `json:"tx_hash"`       // 交易哈希
	Status       string    `json:"status"`        // 状态
	ErrorMessage string    `json:"error_message"` // 错误信息
	GasUsed      uint64    `json:"gas_used"`      // Gas使用量
}

// ConditionChecker 条件检查器
type ConditionChecker struct {
	activeConditions map[string]*core.ConditionalTx // 活跃条件交易
	priceFeeds       map[string]*PriceFeed          // 价格数据源
	checkInterval    time.Duration                  // 检查间隔
	stopChan         chan bool                      // 停止信号
	mu               sync.RWMutex                   // 读写锁
}

// PriceFeed 价格数据源
type PriceFeed struct {
	Symbol      string    `json:"symbol"`       // 交易对符号
	Price       *big.Int  `json:"price"`        // 当前价格
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
	Source      string    `json:"source"`       // 数据源
}

// StrategyEngine 策略引擎
type StrategyEngine struct {
	activeStrategies map[string]*core.TradingStrategy // 活跃策略
	executionQueue   []*StrategyExecution             // 执行队列
	performanceData  map[string]*StrategyAnalysis     // 策略分析数据
	mu               sync.RWMutex                     // 读写锁
}

// StrategyExecution 策略执行
type StrategyExecution struct {
	StrategyID  string                 `json:"strategy_id"`  // 策略ID
	Action      string                 `json:"action"`       // 执行动作
	Parameters  map[string]interface{} `json:"parameters"`   // 参数
	ScheduledAt time.Time              `json:"scheduled_at"` // 计划时间
	ExecutedAt  *time.Time             `json:"executed_at"`  // 执行时间
	Status      string                 `json:"status"`       // 状态
}

// StrategyAnalysis 策略分析
type StrategyAnalysis struct {
	StrategyID      string                    `json:"strategy_id"`      // 策略ID
	PerformanceData *core.StrategyPerformance `json:"performance_data"` // 性能数据
	RiskMetrics     *RiskMetrics              `json:"risk_metrics"`     // 风险指标
	Recommendations []string                  `json:"recommendations"`  // 建议
	LastAnalyzedAt  time.Time                 `json:"last_analyzed_at"` // 最后分析时间
}

// RiskMetrics 风险指标
type RiskMetrics struct {
	VaR95        float64   `json:"var_95"`        // 95% VaR
	MaxDrawdown  float64   `json:"max_drawdown"`  // 最大回撤
	Volatility   float64   `json:"volatility"`    // 波动率
	SharpeRatio  float64   `json:"sharpe_ratio"`  // 夏普比率
	SortinoRatio float64   `json:"sortino_ratio"` // 索提诺比率
	CalmarRatio  float64   `json:"calmar_ratio"`  // 卡尔玛比率
	RiskLevel    string    `json:"risk_level"`    // 风险等级
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// 请求和响应结构体

// BatchTransactionRequest 批量交易请求
type BatchTransactionRequest struct {
	Transactions    []*TransactionItem `json:"transactions" binding:"required"`
	ExecutionMode   string             `json:"execution_mode"`   // parallel/sequential/optimized
	GasOptimization bool               `json:"gas_optimization"` // 是否启用Gas优化
	MaxRetries      int                `json:"max_retries"`      // 最大重试次数
	StopOnError     bool               `json:"stop_on_error"`    // 遇错停止
}

// TransactionItem 交易项目
type TransactionItem struct {
	ID           string            `json:"id"`           // 交易ID
	Type         string            `json:"type"`         // 交易类型（transfer/contract_call）
	From         string            `json:"from"`         // 发送方
	To           string            `json:"to"`           // 接收方
	Value        string            `json:"value"`        // 金额
	Data         string            `json:"data"`         // 交易数据（hex）
	GasLimit     uint64            `json:"gas_limit"`    // Gas限制
	Priority     int               `json:"priority"`     // 优先级
	Dependencies []string          `json:"dependencies"` // 依赖的交易ID
	Metadata     map[string]string `json:"metadata"`     // 元数据
}

// ScheduledTransactionRequest 定时交易请求
type ScheduledTransactionRequest struct {
	TxTemplate      *TxTemplateRequest `json:"tx_template" binding:"required"`
	Schedule        *ScheduleRequest   `json:"schedule" binding:"required"`
	MaxExecutions   int                `json:"max_executions"`    // 最大执行次数
	MaxErrors       int                `json:"max_errors"`        // 最大错误次数
	NotifyOnExecute bool               `json:"notify_on_execute"` // 执行时通知
	ExpiresAt       *time.Time         `json:"expires_at"`        // 过期时间
}

// TxTemplateRequest 交易模板请求
type TxTemplateRequest struct {
	Type         string `json:"type"`          // 交易类型
	From         string `json:"from"`          // 发送方
	To           string `json:"to"`            // 接收方
	Value        string `json:"value"`         // 金额
	TokenAddress string `json:"token_address"` // 代币地址
	Amount       string `json:"amount"`        // 代币数量
	Data         string `json:"data"`          // 交易数据
}

// ScheduleRequest 执行计划请求
type ScheduleRequest struct {
	Type           string     `json:"type"`            // once/recurring/cron
	StartTime      time.Time  `json:"start_time"`      // 开始时间
	EndTime        *time.Time `json:"end_time"`        // 结束时间
	Interval       string     `json:"interval"`        // 间隔（如：1h, 24h, 7d）
	CronExpression string     `json:"cron_expression"` // Cron表达式
	TimeZone       string     `json:"time_zone"`       // 时区
}

// ConditionalTransactionRequest 条件交易请求
type ConditionalTransactionRequest struct {
	TxTemplate      *TxTemplateRequest  `json:"tx_template" binding:"required"`
	Conditions      []*ConditionRequest `json:"conditions" binding:"required"`
	LogicOperator   string              `json:"logic_operator"`    // AND/OR
	ExpiresAt       *time.Time          `json:"expires_at"`        // 过期时间
	NotifyOnTrigger bool                `json:"notify_on_trigger"` // 触发时通知
}

// ConditionRequest 条件请求
type ConditionRequest struct {
	Type          string      `json:"type" binding:"required"`     // price/time/event/balance
	Target        string      `json:"target" binding:"required"`   // 目标对象
	Operator      string      `json:"operator" binding:"required"` // 操作符
	Value         interface{} `json:"value" binding:"required"`    // 比较值
	CheckInterval string      `json:"check_interval"`              // 检查间隔
}

// TradingStrategyRequest 交易策略请求
type TradingStrategyRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Type      string                 `json:"type" binding:"required"` // dca/grid/stop_loss
	Config    map[string]interface{} `json:"config" binding:"required"`
	StartTime *time.Time             `json:"start_time"`
	EndTime   *time.Time             `json:"end_time"`
	AutoStart bool                   `json:"auto_start"`
}

// NewAdvancedTradingService 创建高级交易服务
func NewAdvancedTradingService(multiChain *core.MultiChainManager) (*AdvancedTradingService, error) {
	tradingManager, err := core.NewAdvancedTradingManager(multiChain)
	if err != nil {
		return nil, fmt.Errorf("创建高级交易管理器失败: %w", err)
	}

	service := &AdvancedTradingService{
		tradingManager:   tradingManager,
		multiChain:       multiChain,
		activeSchedules:  make(map[string]*ScheduleMonitor),
		conditionChecker: NewConditionChecker(),
		strategyEngine:   NewStrategyEngine(),
	}

	// 启动后台服务
	service.startBackgroundServices()

	return service, nil
}

// ExecuteBatchTransactions 执行批量交易
func (ats *AdvancedTradingService) ExecuteBatchTransactions(ctx context.Context, request *BatchTransactionRequest) (*core.BatchExecutionResult, error) {
	// 转换请求格式
	batchRequest := &core.BatchExecutionRequest{
		Transactions:    make([]*core.BatchTransaction, len(request.Transactions)),
		ExecutionMode:   request.ExecutionMode,
		MaxRetries:      request.MaxRetries,
		StopOnError:     request.StopOnError,
		GasOptimization: request.GasOptimization,
	}

	for i, txItem := range request.Transactions {
		value, _ := new(big.Int).SetString(txItem.Value, 10)
		batchRequest.Transactions[i] = &core.BatchTransaction{
			ID:        txItem.ID,
			From:      txItem.From,
			To:        txItem.To,
			Value:     value,
			Priority:  txItem.Priority,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
	}

	return ats.tradingManager.ExecuteBatch(ctx, batchRequest)
}

// CreateScheduledTransaction 创建定时交易
func (ats *AdvancedTradingService) CreateScheduledTransaction(request *ScheduledTransactionRequest) (*core.ScheduledTx, error) {
	scheduledTx := &core.ScheduledTx{
		ID:            fmt.Sprintf("scheduled_%d", time.Now().UnixNano()),
		TxTemplate:    ats.convertTxTemplate(request.TxTemplate),
		Schedule:      ats.convertSchedule(request.Schedule),
		Status:        "active",
		CreatedAt:     time.Now(),
		MaxExecutions: request.MaxExecutions,
		MaxErrors:     request.MaxErrors,
	}

	// 计算下次执行时间
	scheduledTx.NextExecutionAt = ats.calculateNextExecution(scheduledTx.Schedule)

	// 保存定时交易
	err := ats.tradingManager.ScheduleTransaction(scheduledTx)
	if err != nil {
		return nil, fmt.Errorf("保存定时交易失败: %w", err)
	}

	// 添加到监控
	ats.addScheduleMonitor(scheduledTx)

	return scheduledTx, nil
}

// CreateConditionalTransaction 创建条件交易
func (ats *AdvancedTradingService) CreateConditionalTransaction(request *ConditionalTransactionRequest) (*core.ConditionalTx, error) {
	conditionalTx := &core.ConditionalTx{
		ID:            fmt.Sprintf("conditional_%d", time.Now().UnixNano()),
		TxTemplate:    ats.convertTxTemplate(request.TxTemplate),
		Conditions:    ats.convertConditions(request.Conditions),
		LogicOperator: request.LogicOperator,
		Status:        "active",
		CreatedAt:     time.Now(),
		ExpiresAt:     request.ExpiresAt,
	}

	// 保存条件交易
	err := ats.tradingManager.CreateConditionalTransaction(conditionalTx)
	if err != nil {
		return nil, fmt.Errorf("保存条件交易失败: %w", err)
	}

	// 添加到条件检查器
	ats.conditionChecker.AddConditionalTx(conditionalTx)

	return conditionalTx, nil
}

// CreateTradingStrategy 创建交易策略
func (ats *AdvancedTradingService) CreateTradingStrategy(request *TradingStrategyRequest) (*core.TradingStrategy, error) {
	strategy := &core.TradingStrategy{
		ID:        fmt.Sprintf("strategy_%d", time.Now().UnixNano()),
		Name:      request.Name,
		Type:      request.Type,
		Config:    ats.convertStrategyConfig(request.Config, request.Type),
		Status:    "created",
		CreatedAt: time.Now(),
	}

	// 保存策略
	err := ats.tradingManager.CreateTradingStrategy(strategy)
	if err != nil {
		return nil, fmt.Errorf("保存交易策略失败: %w", err)
	}

	// 如果设置了自动开始，则启动策略
	if request.AutoStart {
		err = ats.strategyEngine.StartStrategy(strategy.ID)
		if err != nil {
			return nil, fmt.Errorf("启动策略失败: %w", err)
		}
	}

	return strategy, nil
}

// GetStrategyPerformance 获取策略表现
func (ats *AdvancedTradingService) GetStrategyPerformance(strategyID string) (*StrategyAnalysis, error) {
	ats.mu.RLock()
	analysis, exists := ats.strategyEngine.performanceData[strategyID]
	ats.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("策略分析数据不存在")
	}

	return analysis, nil
}

// 私有方法实现

// NewConditionChecker 创建条件检查器
func NewConditionChecker() *ConditionChecker {
	return &ConditionChecker{
		activeConditions: make(map[string]*core.ConditionalTx),
		priceFeeds:       make(map[string]*PriceFeed),
		checkInterval:    30 * time.Second,
		stopChan:         make(chan bool),
	}
}

// NewStrategyEngine 创建策略引擎
func NewStrategyEngine() *StrategyEngine {
	return &StrategyEngine{
		activeStrategies: make(map[string]*core.TradingStrategy),
		executionQueue:   make([]*StrategyExecution, 0),
		performanceData:  make(map[string]*StrategyAnalysis),
	}
}

// startBackgroundServices 启动后台服务
func (ats *AdvancedTradingService) startBackgroundServices() {
	// 启动定时交易监控
	go ats.scheduleMonitorLoop()
	// 启动条件检查循环
	go ats.conditionChecker.Start()
	// 启动策略执行引擎
	go ats.strategyEngine.Start()
}

// scheduleMonitorLoop 定时交易监控循环
func (ats *AdvancedTradingService) scheduleMonitorLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ats.checkScheduledTransactions()
		}
	}
}

// checkScheduledTransactions 检查定时交易
func (ats *AdvancedTradingService) checkScheduledTransactions() {
	now := time.Now()

	ats.mu.RLock()
	for _, monitor := range ats.activeSchedules {
		if monitor.IsActive && now.After(monitor.NextCheck) {
			// 触发定时交易执行
			go ats.executeScheduledTransaction(monitor.ScheduledTxID)
		}
	}
	ats.mu.RUnlock()
}

// executeScheduledTransaction 执行定时交易
func (ats *AdvancedTradingService) executeScheduledTransaction(scheduledTxID string) {
	// 简化实现：记录执行
	record := &ExecutionRecord{
		ExecutedAt: time.Now(),
		TxHash:     "0x" + fmt.Sprintf("%x", time.Now().UnixNano()),
		Status:     "success",
		GasUsed:    21000,
	}

	ats.mu.Lock()
	if monitor, exists := ats.activeSchedules[scheduledTxID]; exists {
		monitor.ExecutionHistory = append(monitor.ExecutionHistory, record)
		monitor.LastExecution = &record.ExecutedAt
	}
	ats.mu.Unlock()
}

// AddConditionalTx 添加条件交易到检查器
func (cc *ConditionChecker) AddConditionalTx(tx *core.ConditionalTx) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.activeConditions[tx.ID] = tx
}

// Start 启动条件检查器
func (cc *ConditionChecker) Start() {
	ticker := time.NewTicker(cc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cc.checkConditions()
		case <-cc.stopChan:
			return
		}
	}
}

// checkConditions 检查所有条件
func (cc *ConditionChecker) checkConditions() {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	for _, tx := range cc.activeConditions {
		if cc.evaluateConditions(tx.Conditions, tx.LogicOperator) {
			// 条件满足，触发交易
			go cc.triggerConditionalTransaction(tx)
		}
	}
}

// evaluateConditions 评估条件
func (cc *ConditionChecker) evaluateConditions(conditions []*core.Condition, operator string) bool {
	// 简化实现：总是返回false
	return false
}

// triggerConditionalTransaction 触发条件交易
func (cc *ConditionChecker) triggerConditionalTransaction(tx *core.ConditionalTx) {
	// 简化实现：记录触发
	now := time.Now()
	tx.TriggeredAt = &now
}

// StartStrategy 启动策略
func (se *StrategyEngine) StartStrategy(strategyID string) error {
	// 简化实现
	return nil
}

// Start 启动策略引擎
func (se *StrategyEngine) Start() {
	// 策略执行逻辑
}

// 转换函数

// convertTxTemplate 转换交易模板
func (ats *AdvancedTradingService) convertTxTemplate(request *TxTemplateRequest) *core.TxTemplate {
	value, _ := new(big.Int).SetString(request.Value, 10)
	amount, _ := new(big.Int).SetString(request.Amount, 10)

	return &core.TxTemplate{
		From:         request.From,
		To:           request.To,
		Value:        value,
		TokenAddress: request.TokenAddress,
		Amount:       amount,
	}
}

// convertSchedule 转换执行计划
func (ats *AdvancedTradingService) convertSchedule(request *ScheduleRequest) *core.Schedule {
	interval, _ := time.ParseDuration(request.Interval)

	return &core.Schedule{
		Type:           request.Type,
		StartTime:      request.StartTime,
		EndTime:        request.EndTime,
		Interval:       interval,
		CronExpression: request.CronExpression,
		TimeZone:       request.TimeZone,
	}
}

// convertConditions 转换条件
func (ats *AdvancedTradingService) convertConditions(requests []*ConditionRequest) []*core.Condition {
	conditions := make([]*core.Condition, len(requests))

	for i, req := range requests {
		checkInterval, _ := time.ParseDuration(req.CheckInterval)
		conditions[i] = &core.Condition{
			Type:          req.Type,
			Target:        req.Target,
			Operator:      req.Operator,
			Value:         req.Value,
			CheckInterval: checkInterval,
		}
	}

	return conditions
}

// convertStrategyConfig 转换策略配置
func (ats *AdvancedTradingService) convertStrategyConfig(config map[string]interface{}, strategyType string) *core.StrategyConfig {
	// 简化实现：返回基础配置
	return &core.StrategyConfig{
		SlippageTolerance: 0.5,
		GasPriceStrategy:  "normal",
	}
}

// calculateNextExecution 计算下次执行时间
func (ats *AdvancedTradingService) calculateNextExecution(schedule *core.Schedule) time.Time {
	switch schedule.Type {
	case "once":
		return schedule.StartTime
	case "recurring":
		return time.Now().Add(schedule.Interval)
	default:
		return time.Now().Add(1 * time.Hour)
	}
}

// addScheduleMonitor 添加调度监控
func (ats *AdvancedTradingService) addScheduleMonitor(tx *core.ScheduledTx) {
	monitor := &ScheduleMonitor{
		ScheduledTxID:    tx.ID,
		NextCheck:        tx.NextExecutionAt,
		IsActive:         true,
		ExecutionHistory: make([]*ExecutionRecord, 0),
	}

	ats.mu.Lock()
	ats.activeSchedules[tx.ID] = monitor
	ats.mu.Unlock()
}
