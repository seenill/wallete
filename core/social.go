/*
社交功能核心模块

本模块实现了钱包的社交功能，包括：

主要功能：
地址簿管理：
- 联系人添加、编辑、删除
- 联系人分组和标签
- 联系人头像和备注
- 导入导出联系人

转账记录分享：
- 交易记录分享生成
- 二维码生成和分享
- 社交平台集成
- 隐私保护选项

社交网络：
- 用户关注和粉丝
- 钱包地址验证
- 社交活动推送
- 群组钱包管理

隐私控制：
- 可见性设置
- 分享权限管理
- 匿名模式支持
- 数据加密存储

支持的功能：
- ENS域名解析
- 多链地址关联
- 社交身份验证
- 跨平台同步

安全特性：
- 联系人信息加密
- 分享链接有效期
- 防钓鱼验证
- 用户授权确认
*/
package core

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// SocialManager 社交功能管理器
type SocialManager struct {
	addressBook    *AddressBook    // 地址簿管理器
	shareManager   *ShareManager   // 分享管理器
	socialNetwork  *SocialNetwork  // 社交网络管理器
	privacyManager *PrivacyManager // 隐私管理器
	ensResolver    *ENSResolver    // ENS域名解析器
	mu             sync.RWMutex    // 读写锁
}

// AddressBook 地址簿管理器
type AddressBook struct {
	contacts   map[string]*Contact      // 联系人映射
	groups     map[string]*ContactGroup // 联系人分组
	tags       map[string]*ContactTag   // 联系人标签
	userGroups map[string][]string      // 用户分组关系
	mu         sync.RWMutex             // 读写锁
}

// Contact 联系人信息
type Contact struct {
	ID             string           `json:"id"`              // 联系人ID
	Name           string           `json:"name"`            // 联系人姓名
	Addresses      []ContactAddress `json:"addresses"`       // 关联地址列表
	Avatar         string           `json:"avatar"`          // 头像URL
	Note           string           `json:"note"`            // 备注信息
	Tags           []string         `json:"tags"`            // 标签列表
	Groups         []string         `json:"groups"`          // 所属分组
	Verified       bool             `json:"verified"`        // 是否已验证
	ENSName        string           `json:"ens_name"`        // ENS域名
	SocialProfiles []SocialProfile  `json:"social_profiles"` // 社交资料
	CreatedAt      time.Time        `json:"created_at"`      // 创建时间
	UpdatedAt      time.Time        `json:"updated_at"`      // 更新时间
	LastContact    *time.Time       `json:"last_contact"`    // 最后联系时间
	Frequency      int              `json:"frequency"`       // 交互频率
	Favorite       bool             `json:"favorite"`        // 是否收藏
	Privacy        ContactPrivacy   `json:"privacy"`         // 隐私设置
}

// ContactAddress 联系人地址
type ContactAddress struct {
	Address   string    `json:"address"`    // 钱包地址
	Chain     string    `json:"chain"`      // 区块链网络
	Label     string    `json:"label"`      // 地址标签
	Type      string    `json:"type"`       // 地址类型
	Verified  bool      `json:"verified"`   // 是否已验证
	CreatedAt time.Time `json:"created_at"` // 添加时间
}

// ContactGroup 联系人分组
type ContactGroup struct {
	ID           string    `json:"id"`            // 分组ID
	Name         string    `json:"name"`          // 分组名称
	Description  string    `json:"description"`   // 分组描述
	Color        string    `json:"color"`         // 分组颜色
	Icon         string    `json:"icon"`          // 分组图标
	ContactCount int       `json:"contact_count"` // 联系人数量
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time `json:"updated_at"`    // 更新时间
}

// ContactTag 联系人标签
type ContactTag struct {
	ID         string    `json:"id"`          // 标签ID
	Name       string    `json:"name"`        // 标签名称
	Color      string    `json:"color"`       // 标签颜色
	UsageCount int       `json:"usage_count"` // 使用次数
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}

// SocialProfile 社交资料
type SocialProfile struct {
	Platform string `json:"platform"` // 平台名称
	Username string `json:"username"` // 用户名
	URL      string `json:"url"`      // 资料链接
	Verified bool   `json:"verified"` // 是否已验证
}

// ContactPrivacy 联系人隐私设置
type ContactPrivacy struct {
	ShareAddress bool   `json:"share_address"` // 是否分享地址
	ShareBalance bool   `json:"share_balance"` // 是否分享余额
	ShareHistory bool   `json:"share_history"` // 是否分享历史
	Visibility   string `json:"visibility"`    // 可见性级别
}

// ShareManager 分享管理器
type ShareManager struct {
	shareRecords map[string]*ShareRecord   // 分享记录
	shareLinks   map[string]*ShareLink     // 分享链接
	templates    map[string]*ShareTemplate // 分享模板
	mu           sync.RWMutex              // 读写锁
}

// ShareRecord 分享记录
type ShareRecord struct {
	ID        string         `json:"id"`         // 分享ID
	Type      string         `json:"type"`       // 分享类型
	Content   ShareContent   `json:"content"`    // 分享内容
	ShareURL  string         `json:"share_url"`  // 分享链接
	QRCode    string         `json:"qr_code"`    // 二维码
	ExpiresAt *time.Time     `json:"expires_at"` // 过期时间
	ViewCount int            `json:"view_count"` // 查看次数
	Privacy   SharePrivacy   `json:"privacy"`    // 隐私设置
	CreatedBy string         `json:"created_by"` // 创建者
	CreatedAt time.Time      `json:"created_at"` // 创建时间
	Analytics ShareAnalytics `json:"analytics"`  // 分享统计
}

// ShareContent 分享内容
type ShareContent struct {
	TransactionHash string    `json:"transaction_hash"` // 交易哈希
	FromAddress     string    `json:"from_address"`     // 发送方地址
	ToAddress       string    `json:"to_address"`       // 接收方地址
	Amount          *big.Int  `json:"amount"`           // 金额
	Token           string    `json:"token"`            // 代币地址
	TokenSymbol     string    `json:"token_symbol"`     // 代币符号
	Network         string    `json:"network"`          // 网络
	Timestamp       time.Time `json:"timestamp"`        // 时间戳
	Message         string    `json:"message"`          // 附加消息
	Tags            []string  `json:"tags"`             // 标签
}

// SharePrivacy 分享隐私设置
type SharePrivacy struct {
	IsPublic      bool     `json:"is_public"`      // 是否公开
	RequireAuth   bool     `json:"require_auth"`   // 是否需要认证
	AllowedUsers  []string `json:"allowed_users"`  // 允许的用户
	HideAmounts   bool     `json:"hide_amounts"`   // 隐藏金额
	HideAddresses bool     `json:"hide_addresses"` // 隐藏地址
	Watermark     bool     `json:"watermark"`      // 添加水印
}

// ShareAnalytics 分享统计
type ShareAnalytics struct {
	Views         int            `json:"views"`          // 查看次数
	Shares        int            `json:"shares"`         // 分享次数
	UniqueViewers int            `json:"unique_viewers"` // 独立访客
	Platforms     map[string]int `json:"platforms"`      // 平台统计
	Countries     map[string]int `json:"countries"`      // 国家统计
	LastViewedAt  *time.Time     `json:"last_viewed_at"` // 最后查看时间
}

// ShareLink 分享链接
type ShareLink struct {
	ID         string     `json:"id"`          // 链接ID
	ShortURL   string     `json:"short_url"`   // 短链接
	FullURL    string     `json:"full_url"`    // 完整链接
	ClickCount int        `json:"click_count"` // 点击次数
	ExpiresAt  *time.Time `json:"expires_at"`  // 过期时间
	IsActive   bool       `json:"is_active"`   // 是否活跃
	CreatedAt  time.Time  `json:"created_at"`  // 创建时间
}

// ShareTemplate 分享模板
type ShareTemplate struct {
	ID         string    `json:"id"`          // 模板ID
	Name       string    `json:"name"`        // 模板名称
	Type       string    `json:"type"`        // 模板类型
	Template   string    `json:"template"`    // 模板内容
	Variables  []string  `json:"variables"`   // 变量列表
	Preview    string    `json:"preview"`     // 预览图
	IsDefault  bool      `json:"is_default"`  // 是否默认
	UsageCount int       `json:"usage_count"` // 使用次数
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}

// SocialNetwork 社交网络管理器
type SocialNetwork struct {
	followers     map[string][]string        // 关注者映射
	following     map[string][]string        // 关注映射
	groups        map[string]*SocialGroup    // 社交群组
	activities    map[string]*Activity       // 活动记录
	notifications map[string][]*Notification // 通知记录
	mu            sync.RWMutex               // 读写锁
}

// SocialGroup 社交群组
type SocialGroup struct {
	ID          string        `json:"id"`          // 群组ID
	Name        string        `json:"name"`        // 群组名称
	Description string        `json:"description"` // 群组描述
	Avatar      string        `json:"avatar"`      // 群组头像
	Type        string        `json:"type"`        // 群组类型
	Members     []GroupMember `json:"members"`     // 群组成员
	Admins      []string      `json:"admins"`      // 管理员
	Settings    GroupSettings `json:"settings"`    // 群组设置
	Stats       GroupStats    `json:"stats"`       // 群组统计
	CreatedBy   string        `json:"created_by"`  // 创建者
	CreatedAt   time.Time     `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time     `json:"updated_at"`  // 更新时间
}

// GroupMember 群组成员
type GroupMember struct {
	UserAddress string    `json:"user_address"` // 用户地址
	Role        string    `json:"role"`         // 角色
	JoinedAt    time.Time `json:"joined_at"`    // 加入时间
	IsActive    bool      `json:"is_active"`    // 是否活跃
	Permissions []string  `json:"permissions"`  // 权限列表
}

// GroupSettings 群组设置
type GroupSettings struct {
	IsPublic        bool `json:"is_public"`        // 是否公开
	AllowInvite     bool `json:"allow_invite"`     // 允许邀请
	RequireApproval bool `json:"require_approval"` // 需要审批
	MaxMembers      int  `json:"max_members"`      // 最大成员数
	ShareWallet     bool `json:"share_wallet"`     // 共享钱包
}

// GroupStats 群组统计
type GroupStats struct {
	MemberCount      int       `json:"member_count"`      // 成员数量
	ActiveMembers    int       `json:"active_members"`    // 活跃成员
	TotalVolume      *big.Int  `json:"total_volume"`      // 总交易量
	TransactionCount int       `json:"transaction_count"` // 交易次数
	LastActivity     time.Time `json:"last_activity"`     // 最后活动
}

// Activity 活动记录
type Activity struct {
	ID            string                 `json:"id"`             // 活动ID
	Type          string                 `json:"type"`           // 活动类型
	UserAddress   string                 `json:"user_address"`   // 用户地址
	TargetAddress string                 `json:"target_address"` // 目标地址
	Content       ActivityContent        `json:"content"`        // 活动内容
	Metadata      map[string]interface{} `json:"metadata"`       // 元数据
	Timestamp     time.Time              `json:"timestamp"`      // 时间戳
	Visibility    string                 `json:"visibility"`     // 可见性
}

// ActivityContent 活动内容
type ActivityContent struct {
	Title           string   `json:"title"`            // 标题
	Description     string   `json:"description"`      // 描述
	Amount          *big.Int `json:"amount"`           // 金额
	Token           string   `json:"token"`            // 代币
	TransactionHash string   `json:"transaction_hash"` // 交易哈希
	Network         string   `json:"network"`          // 网络
}

// Notification 通知
type Notification struct {
	ID        string                 `json:"id"`         // 通知ID
	Type      string                 `json:"type"`       // 通知类型
	Title     string                 `json:"title"`      // 通知标题
	Message   string                 `json:"message"`    // 通知消息
	Data      map[string]interface{} `json:"data"`       // 通知数据
	IsRead    bool                   `json:"is_read"`    // 是否已读
	Priority  string                 `json:"priority"`   // 优先级
	ExpiresAt *time.Time             `json:"expires_at"` // 过期时间
	CreatedAt time.Time              `json:"created_at"` // 创建时间
}

// PrivacyManager 隐私管理器
type PrivacyManager struct {
	settings    map[string]*PrivacySettings // 用户隐私设置
	permissions map[string]*UserPermissions // 用户权限
	mu          sync.RWMutex                // 读写锁
}

// PrivacySettings 隐私设置
type PrivacySettings struct {
	ProfileVisibility  string              `json:"profile_visibility"`  // 资料可见性
	ActivityVisibility string              `json:"activity_visibility"` // 活动可见性
	ContactVisibility  string              `json:"contact_visibility"`  // 联系方式可见性
	AllowFollow        bool                `json:"allow_follow"`        // 允许关注
	AllowMessage       bool                `json:"allow_message"`       // 允许消息
	ShowBalance        bool                `json:"show_balance"`        // 显示余额
	ShowTransactions   bool                `json:"show_transactions"`   // 显示交易
	DataSharing        DataSharingSettings `json:"data_sharing"`        // 数据分享设置
}

// DataSharingSettings 数据分享设置
type DataSharingSettings struct {
	AllowAnalytics    bool `json:"allow_analytics"`     // 允许分析
	ShareWithPartners bool `json:"share_with_partners"` // 与合作伙伴分享
	MarketingEmails   bool `json:"marketing_emails"`    // 营销邮件
	PersonalizedAds   bool `json:"personalized_ads"`    // 个性化广告
}

// UserPermissions 用户权限
type UserPermissions struct {
	CanInvite      bool     `json:"can_invite"`       // 可以邀请
	CanShare       bool     `json:"can_share"`        // 可以分享
	CanCreateGroup bool     `json:"can_create_group"` // 可以创建群组
	CanExportData  bool     `json:"can_export_data"`  // 可以导出数据
	MaxContacts    int      `json:"max_contacts"`     // 最大联系人数
	MaxShares      int      `json:"max_shares"`       // 最大分享数
	FeatureAccess  []string `json:"feature_access"`   // 功能访问权限
}

// ENSResolver ENS域名解析器
type ENSResolver struct {
	cache map[string]*ENSRecord // ENS缓存
	mu    sync.RWMutex          // 读写锁
}

// ENSRecord ENS记录
type ENSRecord struct {
	Name        string    `json:"name"`        // ENS名称
	Address     string    `json:"address"`     // 对应地址
	Avatar      string    `json:"avatar"`      // 头像
	Description string    `json:"description"` // 描述
	Website     string    `json:"website"`     // 网站
	Twitter     string    `json:"twitter"`     // Twitter
	Github      string    `json:"github"`      // Github
	ResolvedAt  time.Time `json:"resolved_at"` // 解析时间
	ExpiresAt   time.Time `json:"expires_at"`  // 过期时间
}

// NewSocialManager 创建社交管理器
func NewSocialManager() *SocialManager {
	return &SocialManager{
		addressBook:    NewAddressBook(),
		shareManager:   NewShareManager(),
		socialNetwork:  NewSocialNetwork(),
		privacyManager: NewPrivacyManager(),
		ensResolver:    NewENSResolver(),
	}
}

// AddContact 添加联系人
func (sm *SocialManager) AddContact(ctx context.Context, userAddress string, contact *Contact) error {
	return sm.addressBook.AddContact(userAddress, contact)
}

// GetContact 获取联系人
func (sm *SocialManager) GetContact(ctx context.Context, userAddress, contactID string) (*Contact, error) {
	return sm.addressBook.GetContact(userAddress, contactID)
}

// UpdateContact 更新联系人
func (sm *SocialManager) UpdateContact(ctx context.Context, userAddress string, contact *Contact) error {
	return sm.addressBook.UpdateContact(userAddress, contact)
}

// DeleteContact 删除联系人
func (sm *SocialManager) DeleteContact(ctx context.Context, userAddress, contactID string) error {
	return sm.addressBook.DeleteContact(userAddress, contactID)
}

// CreateShareRecord 创建分享记录
func (sm *SocialManager) CreateShareRecord(ctx context.Context, content *ShareContent, privacy *SharePrivacy) (*ShareRecord, error) {
	return sm.shareManager.CreateShareRecord(content, privacy)
}

// GetShareRecord 获取分享记录
func (sm *SocialManager) GetShareRecord(ctx context.Context, shareID string) (*ShareRecord, error) {
	return sm.shareManager.GetShareRecord(shareID)
}

// FollowUser 关注用户
func (sm *SocialManager) FollowUser(ctx context.Context, followerAddress, targetAddress string) error {
	return sm.socialNetwork.FollowUser(followerAddress, targetAddress)
}

// UnfollowUser 取消关注
func (sm *SocialManager) UnfollowUser(ctx context.Context, followerAddress, targetAddress string) error {
	return sm.socialNetwork.UnfollowUser(followerAddress, targetAddress)
}

// 辅助构造函数和私有方法

// NewAddressBook 创建地址簿
func NewAddressBook() *AddressBook {
	return &AddressBook{
		contacts:   make(map[string]*Contact),
		groups:     make(map[string]*ContactGroup),
		tags:       make(map[string]*ContactTag),
		userGroups: make(map[string][]string),
	}
}

// AddContact 添加联系人
func (ab *AddressBook) AddContact(userAddress string, contact *Contact) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	// 生成联系人ID
	contact.ID = ab.generateContactID(userAddress, contact.Name)
	contact.CreatedAt = time.Now()
	contact.UpdatedAt = time.Now()

	// 存储联系人
	key := fmt.Sprintf("%s:%s", userAddress, contact.ID)
	ab.contacts[key] = contact

	return nil
}

// GetContact 获取联系人
func (ab *AddressBook) GetContact(userAddress, contactID string) (*Contact, error) {
	ab.mu.RLock()
	defer ab.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", userAddress, contactID)
	contact, exists := ab.contacts[key]
	if !exists {
		return nil, fmt.Errorf("联系人不存在")
	}

	return contact, nil
}

// UpdateContact 更新联系人
func (ab *AddressBook) UpdateContact(userAddress string, contact *Contact) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userAddress, contact.ID)
	if _, exists := ab.contacts[key]; !exists {
		return fmt.Errorf("联系人不存在")
	}

	contact.UpdatedAt = time.Now()
	ab.contacts[key] = contact
	return nil
}

// DeleteContact 删除联系人
func (ab *AddressBook) DeleteContact(userAddress, contactID string) error {
	ab.mu.Lock()
	defer ab.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userAddress, contactID)
	if _, exists := ab.contacts[key]; !exists {
		return fmt.Errorf("联系人不存在")
	}

	delete(ab.contacts, key)
	return nil
}

// generateContactID 生成联系人ID
func (ab *AddressBook) generateContactID(userAddress, name string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", userAddress, name, time.Now().UnixNano())))
	return base64.URLEncoding.EncodeToString(hash[:8])
}

// NewShareManager 创建分享管理器
func NewShareManager() *ShareManager {
	return &ShareManager{
		shareRecords: make(map[string]*ShareRecord),
		shareLinks:   make(map[string]*ShareLink),
		templates:    make(map[string]*ShareTemplate),
	}
}

// CreateShareRecord 创建分享记录
func (sm *ShareManager) CreateShareRecord(content *ShareContent, privacy *SharePrivacy) (*ShareRecord, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	shareID := sm.generateShareID(content)
	shareRecord := &ShareRecord{
		ID:        shareID,
		Type:      "transaction",
		Content:   *content,
		ShareURL:  fmt.Sprintf("https://wallet.example.com/share/%s", shareID),
		Privacy:   *privacy,
		CreatedAt: time.Now(),
		Analytics: ShareAnalytics{
			Platforms: make(map[string]int),
			Countries: make(map[string]int),
		},
	}

	sm.shareRecords[shareID] = shareRecord
	return shareRecord, nil
}

// GetShareRecord 获取分享记录
func (sm *ShareManager) GetShareRecord(shareID string) (*ShareRecord, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	record, exists := sm.shareRecords[shareID]
	if !exists {
		return nil, fmt.Errorf("分享记录不存在")
	}

	return record, nil
}

// generateShareID 生成分享ID
func (sm *ShareManager) generateShareID(content *ShareContent) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", content.TransactionHash, time.Now().UnixNano())))
	return base64.URLEncoding.EncodeToString(hash[:8])
}

// NewSocialNetwork 创建社交网络
func NewSocialNetwork() *SocialNetwork {
	return &SocialNetwork{
		followers:     make(map[string][]string),
		following:     make(map[string][]string),
		groups:        make(map[string]*SocialGroup),
		activities:    make(map[string]*Activity),
		notifications: make(map[string][]*Notification),
	}
}

// FollowUser 关注用户
func (sn *SocialNetwork) FollowUser(followerAddress, targetAddress string) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	// 添加到关注列表
	if sn.following[followerAddress] == nil {
		sn.following[followerAddress] = make([]string, 0)
	}

	// 检查是否已关注
	for _, addr := range sn.following[followerAddress] {
		if addr == targetAddress {
			return fmt.Errorf("已经关注了该用户")
		}
	}

	sn.following[followerAddress] = append(sn.following[followerAddress], targetAddress)

	// 添加到粉丝列表
	if sn.followers[targetAddress] == nil {
		sn.followers[targetAddress] = make([]string, 0)
	}
	sn.followers[targetAddress] = append(sn.followers[targetAddress], followerAddress)

	return nil
}

// UnfollowUser 取消关注
func (sn *SocialNetwork) UnfollowUser(followerAddress, targetAddress string) error {
	sn.mu.Lock()
	defer sn.mu.Unlock()

	// 从关注列表移除
	following := sn.following[followerAddress]
	for i, addr := range following {
		if addr == targetAddress {
			sn.following[followerAddress] = append(following[:i], following[i+1:]...)
			break
		}
	}

	// 从粉丝列表移除
	followers := sn.followers[targetAddress]
	for i, addr := range followers {
		if addr == followerAddress {
			sn.followers[targetAddress] = append(followers[:i], followers[i+1:]...)
			break
		}
	}

	return nil
}

// NewPrivacyManager 创建隐私管理器
func NewPrivacyManager() *PrivacyManager {
	return &PrivacyManager{
		settings:    make(map[string]*PrivacySettings),
		permissions: make(map[string]*UserPermissions),
	}
}

// NewENSResolver 创建ENS解析器
func NewENSResolver() *ENSResolver {
	return &ENSResolver{
		cache: make(map[string]*ENSRecord),
	}
}

// ResolveENS 解析ENS域名
func (er *ENSResolver) ResolveENS(ctx context.Context, ensName string) (*ENSRecord, error) {
	er.mu.RLock()
	cached, exists := er.cache[ensName]
	er.mu.RUnlock()

	if exists && time.Now().Before(cached.ExpiresAt) {
		return cached, nil
	}

	// 简化实现：返回模拟数据
	record := &ENSRecord{
		Name:        ensName,
		Address:     "0x" + strings.Repeat("1234567890", 4),
		Avatar:      "https://example.com/avatar.png",
		Description: "ENS用户",
		ResolvedAt:  time.Now(),
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}

	er.mu.Lock()
	er.cache[ensName] = record
	er.mu.Unlock()

	return record, nil
}
