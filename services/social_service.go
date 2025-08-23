/*
社交功能业务服务层

本文件实现了社交功能的业务服务层，提供地址簿管理、转账记录分享、社交网络等服务。
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

// SocialService 社交功能服务
type SocialService struct {
	socialManager *core.SocialManager     // 社交管理器
	walletService *WalletService          // 钱包服务
	userSessions  map[string]*UserSession // 用户会话
	mu            sync.RWMutex            // 读写锁
}

// UserSession 用户会话
type UserSession struct {
	UserAddress  string              `json:"user_address"`  // 用户地址
	LastActivity time.Time           `json:"last_activity"` // 最后活动时间
	Settings     *UserSocialSettings `json:"settings"`      // 用户社交设置
}

// UserSocialSettings 用户社交设置
type UserSocialSettings struct {
	AutoShare      bool   `json:"auto_share"`      // 自动分享
	DefaultPrivacy string `json:"default_privacy"` // 默认隐私级别
	NotifyFollows  bool   `json:"notify_follows"`  // 关注通知
	NotifyShares   bool   `json:"notify_shares"`   // 分享通知
	AllowSearch    bool   `json:"allow_search"`    // 允许搜索
}

// ContactRequest 联系人请求
type ContactRequest struct {
	Name           string                  `json:"name" binding:"required"`
	Addresses      []ContactAddressRequest `json:"addresses" binding:"required"`
	Avatar         string                  `json:"avatar"`
	Note           string                  `json:"note"`
	Tags           []string                `json:"tags"`
	Groups         []string                `json:"groups"`
	ENSName        string                  `json:"ens_name"`
	SocialProfiles []SocialProfileRequest  `json:"social_profiles"`
}

// ContactAddressRequest 联系人地址请求
type ContactAddressRequest struct {
	Address string `json:"address" binding:"required"`
	Chain   string `json:"chain" binding:"required"`
	Label   string `json:"label"`
	Type    string `json:"type"`
}

// SocialProfileRequest 社交资料请求
type SocialProfileRequest struct {
	Platform string `json:"platform" binding:"required"`
	Username string `json:"username" binding:"required"`
	URL      string `json:"url"`
}

// ContactResponse 联系人响应
type ContactResponse struct {
	Contact          *core.Contact      `json:"contact"`           // 联系人信息
	RecentActivity   []*ActivitySummary `json:"recent_activity"`   // 最近活动
	TransactionCount int                `json:"transaction_count"` // 交易次数
	LastTransaction  *time.Time         `json:"last_transaction"`  // 最后交易时间
}

// ActivitySummary 活动摘要
type ActivitySummary struct {
	Type      string    `json:"type"`      // 活动类型
	Amount    *big.Int  `json:"amount"`    // 金额
	Token     string    `json:"token"`     // 代币
	Direction string    `json:"direction"` // 方向
	Timestamp time.Time `json:"timestamp"` // 时间戳
}

// ContactListRequest 联系人列表请求
type ContactListRequest struct {
	Search    string   `json:"search"`     // 搜索关键词
	Tags      []string `json:"tags"`       // 标签过滤
	Groups    []string `json:"groups"`     // 分组过滤
	SortBy    string   `json:"sort_by"`    // 排序字段
	SortOrder string   `json:"sort_order"` // 排序方向
	Limit     int      `json:"limit"`      // 限制数量
	Offset    int      `json:"offset"`     // 偏移量
}

// ContactListResponse 联系人列表响应
type ContactListResponse struct {
	Contacts []*ContactResponse `json:"contacts"` // 联系人列表
	Groups   []*GroupSummary    `json:"groups"`   // 分组摘要
	Tags     []*TagSummary      `json:"tags"`     // 标签摘要
	Total    int                `json:"total"`    // 总数
	HasMore  bool               `json:"has_more"` // 是否有更多
}

// GroupSummary 分组摘要
type GroupSummary struct {
	Group        *core.ContactGroup `json:"group"`         // 分组信息
	ContactCount int                `json:"contact_count"` // 联系人数量
}

// TagSummary 标签摘要
type TagSummary struct {
	Tag          *core.ContactTag `json:"tag"`           // 标签信息
	ContactCount int              `json:"contact_count"` // 联系人数量
}

// ShareTransactionRequest 分享交易请求
type ShareTransactionRequest struct {
	TransactionHash string              `json:"transaction_hash" binding:"required"`
	Message         string              `json:"message"`
	Privacy         SharePrivacyRequest `json:"privacy"`
	ExpiresIn       int                 `json:"expires_in"` // 过期时间（秒）
	Template        string              `json:"template"`   // 模板ID
}

// SharePrivacyRequest 分享隐私请求
type SharePrivacyRequest struct {
	IsPublic      bool     `json:"is_public"`
	RequireAuth   bool     `json:"require_auth"`
	AllowedUsers  []string `json:"allowed_users"`
	HideAmounts   bool     `json:"hide_amounts"`
	HideAddresses bool     `json:"hide_addresses"`
	Watermark     bool     `json:"watermark"`
}

// ShareTransactionResponse 分享交易响应
type ShareTransactionResponse struct {
	ShareRecord *core.ShareRecord `json:"share_record"` // 分享记录
	ShareURL    string            `json:"share_url"`    // 分享链接
	QRCode      string            `json:"qr_code"`      // 二维码
	ShortURL    string            `json:"short_url"`    // 短链接
}

// SocialNetworkRequest 社交网络请求
type SocialNetworkRequest struct {
	Action        string                 `json:"action" binding:"required"` // 操作类型
	TargetAddress string                 `json:"target_address"`            // 目标地址
	Data          map[string]interface{} `json:"data"`                      // 附加数据
}

// SocialNetworkResponse 社交网络响应
type SocialNetworkResponse struct {
	Success bool                   `json:"success"` // 是否成功
	Message string                 `json:"message"` // 消息
	Data    map[string]interface{} `json:"data"`    // 响应数据
}

// UserSocialProfileRequest 用户社交资料请求
type UserSocialProfileRequest struct {
	DisplayName string                 `json:"display_name"`
	Bio         string                 `json:"bio"`
	Avatar      string                 `json:"avatar"`
	Website     string                 `json:"website"`
	SocialLinks []SocialProfileRequest `json:"social_links"`
	Privacy     UserPrivacyRequest     `json:"privacy"`
}

// UserPrivacyRequest 用户隐私请求
type UserPrivacyRequest struct {
	ProfileVisibility  string `json:"profile_visibility"`
	ActivityVisibility string `json:"activity_visibility"`
	AllowFollow        bool   `json:"allow_follow"`
	AllowMessage       bool   `json:"allow_message"`
	ShowBalance        bool   `json:"show_balance"`
	ShowTransactions   bool   `json:"show_transactions"`
}

// NewSocialService 创建社交功能服务
func NewSocialService(walletService *WalletService) *SocialService {
	return &SocialService{
		socialManager: core.NewSocialManager(),
		walletService: walletService,
		userSessions:  make(map[string]*UserSession),
	}
}

// AddContact 添加联系人
func (ss *SocialService) AddContact(ctx context.Context, userAddress string, request *ContactRequest) (*ContactResponse, error) {
	// 验证用户地址
	if !ss.walletService.IsValidAddress(userAddress) {
		return nil, fmt.Errorf("无效的用户地址")
	}

	// 验证联系人地址
	for _, addr := range request.Addresses {
		if !ss.walletService.IsValidAddress(addr.Address) {
			return nil, fmt.Errorf("无效的联系人地址: %s", addr.Address)
		}
	}

	// 构建联系人对象
	contact := &core.Contact{
		Name:     request.Name,
		Avatar:   request.Avatar,
		Note:     request.Note,
		Tags:     request.Tags,
		Groups:   request.Groups,
		ENSName:  request.ENSName,
		Verified: false,
		Favorite: false,
	}

	// 添加地址
	for _, addrReq := range request.Addresses {
		contactAddr := core.ContactAddress{
			Address:   addrReq.Address,
			Chain:     addrReq.Chain,
			Label:     addrReq.Label,
			Type:      addrReq.Type,
			Verified:  false,
			CreatedAt: time.Now(),
		}
		contact.Addresses = append(contact.Addresses, contactAddr)
	}

	// 添加社交资料
	for _, profileReq := range request.SocialProfiles {
		profile := core.SocialProfile{
			Platform: profileReq.Platform,
			Username: profileReq.Username,
			URL:      profileReq.URL,
			Verified: false,
		}
		contact.SocialProfiles = append(contact.SocialProfiles, profile)
	}

	// 设置默认隐私
	contact.Privacy = core.ContactPrivacy{
		ShareAddress: true,
		ShareBalance: false,
		ShareHistory: false,
		Visibility:   "private",
	}

	// 添加联系人
	err := ss.socialManager.AddContact(ctx, userAddress, contact)
	if err != nil {
		return nil, fmt.Errorf("添加联系人失败: %w", err)
	}

	return ss.buildContactResponse(contact), nil
}

// GetContact 获取联系人
func (ss *SocialService) GetContact(ctx context.Context, userAddress, contactID string) (*ContactResponse, error) {
	contact, err := ss.socialManager.GetContact(ctx, userAddress, contactID)
	if err != nil {
		return nil, fmt.Errorf("获取联系人失败: %w", err)
	}

	return ss.buildContactResponse(contact), nil
}

// UpdateContact 更新联系人
func (ss *SocialService) UpdateContact(ctx context.Context, userAddress, contactID string, request *ContactRequest) (*ContactResponse, error) {
	// 获取现有联系人
	contact, err := ss.socialManager.GetContact(ctx, userAddress, contactID)
	if err != nil {
		return nil, fmt.Errorf("联系人不存在: %w", err)
	}

	// 更新联系人信息
	contact.Name = request.Name
	contact.Avatar = request.Avatar
	contact.Note = request.Note
	contact.Tags = request.Tags
	contact.Groups = request.Groups
	contact.ENSName = request.ENSName

	// 更新地址列表
	contact.Addresses = make([]core.ContactAddress, 0)
	for _, addrReq := range request.Addresses {
		contactAddr := core.ContactAddress{
			Address:   addrReq.Address,
			Chain:     addrReq.Chain,
			Label:     addrReq.Label,
			Type:      addrReq.Type,
			Verified:  false,
			CreatedAt: time.Now(),
		}
		contact.Addresses = append(contact.Addresses, contactAddr)
	}

	// 更新社交资料
	contact.SocialProfiles = make([]core.SocialProfile, 0)
	for _, profileReq := range request.SocialProfiles {
		profile := core.SocialProfile{
			Platform: profileReq.Platform,
			Username: profileReq.Username,
			URL:      profileReq.URL,
			Verified: false,
		}
		contact.SocialProfiles = append(contact.SocialProfiles, profile)
	}

	// 保存更新
	err = ss.socialManager.UpdateContact(ctx, userAddress, contact)
	if err != nil {
		return nil, fmt.Errorf("更新联系人失败: %w", err)
	}

	return ss.buildContactResponse(contact), nil
}

// DeleteContact 删除联系人
func (ss *SocialService) DeleteContact(ctx context.Context, userAddress, contactID string) error {
	return ss.socialManager.DeleteContact(ctx, userAddress, contactID)
}

// GetContactList 获取联系人列表
func (ss *SocialService) GetContactList(ctx context.Context, userAddress string, request *ContactListRequest) (*ContactListResponse, error) {
	// 简化实现：返回示例数据
	response := &ContactListResponse{
		Contacts: make([]*ContactResponse, 0),
		Groups:   make([]*GroupSummary, 0),
		Tags:     make([]*TagSummary, 0),
		Total:    0,
		HasMore:  false,
	}

	return response, nil
}

// ShareTransaction 分享交易
func (ss *SocialService) ShareTransaction(ctx context.Context, userAddress string, request *ShareTransactionRequest) (*ShareTransactionResponse, error) {
	// 构建分享内容
	shareContent := &core.ShareContent{
		TransactionHash: request.TransactionHash,
		Message:         request.Message,
		Timestamp:       time.Now(),
		Tags:            []string{"transaction", "share"},
	}

	// 构建隐私设置
	sharePrivacy := &core.SharePrivacy{
		IsPublic:      request.Privacy.IsPublic,
		RequireAuth:   request.Privacy.RequireAuth,
		AllowedUsers:  request.Privacy.AllowedUsers,
		HideAmounts:   request.Privacy.HideAmounts,
		HideAddresses: request.Privacy.HideAddresses,
		Watermark:     request.Privacy.Watermark,
	}

	// 创建分享记录
	shareRecord, err := ss.socialManager.CreateShareRecord(ctx, shareContent, sharePrivacy)
	if err != nil {
		return nil, fmt.Errorf("创建分享记录失败: %w", err)
	}

	// 生成二维码（简化实现）
	qrCode := fmt.Sprintf("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==")

	// 生成短链接
	shortURL := fmt.Sprintf("https://w.io/%s", shareRecord.ID[:8])

	response := &ShareTransactionResponse{
		ShareRecord: shareRecord,
		ShareURL:    shareRecord.ShareURL,
		QRCode:      qrCode,
		ShortURL:    shortURL,
	}

	return response, nil
}

// GetShareRecord 获取分享记录
func (ss *SocialService) GetShareRecord(ctx context.Context, shareID string) (*core.ShareRecord, error) {
	return ss.socialManager.GetShareRecord(ctx, shareID)
}

// SocialNetworkAction 社交网络操作
func (ss *SocialService) SocialNetworkAction(ctx context.Context, userAddress string, request *SocialNetworkRequest) (*SocialNetworkResponse, error) {
	var err error
	response := &SocialNetworkResponse{
		Data: make(map[string]interface{}),
	}

	switch request.Action {
	case "follow":
		err = ss.socialManager.FollowUser(ctx, userAddress, request.TargetAddress)
		response.Message = "关注成功"
	case "unfollow":
		err = ss.socialManager.UnfollowUser(ctx, userAddress, request.TargetAddress)
		response.Message = "取消关注成功"
	default:
		return nil, fmt.Errorf("不支持的操作: %s", request.Action)
	}

	if err != nil {
		response.Success = false
		response.Message = err.Error()
	} else {
		response.Success = true
	}

	return response, nil
}

// GetUserSocialProfile 获取用户社交资料
func (ss *SocialService) GetUserSocialProfile(ctx context.Context, userAddress string) (map[string]interface{}, error) {
	// 简化实现：返回示例数据
	profile := map[string]interface{}{
		"address":      userAddress,
		"display_name": "钱包用户",
		"bio":          "区块链爱好者",
		"avatar":       "https://example.com/avatar.png",
		"followers":    100,
		"following":    50,
		"joined_at":    time.Now().Add(-6 * time.Hour * 24 * 30),
	}

	return profile, nil
}

// UpdateUserSocialProfile 更新用户社交资料
func (ss *SocialService) UpdateUserSocialProfile(ctx context.Context, userAddress string, request *UserSocialProfileRequest) error {
	// 简化实现：返回成功
	return nil
}

// SearchUsers 搜索用户
func (ss *SocialService) SearchUsers(ctx context.Context, query string, limit int) ([]map[string]interface{}, error) {
	// 简化实现：返回示例数据
	users := make([]map[string]interface{}, 0)

	if query != "" {
		users = append(users, map[string]interface{}{
			"address":      "0x1234567890123456789012345678901234567890",
			"display_name": "示例用户",
			"avatar":       "https://example.com/avatar.png",
			"verified":     true,
		})
	}

	return users, nil
}

// 私有方法

// buildContactResponse 构建联系人响应
func (ss *SocialService) buildContactResponse(contact *core.Contact) *ContactResponse {
	return &ContactResponse{
		Contact:          contact,
		RecentActivity:   make([]*ActivitySummary, 0),
		TransactionCount: 0,
		LastTransaction:  nil,
	}
}

// getUserSession 获取用户会话
func (ss *SocialService) getUserSession(userAddress string) *UserSession {
	ss.mu.RLock()
	session, exists := ss.userSessions[userAddress]
	ss.mu.RUnlock()

	if !exists {
		session = &UserSession{
			UserAddress:  userAddress,
			LastActivity: time.Now(),
			Settings: &UserSocialSettings{
				AutoShare:      false,
				DefaultPrivacy: "private",
				NotifyFollows:  true,
				NotifyShares:   true,
				AllowSearch:    true,
			},
		}
		ss.mu.Lock()
		ss.userSessions[userAddress] = session
		ss.mu.Unlock()
	}

	return session
}

// updateUserActivity 更新用户活动
func (ss *SocialService) updateUserActivity(userAddress string) {
	session := ss.getUserSession(userAddress)
	session.LastActivity = time.Now()
}
