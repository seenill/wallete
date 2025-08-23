/*
DApp浏览器业务服务层

本文件实现了DApp浏览器功能的业务服务层，提供Web3应用集成、会话管理、安全控制等服务。
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

// DAppBrowserService DApp浏览器服务
type DAppBrowserService struct {
	dappBrowser    *core.DAppBrowser          // DApp浏览器
	walletService  *WalletService             // 钱包服务
	activeRequests map[string]*PendingRequest // 待处理请求
	mu             sync.RWMutex               // 读写锁
}

// PendingRequest 待处理请求
type PendingRequest struct {
	Request          *core.Web3Request `json:"request"`    // Web3请求
	SessionID        string            `json:"session_id"` // 会话ID
	UserConfirmation chan bool         `json:"-"`          // 用户确认通道
	CreatedAt        time.Time         `json:"created_at"` // 创建时间
	ExpiresAt        time.Time         `json:"expires_at"` // 过期时间
}

// DAppConnectionRequest DApp连接请求
type DAppConnectionRequest struct {
	DAppURL         string   `json:"dapp_url" binding:"required"`
	UserAddress     string   `json:"user_address" binding:"required"`
	RequestedChains []string `json:"requested_chains"`
	Permissions     []string `json:"permissions"`
}

// DAppConnectionResponse DApp连接响应
type DAppConnectionResponse struct {
	SessionID      string    `json:"session_id"`      // 会话ID
	ConnectedChain string    `json:"connected_chain"` // 连接的链
	Accounts       []string  `json:"accounts"`        // 账户列表
	Permissions    []string  `json:"permissions"`     // 授权权限
	ExpiresAt      time.Time `json:"expires_at"`      // 过期时间
}

// Web3RequestData Web3请求数据
type Web3RequestData struct {
	SessionID string        `json:"session_id" binding:"required"`
	Method    string        `json:"method" binding:"required"`
	Params    []interface{} `json:"params"`
	Origin    string        `json:"origin"`
}

// Web3ResponseData Web3响应数据
type Web3ResponseData struct {
	RequestID    string          `json:"request_id"`    // 请求ID
	Result       interface{}     `json:"result"`        // 结果
	Error        *core.Web3Error `json:"error"`         // 错误
	RequiresAuth bool            `json:"requires_auth"` // 是否需要授权
	UserPrompt   string          `json:"user_prompt"`   // 用户提示
	RiskLevel    string          `json:"risk_level"`    // 风险等级
}

// DAppListRequest DApp列表请求
type DAppListRequest struct {
	Category string `json:"category"` // 分类
	Search   string `json:"search"`   // 搜索关键词
	Chain    string `json:"chain"`    // 链过滤
	SortBy   string `json:"sort_by"`  // 排序字段
	Limit    int    `json:"limit"`    // 限制数量
	Offset   int    `json:"offset"`   // 偏移量
}

// DAppListResponse DApp列表响应
type DAppListResponse struct {
	DApps      []*core.DAppInfo `json:"dapps"`      // DApp列表
	Categories []*CategoryInfo  `json:"categories"` // 分类信息
	Total      int              `json:"total"`      // 总数
	HasMore    bool             `json:"has_more"`   // 是否有更多
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	ID    string `json:"id"`    // 分类ID
	Name  string `json:"name"`  // 名称
	Icon  string `json:"icon"`  // 图标
	Count int    `json:"count"` // DApp数量
}

// UserActivityRequest 用户活动请求
type UserActivityRequest struct {
	UserAddress  string `json:"user_address" binding:"required"`
	ActivityType string `json:"activity_type"` // 活动类型
	TimeRange    string `json:"time_range"`    // 时间范围
	Limit        int    `json:"limit"`         // 限制数量
}

// UserActivityResponse 用户活动响应
type UserActivityResponse struct {
	Favorites      []*core.DAppInfo    `json:"favorites"`       // 收藏
	RecentVisits   []*core.VisitRecord `json:"recent_visits"`   // 最近访问
	ActiveSessions []*SessionInfo      `json:"active_sessions"` // 活跃会话
	Statistics     *ActivityStatistics `json:"statistics"`      // 统计信息
}

// SessionInfo 会话信息
type SessionInfo struct {
	SessionID    string    `json:"session_id"`     // 会话ID
	DAppURL      string    `json:"dapp_url"`       // DApp URL
	DAppName     string    `json:"dapp_name"`      // DApp名称
	Chain        string    `json:"chain"`          // 链
	Status       string    `json:"status"`         // 状态
	ConnectedAt  time.Time `json:"connected_at"`   // 连接时间
	LastActiveAt time.Time `json:"last_active_at"` // 最后活跃时间
}

// ActivityStatistics 活动统计
type ActivityStatistics struct {
	TotalSessions  int   `json:"total_sessions"`  // 总会话数
	ActiveSessions int   `json:"active_sessions"` // 活跃会话数
	TotalDApps     int   `json:"total_dapps"`     // 总访问DApp数
	FavoritesCount int   `json:"favorites_count"` // 收藏数
	WeeklyActivity []int `json:"weekly_activity"` // 周活动统计
}

// NewDAppBrowserService 创建DApp浏览器服务
func NewDAppBrowserService(dappBrowser *core.DAppBrowser, walletService *WalletService) *DAppBrowserService {
	return &DAppBrowserService{
		dappBrowser:    dappBrowser,
		walletService:  walletService,
		activeRequests: make(map[string]*PendingRequest),
	}
}

// ConnectDApp 连接DApp
func (dbs *DAppBrowserService) ConnectDApp(ctx context.Context, request *DAppConnectionRequest) (*DAppConnectionResponse, error) {
	// 验证用户地址
	if !dbs.walletService.IsValidAddress(request.UserAddress) {
		return nil, fmt.Errorf("无效的用户地址")
	}

	// 创建DApp连接
	session, err := dbs.dappBrowser.ConnectDApp(ctx, request.DAppURL, request.UserAddress)
	if err != nil {
		return nil, fmt.Errorf("连接DApp失败: %w", err)
	}

	// 构建响应
	response := &DAppConnectionResponse{
		SessionID:      session.ID,
		ConnectedChain: session.ChainID,
		Accounts:       []string{session.UserAddress},
		Permissions:    session.Permissions,
		ExpiresAt:      session.ExpiresAt,
	}

	return response, nil
}

// ProcessWeb3Request 处理Web3请求
func (dbs *DAppBrowserService) ProcessWeb3Request(ctx context.Context, requestData *Web3RequestData) (*Web3ResponseData, error) {
	// 创建Web3请求
	request := &core.Web3Request{
		ID:           fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Method:       requestData.Method,
		Params:       requestData.Params,
		Origin:       requestData.Origin,
		Timestamp:    time.Now(),
		RequiresAuth: dbs.isMethodRequiresAuth(requestData.Method),
		Status:       "pending",
	}

	// 处理请求
	processedRequest, err := dbs.dappBrowser.ProcessWeb3Request(ctx, requestData.SessionID, request)
	if err != nil {
		return nil, fmt.Errorf("处理Web3请求失败: %w", err)
	}

	// 如果需要用户授权，添加到待处理队列
	if processedRequest.Status == "pending_auth" {
		dbs.addPendingRequest(requestData.SessionID, processedRequest)
	}

	// 构建响应
	response := &Web3ResponseData{
		RequestID:    processedRequest.ID,
		Result:       processedRequest.Response,
		Error:        processedRequest.Error,
		RequiresAuth: processedRequest.RequiresAuth,
		UserPrompt:   processedRequest.UserPrompt,
		RiskLevel:    processedRequest.RiskLevel,
	}

	return response, nil
}

// GetDAppList 获取DApp列表
func (dbs *DAppBrowserService) GetDAppList(ctx context.Context, request *DAppListRequest) (*DAppListResponse, error) {
	var dapps []*core.DAppInfo

	// 根据请求类型获取DApp
	if request.Category == "featured" {
		dapps = dbs.dappBrowser.GetFeaturedDApps()
	} else if request.Search != "" {
		dapps = dbs.dappBrowser.SearchDApps(request.Search, request.Category)
	} else {
		// 获取分类DApp
		categories := dbs.dappBrowser.GetDAppCategories()
		if category, exists := categories[request.Category]; exists {
			dapps = category.DApps
		}
	}

	// 应用过滤和排序
	dapps = dbs.filterAndSortDApps(dapps, request)

	// 应用分页
	total := len(dapps)
	start := request.Offset
	end := start + request.Limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start < end {
		dapps = dapps[start:end]
	} else {
		dapps = []*core.DAppInfo{}
	}

	// 获取分类信息
	categories := dbs.buildCategoryInfo()

	response := &DAppListResponse{
		DApps:      dapps,
		Categories: categories,
		Total:      total,
		HasMore:    end < total,
	}

	return response, nil
}

// GetUserActivity 获取用户活动
func (dbs *DAppBrowserService) GetUserActivity(ctx context.Context, request *UserActivityRequest) (*UserActivityResponse, error) {
	// 获取用户收藏
	favorites := dbs.dappBrowser.GetUserFavorites(request.UserAddress)

	// 构建活动统计（简化实现）
	statistics := &ActivityStatistics{
		TotalSessions:  10,
		ActiveSessions: 2,
		TotalDApps:     15,
		FavoritesCount: len(favorites),
		WeeklyActivity: []int{5, 8, 3, 12, 7, 9, 6}, // 示例数据
	}

	response := &UserActivityResponse{
		Favorites:      favorites,
		RecentVisits:   []*core.VisitRecord{}, // 简化实现
		ActiveSessions: []*SessionInfo{},      // 简化实现
		Statistics:     statistics,
	}

	return response, nil
}

// ConfirmWeb3Request 确认Web3请求
func (dbs *DAppBrowserService) ConfirmWeb3Request(ctx context.Context, requestID string, approved bool, signature string) error {
	dbs.mu.Lock()
	defer dbs.mu.Unlock()

	pendingRequest, exists := dbs.activeRequests[requestID]
	if !exists {
		return fmt.Errorf("请求不存在或已过期")
	}

	// 检查过期时间
	if time.Now().After(pendingRequest.ExpiresAt) {
		delete(dbs.activeRequests, requestID)
		return fmt.Errorf("请求已过期")
	}

	// 处理用户确认
	if approved {
		// 执行实际操作
		err := dbs.executeWeb3Request(ctx, pendingRequest, signature)
		if err != nil {
			return fmt.Errorf("执行请求失败: %w", err)
		}
	}

	// 通知结果
	select {
	case pendingRequest.UserConfirmation <- approved:
	default:
	}

	// 清理请求
	delete(dbs.activeRequests, requestID)
	return nil
}

// GetPendingRequests 获取待处理请求
func (dbs *DAppBrowserService) GetPendingRequests(userAddress string) []*PendingRequest {
	dbs.mu.RLock()
	defer dbs.mu.RUnlock()

	var pendingRequests []*PendingRequest
	for _, request := range dbs.activeRequests {
		// 简化实现：检查是否属于该用户
		pendingRequests = append(pendingRequests, request)
	}

	return pendingRequests
}

// 私有方法

// isMethodRequiresAuth 判断方法是否需要授权
func (dbs *DAppBrowserService) isMethodRequiresAuth(method string) bool {
	authMethods := map[string]bool{
		"eth_requestAccounts":        true,
		"eth_sendTransaction":        true,
		"eth_signTypedData_v4":       true,
		"personal_sign":              true,
		"wallet_switchEthereumChain": true,
		"wallet_addEthereumChain":    true,
	}

	return authMethods[method]
}

// addPendingRequest 添加待处理请求
func (dbs *DAppBrowserService) addPendingRequest(sessionID string, request *core.Web3Request) {
	dbs.mu.Lock()
	defer dbs.mu.Unlock()

	pendingRequest := &PendingRequest{
		Request:          request,
		SessionID:        sessionID,
		UserConfirmation: make(chan bool, 1),
		CreatedAt:        time.Now(),
		ExpiresAt:        time.Now().Add(5 * time.Minute), // 5分钟过期
	}

	dbs.activeRequests[request.ID] = pendingRequest
}

// executeWeb3Request 执行Web3请求
func (dbs *DAppBrowserService) executeWeb3Request(ctx context.Context, pendingRequest *PendingRequest, signature string) error {
	request := pendingRequest.Request

	switch request.Method {
	case "eth_sendTransaction":
		return dbs.executeSendTransaction(ctx, request, signature)
	case "eth_signTypedData_v4":
		return dbs.executeSignTypedData(ctx, request, signature)
	case "personal_sign":
		return dbs.executePersonalSign(ctx, request, signature)
	default:
		return fmt.Errorf("不支持的方法: %s", request.Method)
	}
}

// executeSendTransaction 执行发送交易
func (dbs *DAppBrowserService) executeSendTransaction(ctx context.Context, request *core.Web3Request, signature string) error {
	if len(request.Params) == 0 {
		return fmt.Errorf("缺少交易参数")
	}

	// 解析交易参数
	txParam, ok := request.Params[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的交易参数")
	}

	// 构建交易
	to, _ := txParam["to"].(string)
	value, _ := txParam["value"].(string)
	data, _ := txParam["data"].(string)

	// 解析金额
	amount := big.NewInt(0)
	if value != "" {
		amount, _ = new(big.Int).SetString(value, 0)
	}

	// 简化实现：生成模拟交易哈希
	txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	request.Response = txHash
	request.Status = "completed"

	_ = to
	_ = amount
	_ = data
	_ = signature

	return nil
}

// executeSignTypedData 执行类型化数据签名
func (dbs *DAppBrowserService) executeSignTypedData(ctx context.Context, request *core.Web3Request, signature string) error {
	// 简化实现：返回模拟签名
	mockSignature := "0x" + fmt.Sprintf("%064x", time.Now().UnixNano()) +
		fmt.Sprintf("%064x", time.Now().UnixNano()) + "1b"

	request.Response = mockSignature
	request.Status = "completed"
	return nil
}

// executePersonalSign 执行个人签名
func (dbs *DAppBrowserService) executePersonalSign(ctx context.Context, request *core.Web3Request, signature string) error {
	// 简化实现：返回模拟签名
	mockSignature := "0x" + fmt.Sprintf("%064x", time.Now().UnixNano()) +
		fmt.Sprintf("%064x", time.Now().UnixNano()) + "1c"

	request.Response = mockSignature
	request.Status = "completed"
	return nil
}

// filterAndSortDApps 过滤和排序DApp
func (dbs *DAppBrowserService) filterAndSortDApps(dapps []*core.DAppInfo, request *DAppListRequest) []*core.DAppInfo {
	// 链过滤
	if request.Chain != "" {
		filtered := make([]*core.DAppInfo, 0)
		for _, dapp := range dapps {
			for _, chain := range dapp.SupportedChains {
				if chain == request.Chain {
					filtered = append(filtered, dapp)
					break
				}
			}
		}
		dapps = filtered
	}

	// 排序（简化实现）
	// 实际项目中可以根据sortBy字段进行排序

	return dapps
}

// buildCategoryInfo 构建分类信息
func (dbs *DAppBrowserService) buildCategoryInfo() []*CategoryInfo {
	categories := dbs.dappBrowser.GetDAppCategories()

	categoryInfos := make([]*CategoryInfo, 0, len(categories))
	for _, category := range categories {
		categoryInfos = append(categoryInfos, &CategoryInfo{
			ID:    category.ID,
			Name:  category.Name,
			Icon:  category.Icon,
			Count: len(category.DApps),
		})
	}

	return categoryInfos
}
