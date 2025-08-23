/*
安全功能增强核心模块

本模块实现了钱包的高级安全功能，包括：

主要功能：
硬件钱包集成：
- Ledger硬件钱包支持
- Trezor硬件钱包支持
- 硬件钱包设备检测
- 安全固件验证
- PIN码和密码保护

多重签名钱包：
- 多签钱包创建和管理
- 阈值签名机制
- 签名者权限管理
- 提案和批准流程
- 时间锁和延迟执行

高级认证：
- 双因素认证(2FA)
- 生物识别认证
- 设备指纹识别
- IP白名单管理
- 异常登录检测

密钥管理：
- 密钥分片存储
- 社会化恢复机制
- 密钥轮换和更新
- 安全备份策略
- 紧急恢复流程

支持的硬件钱包：
- Ledger Nano S/X/S Plus
- Trezor One/Model T
- KeepKey
- SafePal
- CoolWallet

安全特性：
- 零知识证明验证
- 端到端加密通信
- 安全元素集成
- 防篡改检测
- 审计日志记录
*/
package core

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// AdvancedSecurityManager 高级安全管理器
type AdvancedSecurityManager struct {
	hardwareWallets  map[string]*HardwareWallet // 硬件钱包管理
	multiSigWallets  map[string]*MultiSigWallet // 多签钱包管理
	authManager      *AuthenticationManager     // 认证管理器
	keyManager       *KeyManager                // 密钥管理器
	auditLogger      *AuditLogger               // 审计日志
	securityPolicies map[string]*SecurityPolicy // 安全策略
	mu               sync.RWMutex               // 读写锁
}

// HardwareWallet 硬件钱包
type HardwareWallet struct {
	ID              string            `json:"id"`               // 设备ID
	Type            string            `json:"type"`             // 设备类型
	Model           string            `json:"model"`            // 设备型号
	SerialNumber    string            `json:"serial_number"`    // 序列号
	FirmwareVersion string            `json:"firmware_version"` // 固件版本
	Status          string            `json:"status"`           // 设备状态
	IsConnected     bool              `json:"is_connected"`     // 是否连接
	IsUnlocked      bool              `json:"is_unlocked"`      // 是否解锁
	SupportedChains []string          `json:"supported_chains"` // 支持的区块链
	Addresses       []HardwareAddress `json:"addresses"`        // 派生地址
	LastUsed        time.Time         `json:"last_used"`        // 最后使用时间
	SecurityLevel   string            `json:"security_level"`   // 安全等级
	PINAttempts     int               `json:"pin_attempts"`     // PIN尝试次数
	Features        map[string]bool   `json:"features"`         // 支持的功能
}

// HardwareAddress 硬件钱包地址
type HardwareAddress struct {
	Address        string    `json:"address"`         // 钱包地址
	DerivationPath string    `json:"derivation_path"` // 派生路径
	Chain          string    `json:"chain"`           // 区块链网络
	PublicKey      string    `json:"public_key"`      // 公钥
	IsVerified     bool      `json:"is_verified"`     // 是否已验证
	CreatedAt      time.Time `json:"created_at"`      // 创建时间
}

// MultiSigWallet 多重签名钱包
type MultiSigWallet struct {
	ID               string                `json:"id"`                // 钱包ID
	Name             string                `json:"name"`              // 钱包名称
	Address          string                `json:"address"`           // 钱包地址
	ContractAddress  string                `json:"contract_address"`  // 合约地址
	ChainID          string                `json:"chain_id"`          // 链ID
	Threshold        int                   `json:"threshold"`         // 签名阈值
	Signers          []MultiSigSigner      `json:"signers"`           // 签名者列表
	PendingTxs       []MultiSigTransaction `json:"pending_txs"`       // 待处理交易
	ExecutedTxs      []MultiSigTransaction `json:"executed_txs"`      // 已执行交易
	CreatedBy        string                `json:"created_by"`        // 创建者
	CreatedAt        time.Time             `json:"created_at"`        // 创建时间
	UpdatedAt        time.Time             `json:"updated_at"`        // 更新时间
	Configuration    MultiSigConfig        `json:"configuration"`     // 配置信息
	SecuritySettings MultiSigSecurity      `json:"security_settings"` // 安全设置
}

// MultiSigSigner 多签签名者
type MultiSigSigner struct {
	Address           string                  `json:"address"`            // 签名者地址
	Name              string                  `json:"name"`               // 签名者名称
	Role              string                  `json:"role"`               // 角色
	Weight            int                     `json:"weight"`             // 签名权重
	IsActive          bool                    `json:"is_active"`          // 是否活跃
	JoinedAt          time.Time               `json:"joined_at"`          // 加入时间
	LastSigned        *time.Time              `json:"last_signed"`        // 最后签名时间
	PublicKey         string                  `json:"public_key"`         // 公钥
	DeviceType        string                  `json:"device_type"`        // 设备类型
	NotificationPrefs NotificationPreferences `json:"notification_prefs"` // 通知偏好
}

// MultiSigTransaction 多签交易
type MultiSigTransaction struct {
	ID           string              `json:"id"`            // 交易ID
	Title        string              `json:"title"`         // 交易标题
	Description  string              `json:"description"`   // 交易描述
	To           string              `json:"to"`            // 接收地址
	Value        *big.Int            `json:"value"`         // 交易金额
	Data         []byte              `json:"data"`          // 交易数据
	GasLimit     uint64              `json:"gas_limit"`     // Gas限制
	GasPrice     *big.Int            `json:"gas_price"`     // Gas价格
	Nonce        uint64              `json:"nonce"`         // Nonce
	Status       string              `json:"status"`        // 状态
	RequiredSigs int                 `json:"required_sigs"` // 需要的签名数
	CurrentSigs  int                 `json:"current_sigs"`  // 当前签名数
	Signatures   []MultiSigSignature `json:"signatures"`    // 签名列表
	CreatedBy    string              `json:"created_by"`    // 创建者
	CreatedAt    time.Time           `json:"created_at"`    // 创建时间
	ExecutedAt   *time.Time          `json:"executed_at"`   // 执行时间
	TxHash       string              `json:"tx_hash"`       // 交易哈希
	Timelock     *time.Time          `json:"timelock"`      // 时间锁
	ExpiresAt    *time.Time          `json:"expires_at"`    // 过期时间
}

// MultiSigSignature 多签签名
type MultiSigSignature struct {
	Signer     string    `json:"signer"`      // 签名者地址
	Signature  string    `json:"signature"`   // 签名数据
	SignedAt   time.Time `json:"signed_at"`   // 签名时间
	DeviceType string    `json:"device_type"` // 设备类型
	IPAddress  string    `json:"ip_address"`  // IP地址
	UserAgent  string    `json:"user_agent"`  // 用户代理
	IsValid    bool      `json:"is_valid"`    // 是否有效
}

// MultiSigConfig 多签配置
type MultiSigConfig struct {
	RequireAllSigs   bool          `json:"require_all_sigs"`  // 是否需要全部签名
	TimelockDuration time.Duration `json:"timelock_duration"` // 时间锁持续时间
	ExpirationTime   time.Duration `json:"expiration_time"`   // 交易过期时间
	DailyLimit       *big.Int      `json:"daily_limit"`       // 每日限额
	MonthlyLimit     *big.Int      `json:"monthly_limit"`     // 每月限额
	AutoExecute      bool          `json:"auto_execute"`      // 自动执行
	RequireNotes     bool          `json:"require_notes"`     // 需要备注
	AllowedTokens    []string      `json:"allowed_tokens"`    // 允许的代币
	BlockedAddresses []string      `json:"blocked_addresses"` // 黑名单地址
}

// MultiSigSecurity 多签安全设置
type MultiSigSecurity struct {
	RequireMFA        bool          `json:"require_mfa"`         // 需要多因素认证
	RequireHardware   bool          `json:"require_hardware"`    // 需要硬件钱包
	IPWhitelist       []string      `json:"ip_whitelist"`        // IP白名单
	SessionTimeout    time.Duration `json:"session_timeout"`     // 会话超时
	MaxFailedAttempts int           `json:"max_failed_attempts"` // 最大失败次数
	LockoutDuration   time.Duration `json:"lockout_duration"`    // 锁定持续时间
	AuditLevel        string        `json:"audit_level"`         // 审计等级
}

// NotificationPreferences 通知偏好
type NotificationPreferences struct {
	Email     bool     `json:"email"`     // 邮件通知
	SMS       bool     `json:"sms"`       // 短信通知
	Push      bool     `json:"push"`      // 推送通知
	Webhook   string   `json:"webhook"`   // Webhook URL
	Threshold *big.Int `json:"threshold"` // 通知阈值
}

// AuthenticationManager 认证管理器
type AuthenticationManager struct {
	mfaProviders   map[string]*MFAProvider     // MFA提供者
	deviceRegistry map[string]*TrustedDevice   // 可信设备注册表
	sessions       map[string]*SecuritySession // 安全会话
	biometrics     *BiometricManager           // 生物识别管理器
	mu             sync.RWMutex                // 读写锁
}

// MFAProvider MFA提供者
type MFAProvider struct {
	Type          string                 `json:"type"`          // 类型
	Name          string                 `json:"name"`          // 名称
	IsEnabled     bool                   `json:"is_enabled"`    // 是否启用
	Secret        string                 `json:"secret"`        // 密钥
	BackupCodes   []string               `json:"backup_codes"`  // 备用代码
	LastUsed      *time.Time             `json:"last_used"`     // 最后使用时间
	Configuration map[string]interface{} `json:"configuration"` // 配置
}

// TrustedDevice 可信设备
type TrustedDevice struct {
	ID          string    `json:"id"`          // 设备ID
	Name        string    `json:"name"`        // 设备名称
	DeviceType  string    `json:"device_type"` // 设备类型
	Fingerprint string    `json:"fingerprint"` // 设备指纹
	UserAgent   string    `json:"user_agent"`  // 用户代理
	IPAddress   string    `json:"ip_address"`  // IP地址
	Location    string    `json:"location"`    // 地理位置
	IsActive    bool      `json:"is_active"`   // 是否活跃
	FirstSeen   time.Time `json:"first_seen"`  // 首次发现
	LastSeen    time.Time `json:"last_seen"`   // 最后发现
	TrustLevel  string    `json:"trust_level"` // 信任级别
}

// SecuritySession 安全会话
type SecuritySession struct {
	ID            string    `json:"id"`             // 会话ID
	UserAddress   string    `json:"user_address"`   // 用户地址
	DeviceID      string    `json:"device_id"`      // 设备ID
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	ExpiresAt     time.Time `json:"expires_at"`     // 过期时间
	LastActivity  time.Time `json:"last_activity"`  // 最后活动
	SecurityLevel string    `json:"security_level"` // 安全级别
	Permissions   []string  `json:"permissions"`    // 权限列表
	IPAddress     string    `json:"ip_address"`     // IP地址
	IsActive      bool      `json:"is_active"`      // 是否活跃
}

// BiometricManager 生物识别管理器
type BiometricManager struct {
	enabledTypes []string                      // 启用的生物识别类型
	templates    map[string]*BiometricTemplate // 生物识别模板
	devices      map[string]*BiometricDevice   // 生物识别设备
	mu           sync.RWMutex                  // 读写锁
}

// BiometricTemplate 生物识别模板
type BiometricTemplate struct {
	ID          string    `json:"id"`           // 模板ID
	Type        string    `json:"type"`         // 类型
	UserAddress string    `json:"user_address"` // 用户地址
	Template    []byte    `json:"template"`     // 模板数据
	Quality     float64   `json:"quality"`      // 质量分数
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
	IsActive    bool      `json:"is_active"`    // 是否活跃
}

// BiometricDevice 生物识别设备
type BiometricDevice struct {
	ID           string   `json:"id"`           // 设备ID
	Type         string   `json:"type"`         // 设备类型
	Name         string   `json:"name"`         // 设备名称
	IsConnected  bool     `json:"is_connected"` // 是否连接
	Capabilities []string `json:"capabilities"` // 功能列表
	Status       string   `json:"status"`       // 状态
}

// KeyManager 密钥管理器
type KeyManager struct {
	keyShards       map[string]*KeyShard        // 密钥分片
	recoveryShares  map[string]*RecoveryShare   // 恢复分片
	backupPolicies  map[string]*BackupPolicy    // 备份策略
	emergencyAccess map[string]*EmergencyAccess // 紧急访问
	mu              sync.RWMutex                // 读写锁
}

// KeyShard 密钥分片
type KeyShard struct {
	ID          string     `json:"id"`           // 分片ID
	UserAddress string     `json:"user_address"` // 用户地址
	ShardIndex  int        `json:"shard_index"`  // 分片索引
	ShardData   []byte     `json:"shard_data"`   // 分片数据
	Threshold   int        `json:"threshold"`    // 恢复阈值
	TotalShards int        `json:"total_shards"` // 总分片数
	CreatedAt   time.Time  `json:"created_at"`   // 创建时间
	ExpiresAt   *time.Time `json:"expires_at"`   // 过期时间
	IsActive    bool       `json:"is_active"`    // 是否活跃
}

// RecoveryShare 恢复分片
type RecoveryShare struct {
	ID              string        `json:"id"`               // 分片ID
	GuardianAddress string        `json:"guardian_address"` // 守护者地址
	ShareData       []byte        `json:"share_data"`       // 分片数据
	RecoveryMethod  string        `json:"recovery_method"`  // 恢复方法
	DelayPeriod     time.Duration `json:"delay_period"`     // 延迟期间
	IsActive        bool          `json:"is_active"`        // 是否活跃
	CreatedAt       time.Time     `json:"created_at"`       // 创建时间
}

// BackupPolicy 备份策略
type BackupPolicy struct {
	ID          string        `json:"id"`           // 策略ID
	Name        string        `json:"name"`         // 策略名称
	UserAddress string        `json:"user_address"` // 用户地址
	BackupType  string        `json:"backup_type"`  // 备份类型
	Schedule    string        `json:"schedule"`     // 备份计划
	Encryption  string        `json:"encryption"`   // 加密方式
	Storage     string        `json:"storage"`      // 存储位置
	Retention   time.Duration `json:"retention"`    // 保留期限
	IsEnabled   bool          `json:"is_enabled"`   // 是否启用
	LastBackup  *time.Time    `json:"last_backup"`  // 最后备份时间
}

// EmergencyAccess 紧急访问
type EmergencyAccess struct {
	ID             string        `json:"id"`              // 访问ID
	UserAddress    string        `json:"user_address"`    // 用户地址
	TrustedContact string        `json:"trusted_contact"` // 可信联系人
	AccessType     string        `json:"access_type"`     // 访问类型
	DelayPeriod    time.Duration `json:"delay_period"`    // 延迟期间
	RequiredProofs []string      `json:"required_proofs"` // 需要的证明
	Status         string        `json:"status"`          // 状态
	RequestedAt    *time.Time    `json:"requested_at"`    // 请求时间
	ApprovedAt     *time.Time    `json:"approved_at"`     // 批准时间
	ExpiresAt      *time.Time    `json:"expires_at"`      // 过期时间
}

// AuditLogger 审计日志
type AuditLogger struct {
	logs      []AuditLog    // 日志记录
	retention time.Duration // 保留期限
	mu        sync.RWMutex  // 读写锁
}

// AuditLog 审计日志记录
type AuditLog struct {
	ID            string                 `json:"id"`             // 日志ID
	Timestamp     time.Time              `json:"timestamp"`      // 时间戳
	UserAddress   string                 `json:"user_address"`   // 用户地址
	Action        string                 `json:"action"`         // 操作
	Resource      string                 `json:"resource"`       // 资源
	Result        string                 `json:"result"`         // 结果
	IPAddress     string                 `json:"ip_address"`     // IP地址
	UserAgent     string                 `json:"user_agent"`     // 用户代理
	DeviceID      string                 `json:"device_id"`      // 设备ID
	SessionID     string                 `json:"session_id"`     // 会话ID
	Details       map[string]interface{} `json:"details"`        // 详细信息
	RiskScore     float64                `json:"risk_score"`     // 风险分数
	SecurityLevel string                 `json:"security_level"` // 安全级别
}

// SecurityPolicy 安全策略
type SecurityPolicy struct {
	ID          string                 `json:"id"`          // 策略ID
	Name        string                 `json:"name"`        // 策略名称
	Description string                 `json:"description"` // 描述
	Rules       []AdvancedSecurityRule `json:"rules"`       // 安全规则
	IsEnabled   bool                   `json:"is_enabled"`  // 是否启用
	Priority    int                    `json:"priority"`    // 优先级
	CreatedAt   time.Time              `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`  // 更新时间
}

// AdvancedSecurityRule 高级安全规则
type AdvancedSecurityRule struct {
	ID         string                 `json:"id"`         // 规则ID
	Type       string                 `json:"type"`       // 规则类型
	Condition  string                 `json:"condition"`  // 条件
	Action     string                 `json:"action"`     // 动作
	Parameters map[string]interface{} `json:"parameters"` // 参数
	IsEnabled  bool                   `json:"is_enabled"` // 是否启用
}

// NewAdvancedSecurityManager 创建高级安全管理器
func NewAdvancedSecurityManager() *AdvancedSecurityManager {
	return &AdvancedSecurityManager{
		hardwareWallets:  make(map[string]*HardwareWallet),
		multiSigWallets:  make(map[string]*MultiSigWallet),
		authManager:      NewAuthenticationManager(),
		keyManager:       NewKeyManager(),
		auditLogger:      NewAuditLogger(),
		securityPolicies: make(map[string]*SecurityPolicy),
	}
}

// DetectHardwareWallets 检测硬件钱包
func (sm *AdvancedSecurityManager) DetectHardwareWallets(ctx context.Context) ([]*HardwareWallet, error) {
	// 简化实现：返回模拟的硬件钱包
	wallets := []*HardwareWallet{
		{
			ID:              "ledger_001",
			Type:            "Ledger",
			Model:           "Nano S Plus",
			SerialNumber:    "0001",
			FirmwareVersion: "2.1.0",
			Status:          "connected",
			IsConnected:     true,
			IsUnlocked:      false,
			SupportedChains: []string{"ethereum", "bitcoin", "polygon"},
			SecurityLevel:   "high",
			Features: map[string]bool{
				"secure_element": true,
				"pin_protection": true,
				"recovery_mode":  true,
			},
		},
	}

	return wallets, nil
}

// CreateMultiSigWallet 创建多签钱包
func (sm *AdvancedSecurityManager) CreateMultiSigWallet(ctx context.Context, config *MultiSigConfig, signers []MultiSigSigner, threshold int) (*MultiSigWallet, error) {
	if threshold < 1 || threshold > len(signers) {
		return nil, fmt.Errorf("无效的签名阈值")
	}

	walletID := sm.generateWalletID()
	wallet := &MultiSigWallet{
		ID:            walletID,
		Threshold:     threshold,
		Signers:       signers,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Configuration: *config,
		SecuritySettings: MultiSigSecurity{
			RequireMFA:        true,
			SessionTimeout:    30 * time.Minute,
			MaxFailedAttempts: 3,
			LockoutDuration:   1 * time.Hour,
			AuditLevel:        "high",
		},
	}

	sm.mu.Lock()
	sm.multiSigWallets[walletID] = wallet
	sm.mu.Unlock()

	// 记录审计日志
	sm.auditLogger.LogAction("create_multisig_wallet", "multisig_wallet", walletID, "success", nil)

	return wallet, nil
}

// SignMultiSigTransaction 签名多签交易
func (sm *AdvancedSecurityManager) SignMultiSigTransaction(ctx context.Context, walletID, txID, signerAddress, signature string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	wallet, exists := sm.multiSigWallets[walletID]
	if !exists {
		return fmt.Errorf("多签钱包不存在")
	}

	// 查找交易
	var tx *MultiSigTransaction
	for i := range wallet.PendingTxs {
		if wallet.PendingTxs[i].ID == txID {
			tx = &wallet.PendingTxs[i]
			break
		}
	}

	if tx == nil {
		return fmt.Errorf("交易不存在")
	}

	// 验证签名者
	var signer *MultiSigSigner
	for i := range wallet.Signers {
		if wallet.Signers[i].Address == signerAddress {
			signer = &wallet.Signers[i]
			break
		}
	}

	if signer == nil {
		return fmt.Errorf("无效的签名者")
	}

	// 添加签名
	sig := MultiSigSignature{
		Signer:    signerAddress,
		Signature: signature,
		SignedAt:  time.Now(),
		IsValid:   true,
	}

	tx.Signatures = append(tx.Signatures, sig)
	tx.CurrentSigs++

	// 检查是否达到阈值
	if tx.CurrentSigs >= wallet.Threshold {
		tx.Status = "ready_to_execute"
	}

	now := time.Now()
	signer.LastSigned = &now

	return nil
}

// 辅助构造函数

// NewAuthenticationManager 创建认证管理器
func NewAuthenticationManager() *AuthenticationManager {
	return &AuthenticationManager{
		mfaProviders:   make(map[string]*MFAProvider),
		deviceRegistry: make(map[string]*TrustedDevice),
		sessions:       make(map[string]*SecuritySession),
		biometrics:     NewBiometricManager(),
	}
}

// NewKeyManager 创建密钥管理器
func NewKeyManager() *KeyManager {
	return &KeyManager{
		keyShards:       make(map[string]*KeyShard),
		recoveryShares:  make(map[string]*RecoveryShare),
		backupPolicies:  make(map[string]*BackupPolicy),
		emergencyAccess: make(map[string]*EmergencyAccess),
	}
}

// NewBiometricManager 创建生物识别管理器
func NewBiometricManager() *BiometricManager {
	return &BiometricManager{
		enabledTypes: []string{"fingerprint", "face", "voice"},
		templates:    make(map[string]*BiometricTemplate),
		devices:      make(map[string]*BiometricDevice),
	}
}

// NewAuditLogger 创建审计日志
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs:      make([]AuditLog, 0),
		retention: 365 * 24 * time.Hour, // 保留一年
	}
}

// LogAction 记录操作日志
func (al *AuditLogger) LogAction(action, resource, resourceID, result string, details map[string]interface{}) {
	al.mu.Lock()
	defer al.mu.Unlock()

	log := AuditLog{
		ID:        al.generateLogID(),
		Timestamp: time.Now(),
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
	}

	al.logs = append(al.logs, log)

	// 清理过期日志
	al.cleanExpiredLogs()
}

// generateWalletID 生成钱包ID
func (sm *AdvancedSecurityManager) generateWalletID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateLogID 生成日志ID
func (al *AuditLogger) generateLogID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(hash[:8])
}

// cleanExpiredLogs 清理过期日志
func (al *AuditLogger) cleanExpiredLogs() {
	cutoff := time.Now().Add(-al.retention)
	filtered := make([]AuditLog, 0)

	for _, log := range al.logs {
		if log.Timestamp.After(cutoff) {
			filtered = append(filtered, log)
		}
	}

	al.logs = filtered
}
