/**
 * 数据模型定义包
 *
 * 这个包定义了钱包应用的所有数据模型，使用GORM作为ORM框架
 *
 * 后端学习要点：
 * 1. GORM标签 - 定义数据库字段映射和约束
 * 2. JSON标签 - 定义API响应时的JSON字段名
 * 3. 结构体嵌入 - 复用通用字段如时间戳
 * 4. 指针类型 - 处理可空字段
 * 5. 关联关系 - 定义表之间的外键关系
 */
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// =============================================================================
// 通用模型
// =============================================================================

/**
 * 基础模型结构
 * 包含所有数据表的通用字段
 */
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"` // 软删除
}

/**
 * JSON类型支持
 * 用于存储JSON格式的数据到数据库
 */
type JSON map[string]interface{}

// 实现driver.Valuer接口，用于写入数据库
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.Marshal(j)
}

// 实现sql.Scanner接口，用于从数据库读取
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-bytes into JSON")
	}

	return json.Unmarshal(bytes, j)
}

// =============================================================================
// 用户相关模型
// =============================================================================

/**
 * 用户模型
 * 存储用户基本信息和认证数据
 */
type User struct {
	BaseModel

	// 基本信息
	Username     string  `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string  `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string  `gorm:"size:255;not null" json:"-"` // 不在JSON中返回
	Salt         string  `gorm:"size:255;not null" json:"-"` // 不在JSON中返回
	AvatarURL    *string `gorm:"size:500" json:"avatar_url,omitempty"`

	// 状态信息
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// 关联关系
	Sessions       []UserSession   `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	WatchAddresses []WatchAddress  `gorm:"foreignKey:UserID" json:"watch_addresses,omitempty"`
	Wallets        []UserWallet    `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
	Preferences    *UserPreference `gorm:"foreignKey:UserID" json:"preferences,omitempty"`
	ActivityLogs   []ActivityLog   `gorm:"foreignKey:UserID" json:"activity_logs,omitempty"`
}

/**
 * 用户会话模型
 * 管理用户登录会话和JWT令牌
 */
type UserSession struct {
	BaseModel

	UserID       uint      `gorm:"not null;index" json:"user_id"`
	SessionToken string    `gorm:"uniqueIndex;size:255;not null" json:"session_token"`
	RefreshToken string    `gorm:"uniqueIndex;size:255;not null" json:"refresh_token"`
	DeviceInfo   JSON      `gorm:"type:jsonb" json:"device_info,omitempty"`
	IPAddress    string    `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent    string    `gorm:"type:text" json:"user_agent,omitempty"`
	ExpiresAt    time.Time `gorm:"not null;index" json:"expires_at"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

/**
 * 用户偏好设置模型
 * 存储用户的个性化配置
 */
type UserPreference struct {
	BaseModel

	UserID          uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	DefaultCurrency string `gorm:"size:10;default:'USD'" json:"default_currency"`
	Theme           string `gorm:"size:20;default:'light'" json:"theme"` // light, dark, auto
	Language        string `gorm:"size:10;default:'zh-CN'" json:"language"`
	Notifications   JSON   `gorm:"type:jsonb;default:'{}'" json:"notifications"`
	DisplaySettings JSON   `gorm:"type:jsonb;default:'{}'" json:"display_settings"`
	PrivacySettings JSON   `gorm:"type:jsonb;default:'{}'" json:"privacy_settings"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// =============================================================================
// 地址和钱包相关模型
// =============================================================================

/**
 * 观察地址模型
 * 存储用户添加的监控地址
 */
type WatchAddress struct {
	BaseModel

	UserID              uint       `gorm:"not null;index" json:"user_id"`
	Address             string     `gorm:"size:42;not null;index" json:"address"`
	Label               *string    `gorm:"size:100" json:"label,omitempty"`
	NetworkID           int        `gorm:"default:1;index" json:"network_id"`         // 1=以太坊主网
	AddressType         string     `gorm:"size:20;default:'EOA'" json:"address_type"` // EOA, Contract, MultiSig
	Tags                JSON       `gorm:"type:jsonb" json:"tags,omitempty"`
	Notes               *string    `gorm:"type:text" json:"notes,omitempty"`
	IsFavorite          bool       `gorm:"default:false" json:"is_favorite"`
	NotificationEnabled bool       `gorm:"default:true" json:"notification_enabled"`
	BalanceCache        *string    `gorm:"type:decimal(36,18)" json:"balance_cache,omitempty"`
	LastActivityAt      *time.Time `json:"last_activity_at,omitempty"`

	// 关联
	User           User                    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BalanceHistory []AddressBalanceHistory `gorm:"foreignKey:WatchAddressID" json:"balance_history,omitempty"`
}

/**
 * 用户钱包记录模型
 * 记录用户导入/创建的钱包(不存储私钥)
 */
type UserWallet struct {
	BaseModel

	UserID         uint       `gorm:"not null;index" json:"user_id"`
	Address        string     `gorm:"size:42;not null;index" json:"address"`
	WalletName     string     `gorm:"size:100;not null" json:"wallet_name"`
	WalletType     string     `gorm:"size:20;default:'imported'" json:"wallet_type"` // imported, created, hardware
	DerivationPath *string    `gorm:"size:100" json:"derivation_path,omitempty"`
	NetworkID      int        `gorm:"default:1" json:"network_id"`
	IsPrimary      bool       `gorm:"default:false" json:"is_primary"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

/**
 * 地址余额历史模型
 * 记录观察地址的余额变化
 */
type AddressBalanceHistory struct {
	BaseModel

	WatchAddressID uint      `gorm:"not null;index" json:"watch_address_id"`
	Balance        string    `gorm:"type:decimal(36,18);not null" json:"balance"`
	TokenAddress   *string   `gorm:"size:42;index" json:"token_address,omitempty"` // NULL表示主币
	TokenSymbol    *string   `gorm:"size:20" json:"token_symbol,omitempty"`
	BlockNumber    *uint64   `json:"block_number,omitempty"`
	RecordedAt     time.Time `gorm:"index;default:CURRENT_TIMESTAMP" json:"recorded_at"`

	// 关联
	WatchAddress WatchAddress `gorm:"foreignKey:WatchAddressID" json:"watch_address,omitempty"`
}

// =============================================================================
// 日志和审计模型
// =============================================================================

/**
 * 用户活动日志模型
 * 记录用户重要操作，用于安全审计
 */
type ActivityLog struct {
	BaseModel

	UserID       *uint   `gorm:"index" json:"user_id,omitempty"` // 可为空，支持匿名操作记录
	Action       string  `gorm:"size:50;not null;index" json:"action"`
	ResourceType *string `gorm:"size:50" json:"resource_type,omitempty"`
	ResourceID   *string `gorm:"size:100" json:"resource_id,omitempty"`
	Details      JSON    `gorm:"type:jsonb" json:"details,omitempty"`
	IPAddress    *string `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent    *string `gorm:"type:text" json:"user_agent,omitempty"`
	Status       string  `gorm:"size:20;default:'success'" json:"status"` // success, failed, pending

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// =============================================================================
// 模型方法
// =============================================================================

/**
 * User模型的方法
 */

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 确保用户名和邮箱转为小写
	// u.Username = strings.ToLower(u.Username)
	// u.Email = strings.ToLower(u.Email)
	return nil
}

// GetActiveWallets 获取用户的活跃钱包
func (u *User) GetActiveWallets(db *gorm.DB) ([]UserWallet, error) {
	var wallets []UserWallet
	err := db.Where("user_id = ? AND deleted_at IS NULL", u.ID).
		Order("is_primary DESC, last_used_at DESC").
		Find(&wallets).Error
	return wallets, err
}

// GetWatchAddresses 获取用户的观察地址
func (u *User) GetWatchAddresses(db *gorm.DB, networkID *int) ([]WatchAddress, error) {
	var addresses []WatchAddress
	query := db.Where("user_id = ? AND deleted_at IS NULL", u.ID)

	if networkID != nil {
		query = query.Where("network_id = ?", *networkID)
	}

	err := query.Order("is_favorite DESC, created_at DESC").Find(&addresses).Error
	return addresses, err
}

/**
 * WatchAddress模型的方法
 */

// UpdateBalance 更新地址余额缓存
func (wa *WatchAddress) UpdateBalance(db *gorm.DB, balance string, blockNumber *uint64) error {
	// 更新缓存
	wa.BalanceCache = &balance
	wa.LastActivityAt = &time.Time{}
	*wa.LastActivityAt = time.Now()

	// 保存到数据库
	if err := db.Save(wa).Error; err != nil {
		return err
	}

	// 记录历史
	history := AddressBalanceHistory{
		WatchAddressID: wa.ID,
		Balance:        balance,
		BlockNumber:    blockNumber,
		RecordedAt:     time.Now(),
	}

	return db.Create(&history).Error
}

/**
 * ActivityLog模型的方法
 */

// LogUserAction 记录用户操作
func LogUserAction(db *gorm.DB, userID *uint, action string, resourceType, resourceID *string, details JSON, ipAddress, userAgent *string) error {
	log := ActivityLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Status:       "success",
	}

	return db.Create(&log).Error
}
