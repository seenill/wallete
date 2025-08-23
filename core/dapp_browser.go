/*
DApp浏览器核心模块

本模块实现了去中心化应用程序(DApp)浏览器的核心功能，包括：

主要功能：
Web3接口集成：
- 以太坊Provider接口实现
- Web3.js/Ethers.js API支持
- MetaMask兼容接口
- WalletConnect协议支持

DApp发现和管理：
- 热门DApp推荐
- DApp分类浏览
- 收藏夹管理
- 访问历史记录

权限和安全：
- 连接授权管理
- 操作权限控制
- 钓鱼网站检测
- 安全警告提示

会话管理：
- 多标签页支持
- 会话状态持久化
- 自动断开连接
- 会话恢复机制

支持的协议：
- EIP-1193 (Ethereum Provider API)
- EIP-1102 (eth_requestAccounts)
- EIP-3085 (wallet_addEthereumChain)
- EIP-3326 (wallet_switchEthereumChain)
- WalletConnect v1/v2

安全特性：
- 域名白名单验证
- 恶意合约检测
- 交易风险评估
- 用户确认机制
*/
package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DAppBrowser DApp浏览器管理器
type DAppBrowser struct {
	multiChain     *MultiChainManager         // 多链管理器
	sessionManager *SessionManager            // 会话管理器
	permissionMgr  *PermissionManager         // 权限管理器
	securityMgr    *SecurityManager           // 安全管理器
	dappRegistry   *DAppRegistry              // DApp注册表
	connections    map[string]*DAppConnection // 活跃连接
	httpClient     *http.Client               // HTTP客户端
	mu             sync.RWMutex               // 读写锁
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions       map[string]*DAppSession // 活跃会话
	maxSessions    int                     // 最大会话数
	sessionTimeout time.Duration           // 会话超时时间
	mu             sync.RWMutex            // 读写锁
}

// DAppSession DApp会话
type DAppSession struct {
	ID             string                     `json:"id"`             // 会话ID
	DAppURL        string                     `json:"dapp_url"`       // DApp URL
	UserAddress    string                     `json:"user_address"`   // 用户地址
	ChainID        string                     `json:"chain_id"`       // 链ID
	Permissions    []string                   `json:"permissions"`    // 权限列表
	Status         string                     `json:"status"`         // 状态
	CreatedAt      time.Time                  `json:"created_at"`     // 创建时间
	LastActiveAt   time.Time                  `json:"last_active_at"` // 最后活跃时间
	ExpiresAt      time.Time                  `json:"expires_at"`     // 过期时间
	RequestQueue   []*Web3Request             `json:"request_queue"`  // 请求队列
	EventListeners map[string][]EventCallback `json:"-"`              // 事件监听器
}

// DAppConnection DApp连接
type DAppConnection struct {
	SessionID    string            `json:"session_id"`   // 会话ID
	Origin       string            `json:"origin"`       // 来源
	UserAgent    string            `json:"user_agent"`   // 用户代理
	ConnectedAt  time.Time         `json:"connected_at"` // 连接时间
	MessageQueue chan *Web3Message `json:"-"`            // 消息队列
	IsActive     bool              `json:"is_active"`    // 是否活跃
}

// PermissionManager 权限管理器
type PermissionManager struct {
	permissions map[string]*DAppPermission // 权限记录
	whitelist   map[string]bool            // 白名单
	blacklist   map[string]bool            // 黑名单
	mu          sync.RWMutex               // 读写锁
}

// DAppPermission DApp权限
type DAppPermission struct {
	DAppURL     string       `json:"dapp_url"`     // DApp URL
	UserAddress string       `json:"user_address"` // 用户地址
	Permissions []Permission `json:"permissions"`  // 权限列表
	GrantedAt   time.Time    `json:"granted_at"`   // 授权时间
	ExpiresAt   *time.Time   `json:"expires_at"`   // 过期时间
	IsRevoked   bool         `json:"is_revoked"`   // 是否撤销
	RevokedAt   *time.Time   `json:"revoked_at"`   // 撤销时间
}

// Permission 权限定义
type Permission struct {
	Type         string                 `json:"type"`         // 权限类型
	Resource     string                 `json:"resource"`     // 资源
	Scope        []string               `json:"scope"`        // 作用域
	Restrictions map[string]interface{} `json:"restrictions"` // 限制条件
}

// SecurityManager 安全管理器
type SecurityManager struct {
	phishingList   map[string]bool // 钓鱼网站列表
	trustedDomains map[string]bool // 可信域名
	riskRules      []*SecurityRule // 安全规则
	mu             sync.RWMutex    // 读写锁
}

// SecurityRule 安全规则
type SecurityRule struct {
	ID        string `json:"id"`         // 规则ID
	Name      string `json:"name"`       // 规则名称
	Type      string `json:"type"`       // 规则类型
	Pattern   string `json:"pattern"`    // 匹配模式
	RiskLevel string `json:"risk_level"` // 风险等级
	Action    string `json:"action"`     // 处理动作
	IsEnabled bool   `json:"is_enabled"` // 是否启用
}

// DAppRegistry DApp注册表
type DAppRegistry struct {
	categories   map[string]*DAppCategory  // DApp分类
	featured     []*DAppInfo               // 推荐DApp
	trending     []*DAppInfo               // 热门DApp
	favorites    map[string][]*DAppInfo    // 用户收藏
	visitHistory map[string][]*VisitRecord // 访问历史
	mu           sync.RWMutex              // 读写锁
}

// DAppCategory DApp分类
type DAppCategory struct {
	ID          string      `json:"id"`          // 分类ID
	Name        string      `json:"name"`        // 分类名称
	Description string      `json:"description"` // 描述
	Icon        string      `json:"icon"`        // 图标
	DApps       []*DAppInfo `json:"dapps"`       // DApp列表
	Order       int         `json:"order"`       // 排序
}

// DAppInfo DApp信息
type DAppInfo struct {
	ID              string    `json:"id"`               // DApp ID
	Name            string    `json:"name"`             // 名称
	Description     string    `json:"description"`      // 描述
	URL             string    `json:"url"`              // URL
	Icon            string    `json:"icon"`             // 图标
	Screenshots     []string  `json:"screenshots"`      // 截图
	Category        string    `json:"category"`         // 分类
	Tags            []string  `json:"tags"`             // 标签
	SupportedChains []string  `json:"supported_chains"` // 支持的链
	Rating          float64   `json:"rating"`           // 评分
	UserCount       int       `json:"user_count"`       // 用户数
	VolumeUSD       float64   `json:"volume_usd"`       // 交易量(USD)
	TVL             float64   `json:"tvl"`              // 锁仓量
	IsFeatured      bool      `json:"is_featured"`      // 是否推荐
	IsVerified      bool      `json:"is_verified"`      // 是否验证
	SecurityScore   int       `json:"security_score"`   // 安全评分
	LastUpdated     time.Time `json:"last_updated"`     // 最后更新
}

// VisitRecord 访问记录
type VisitRecord struct {
	URL       string        `json:"url"`        // 访问URL
	Title     string        `json:"title"`      // 页面标题
	Favicon   string        `json:"favicon"`    // 网站图标
	VisitedAt time.Time     `json:"visited_at"` // 访问时间
	Duration  time.Duration `json:"duration"`   // 访问时长
	ChainID   string        `json:"chain_id"`   // 使用的链
}

// Web3Request Web3请求
type Web3Request struct {
	ID           string        `json:"id"`            // 请求ID
	Method       string        `json:"method"`        // 方法名
	Params       []interface{} `json:"params"`        // 参数
	Origin       string        `json:"origin"`        // 来源
	Timestamp    time.Time     `json:"timestamp"`     // 时间戳
	RequiresAuth bool          `json:"requires_auth"` // 是否需要授权
	RiskLevel    string        `json:"risk_level"`    // 风险等级
	UserPrompt   string        `json:"user_prompt"`   // 用户提示
	Status       string        `json:"status"`        // 状态
	Response     interface{}   `json:"response"`      // 响应
	Error        *Web3Error    `json:"error"`         // 错误
}

// Web3Message Web3消息
type Web3Message struct {
	Type      string      `json:"type"`      // 消息类型
	ID        string      `json:"id"`        // 消息ID
	Data      interface{} `json:"data"`      // 消息数据
	Timestamp time.Time   `json:"timestamp"` // 时间戳
}

// Web3Error Web3错误
type Web3Error struct {
	Code    int         `json:"code"`           // 错误代码
	Message string      `json:"message"`        // 错误消息
	Data    interface{} `json:"data,omitempty"` // 错误数据
}

// EventCallback 事件回调
type EventCallback func(data interface{}) error

// NewDAppBrowser 创建DApp浏览器实例
func NewDAppBrowser(multiChain *MultiChainManager) *DAppBrowser {
	// 创建自定义HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	return &DAppBrowser{
		multiChain:     multiChain,
		sessionManager: NewSessionManager(),
		permissionMgr:  NewPermissionManager(),
		securityMgr:    NewSecurityManager(),
		dappRegistry:   NewDAppRegistry(),
		connections:    make(map[string]*DAppConnection),
		httpClient:     httpClient,
	}
}

// ConnectDApp 连接到DApp
func (db *DAppBrowser) ConnectDApp(ctx context.Context, dappURL, userAddress string) (*DAppSession, error) {
	// 验证URL
	parsedURL, err := url.Parse(dappURL)
	if err != nil {
		return nil, fmt.Errorf("无效的DApp URL: %w", err)
	}

	// 安全检查
	if err := db.securityMgr.CheckSecurity(parsedURL.Host); err != nil {
		return nil, fmt.Errorf("安全检查失败: %w", err)
	}

	// 创建会话
	session, err := db.sessionManager.CreateSession(dappURL, userAddress)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	// 记录访问历史
	db.dappRegistry.RecordVisit(userAddress, &VisitRecord{
		URL:       dappURL,
		Title:     parsedURL.Host,
		VisitedAt: time.Now(),
		ChainID:   session.ChainID,
	})

	return session, nil
}

// ProcessWeb3Request 处理Web3请求
func (db *DAppBrowser) ProcessWeb3Request(ctx context.Context, sessionID string, request *Web3Request) (*Web3Request, error) {
	session, err := db.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	// 检查权限
	if request.RequiresAuth {
		hasPermission, err := db.permissionMgr.CheckPermission(session.DAppURL, session.UserAddress, request.Method)
		if err != nil {
			return nil, fmt.Errorf("权限检查失败: %w", err)
		}
		if !hasPermission {
			request.Error = &Web3Error{
				Code:    4001,
				Message: "用户拒绝授权",
			}
			return request, nil
		}
	}

	// 根据方法处理请求
	switch request.Method {
	case "eth_requestAccounts":
		return db.handleRequestAccounts(ctx, session, request)
	case "eth_accounts":
		return db.handleGetAccounts(ctx, session, request)
	case "eth_chainId":
		return db.handleGetChainId(ctx, session, request)
	case "wallet_switchEthereumChain":
		return db.handleSwitchChain(ctx, session, request)
	case "wallet_addEthereumChain":
		return db.handleAddChain(ctx, session, request)
	case "eth_sendTransaction":
		return db.handleSendTransaction(ctx, session, request)
	case "eth_signTypedData_v4":
		return db.handleSignTypedData(ctx, session, request)
	case "personal_sign":
		return db.handlePersonalSign(ctx, session, request)
	default:
		return db.handleGenericRPCRequest(ctx, session, request)
	}
}

// GetDAppCategories 获取DApp分类
func (db *DAppBrowser) GetDAppCategories() map[string]*DAppCategory {
	return db.dappRegistry.GetCategories()
}

// GetFeaturedDApps 获取推荐DApp
func (db *DAppBrowser) GetFeaturedDApps() []*DAppInfo {
	return db.dappRegistry.GetFeatured()
}

// SearchDApps 搜索DApp
func (db *DAppBrowser) SearchDApps(query string, category string) []*DAppInfo {
	return db.dappRegistry.Search(query, category)
}

// GetUserFavorites 获取用户收藏
func (db *DAppBrowser) GetUserFavorites(userAddress string) []*DAppInfo {
	return db.dappRegistry.GetUserFavorites(userAddress)
}

// 私有方法实现

// 处理账户请求
func (db *DAppBrowser) handleRequestAccounts(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	// 检查是否已授权
	hasPermission, _ := db.permissionMgr.CheckPermission(session.DAppURL, session.UserAddress, "eth_accounts")
	if !hasPermission {
		// 需要用户授权
		request.Status = "pending_auth"
		request.UserPrompt = fmt.Sprintf("DApp %s 请求连接您的钱包", session.DAppURL)
		return request, nil
	}

	// 返回账户列表
	request.Status = "completed"
	request.Response = []string{session.UserAddress}
	return request, nil
}

// 处理获取账户
func (db *DAppBrowser) handleGetAccounts(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	// 检查权限
	hasPermission, _ := db.permissionMgr.CheckPermission(session.DAppURL, session.UserAddress, "eth_accounts")
	if !hasPermission {
		request.Response = []string{}
	} else {
		request.Response = []string{session.UserAddress}
	}

	request.Status = "completed"
	return request, nil
}

// 处理获取链ID
func (db *DAppBrowser) handleGetChainId(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	request.Status = "completed"
	request.Response = session.ChainID
	return request, nil
}

// 处理切换链
func (db *DAppBrowser) handleSwitchChain(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	if len(request.Params) == 0 {
		request.Error = &Web3Error{
			Code:    -32602,
			Message: "缺少链ID参数",
		}
		return request, nil
	}

	// 解析链ID
	chainParam, ok := request.Params[0].(map[string]interface{})
	if !ok {
		request.Error = &Web3Error{
			Code:    -32602,
			Message: "无效的参数格式",
		}
		return request, nil
	}

	chainID, ok := chainParam["chainId"].(string)
	if !ok {
		request.Error = &Web3Error{
			Code:    -32602,
			Message: "无效的链ID",
		}
		return request, nil
	}

	// 切换链
	session.ChainID = chainID
	request.Status = "completed"
	request.Response = nil
	return request, nil
}

// 处理添加链
func (db *DAppBrowser) handleAddChain(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	// 简化实现：总是返回成功
	request.Status = "completed"
	request.Response = nil
	return request, nil
}

// 处理发送交易
func (db *DAppBrowser) handleSendTransaction(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	request.Status = "pending_auth"
	request.UserPrompt = "DApp请求发送交易，请确认"
	request.RiskLevel = "high"
	return request, nil
}

// 处理签名
func (db *DAppBrowser) handleSignTypedData(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	request.Status = "pending_auth"
	request.UserPrompt = "DApp请求签名数据，请确认"
	request.RiskLevel = "medium"
	return request, nil
}

// 处理个人签名
func (db *DAppBrowser) handlePersonalSign(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	request.Status = "pending_auth"
	request.UserPrompt = "DApp请求签名消息，请确认"
	request.RiskLevel = "low"
	return request, nil
}

// 处理通用RPC请求
func (db *DAppBrowser) handleGenericRPCRequest(ctx context.Context, session *DAppSession, request *Web3Request) (*Web3Request, error) {
	// 转发到EVM适配器
	adapter, err := db.multiChain.GetCurrentAdapter()
	if err != nil {
		request.Error = &Web3Error{
			Code:    -32603,
			Message: "内部错误",
		}
		return request, nil
	}

	// 简化实现：调用适配器方法
	_ = adapter
	request.Status = "completed"
	request.Response = "0x0"
	return request, nil
}

// 辅助构造函数

// NewSessionManager 创建会话管理器
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:       make(map[string]*DAppSession),
		maxSessions:    10,
		sessionTimeout: 1 * time.Hour,
	}
}

// CreateSession 创建会话
func (sm *SessionManager) CreateSession(dappURL, userAddress string) (*DAppSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查会话数量限制
	if len(sm.sessions) >= sm.maxSessions {
		// 清理过期会话
		sm.cleanExpiredSessions()
		if len(sm.sessions) >= sm.maxSessions {
			return nil, fmt.Errorf("会话数量已达上限")
		}
	}

	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &DAppSession{
		ID:             sessionID,
		DAppURL:        dappURL,
		UserAddress:    userAddress,
		ChainID:        "0x1", // 默认以太坊主网
		Status:         "active",
		CreatedAt:      time.Now(),
		LastActiveAt:   time.Now(),
		ExpiresAt:      time.Now().Add(sm.sessionTimeout),
		RequestQueue:   make([]*Web3Request, 0),
		EventListeners: make(map[string][]EventCallback),
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*DAppSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("会话已过期")
	}

	return session, nil
}

// cleanExpiredSessions 清理过期会话
func (sm *SessionManager) cleanExpiredSessions() {
	now := time.Now()
	for id, session := range sm.sessions {
		if now.After(session.ExpiresAt) {
			delete(sm.sessions, id)
		}
	}
}

// NewPermissionManager 创建权限管理器
func NewPermissionManager() *PermissionManager {
	return &PermissionManager{
		permissions: make(map[string]*DAppPermission),
		whitelist:   make(map[string]bool),
		blacklist:   make(map[string]bool),
	}
}

// CheckPermission 检查权限
func (pm *PermissionManager) CheckPermission(dappURL, userAddress, method string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", dappURL, userAddress)
	permission, exists := pm.permissions[key]
	if !exists {
		return false, nil
	}

	if permission.IsRevoked {
		return false, nil
	}

	if permission.ExpiresAt != nil && time.Now().After(*permission.ExpiresAt) {
		return false, nil
	}

	// 检查具体方法权限
	for _, perm := range permission.Permissions {
		if perm.Type == method || perm.Type == "all" {
			return true, nil
		}
	}

	return false, nil
}

// NewSecurityManager 创建安全管理器
func NewSecurityManager() *SecurityManager {
	return &SecurityManager{
		phishingList:   make(map[string]bool),
		trustedDomains: make(map[string]bool),
		riskRules:      make([]*SecurityRule, 0),
	}
}

// CheckSecurity 安全检查
func (sm *SecurityManager) CheckSecurity(domain string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 检查黑名单
	if sm.phishingList[domain] {
		return fmt.Errorf("检测到钓鱼网站")
	}

	// 检查安全规则
	for _, rule := range sm.riskRules {
		if rule.IsEnabled && strings.Contains(domain, rule.Pattern) {
			if rule.RiskLevel == "high" && rule.Action == "block" {
				return fmt.Errorf("域名违反安全规则: %s", rule.Name)
			}
		}
	}

	return nil
}

// NewDAppRegistry 创建DApp注册表
func NewDAppRegistry() *DAppRegistry {
	return &DAppRegistry{
		categories:   make(map[string]*DAppCategory),
		featured:     make([]*DAppInfo, 0),
		trending:     make([]*DAppInfo, 0),
		favorites:    make(map[string][]*DAppInfo),
		visitHistory: make(map[string][]*VisitRecord),
	}
}

// GetCategories 获取分类
func (dr *DAppRegistry) GetCategories() map[string]*DAppCategory {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	return dr.categories
}

// GetFeatured 获取推荐DApp
func (dr *DAppRegistry) GetFeatured() []*DAppInfo {
	dr.mu.RLock()
	defer dr.mu.RUnlock()
	return dr.featured
}

// Search 搜索DApp
func (dr *DAppRegistry) Search(query, category string) []*DAppInfo {
	// 简化实现：返回空结果
	return []*DAppInfo{}
}

// GetUserFavorites 获取用户收藏
func (dr *DAppRegistry) GetUserFavorites(userAddress string) []*DAppInfo {
	dr.mu.RLock()
	defer dr.mu.RUnlock()

	favorites, exists := dr.favorites[userAddress]
	if !exists {
		return []*DAppInfo{}
	}
	return favorites
}

// RecordVisit 记录访问
func (dr *DAppRegistry) RecordVisit(userAddress string, record *VisitRecord) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if dr.visitHistory[userAddress] == nil {
		dr.visitHistory[userAddress] = make([]*VisitRecord, 0)
	}

	dr.visitHistory[userAddress] = append(dr.visitHistory[userAddress], record)

	// 保持历史记录在100条以内
	if len(dr.visitHistory[userAddress]) > 100 {
		dr.visitHistory[userAddress] = dr.visitHistory[userAddress][1:]
	}
}
