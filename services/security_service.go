/*
安全功能增强业务服务层

本文件实现了安全功能增强的业务服务层，提供硬件钱包集成、多重签名管理、高级认证等服务。
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

// SecurityService 安全功能服务
type SecurityService struct {
	securityManager *core.AdvancedSecurityManager   // 高级安全管理器
	walletService   *WalletService                  // 钱包服务
	activeSessions  map[string]*SecuritySessionInfo // 活跃安全会话
	mu              sync.RWMutex                    // 读写锁
}

// SecuritySessionInfo 安全会话信息
type SecuritySessionInfo struct {
	SessionID     string    `json:"session_id"`     // 会话ID
	UserAddress   string    `json:"user_address"`   // 用户地址
	SecurityLevel string    `json:"security_level"` // 安全级别
	AuthMethods   []string  `json:"auth_methods"`   // 认证方法
	CreatedAt     time.Time `json:"created_at"`     // 创建时间
	ExpiresAt     time.Time `json:"expires_at"`     // 过期时间
	LastActivity  time.Time `json:"last_activity"`  // 最后活动
}

// HardwareWalletRequest 硬件钱包请求
type HardwareWalletRequest struct {
	Action         string                 `json:"action" binding:"required"` // 操作类型
	DeviceType     string                 `json:"device_type"`               // 设备类型
	DerivationPath string                 `json:"derivation_path"`           // 派生路径
	Data           map[string]interface{} `json:"data"`                      // 附加数据
}

// HardwareWalletResponse 硬件钱包响应
type HardwareWalletResponse struct {
	Success    bool                   `json:"success"`     // 是否成功
	Message    string                 `json:"message"`     // 消息
	Data       map[string]interface{} `json:"data"`        // 响应数据
	DeviceInfo *core.HardwareWallet   `json:"device_info"` // 设备信息
}

// MultiSigWalletRequest 多签钱包创建请求
type MultiSigWalletRequest struct {
	Name          string                  `json:"name" binding:"required"`
	Threshold     int                     `json:"threshold" binding:"required"`
	Signers       []MultiSigSignerRequest `json:"signers" binding:"required"`
	ChainID       string                  `json:"chain_id" binding:"required"`
	Configuration MultiSigConfigRequest   `json:"configuration"`
}

// MultiSigSignerRequest 多签签名者请求
type MultiSigSignerRequest struct {
	Address           string `json:"address" binding:"required"`
	Name              string `json:"name" binding:"required"`
	Role              string `json:"role"`
	Weight            int    `json:"weight"`
	DeviceType        string `json:"device_type"`
	NotificationEmail string `json:"notification_email"`
}

// MultiSigConfigRequest 多签配置请求
type MultiSigConfigRequest struct {
	RequireAllSigs  bool   `json:"require_all_sigs"`
	TimelockHours   int    `json:"timelock_hours"`
	ExpirationHours int    `json:"expiration_hours"`
	DailyLimit      string `json:"daily_limit"`
	MonthlyLimit    string `json:"monthly_limit"`
	RequireMFA      bool   `json:"require_mfa"`
	RequireHardware bool   `json:"require_hardware"`
}

// MultiSigWalletResponse 多签钱包响应
type MultiSigWalletResponse struct {
	Wallet           *core.MultiSigWallet `json:"wallet"`             // 钱包信息
	ContractAddress  string               `json:"contract_address"`   // 合约地址
	DeploymentTxHash string               `json:"deployment_tx_hash"` // 部署交易哈希
}

// MultiSigTransactionRequest 多签交易请求
type MultiSigTransactionRequest struct {
	WalletID        string `json:"wallet_id" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	To              string `json:"to" binding:"required"`
	Value           string `json:"value"`
	Data            string `json:"data"`
	TokenAddress    string `json:"token_address"`
	TokenAmount     string `json:"token_amount"`
	TimelockHours   int    `json:"timelock_hours"`
	ExpirationHours int    `json:"expiration_hours"`
}

// MultiSigTransactionResponse 多签交易响应
type MultiSigTransactionResponse struct {
	Transaction    *core.MultiSigTransaction `json:"transaction"`     // 交易信息
	RequiredSigs   int                       `json:"required_sigs"`   // 需要的签名数
	PendingSigners []string                  `json:"pending_signers"` // 待签名者
	EstimatedGas   uint64                    `json:"estimated_gas"`   // 预估Gas
}

// SignTransactionRequest 签名交易请求
type SignTransactionRequest struct {
	WalletID      string `json:"wallet_id" binding:"required"`
	TransactionID string `json:"transaction_id" binding:"required"`
	SignerAddress string `json:"signer_address" binding:"required"`
	Signature     string `json:"signature"`
	DeviceType    string `json:"device_type"`
	MFACode       string `json:"mfa_code"`
	Comments      string `json:"comments"`
}

// MFASetupRequest MFA设置请求
type MFASetupRequest struct {
	MFAType     string `json:"mfa_type" binding:"required"` // TOTP, SMS, Email
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	BackupCodes bool   `json:"backup_codes"`
}

// MFASetupResponse MFA设置响应
type MFASetupResponse struct {
	Secret        string   `json:"secret"`         // TOTP密钥
	QRCode        string   `json:"qr_code"`        // 二维码
	BackupCodes   []string `json:"backup_codes"`   // 备用代码
	SetupComplete bool     `json:"setup_complete"` // 设置完成
}

// SecurityAuditRequest 安全审计请求
type SecurityAuditRequest struct {
	UserAddress    string     `json:"user_address"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	ActionTypes    []string   `json:"action_types"`
	SecurityLevels []string   `json:"security_levels"`
	Limit          int        `json:"limit"`
	Offset         int        `json:"offset"`
}

// SecurityAuditResponse 安全审计响应
type SecurityAuditResponse struct {
	Logs         []core.AuditLog `json:"logs"`          // 审计日志
	Summary      *AuditSummary   `json:"summary"`       // 审计摘要
	RiskAnalysis *RiskAnalysis   `json:"risk_analysis"` // 风险分析
	Total        int             `json:"total"`         // 总数
	HasMore      bool            `json:"has_more"`      // 是否有更多
}

// AuditSummary 审计摘要
type AuditSummary struct {
	TotalActions      int    `json:"total_actions"`      // 总操作数
	SuccessfulActions int    `json:"successful_actions"` // 成功操作数
	FailedActions     int    `json:"failed_actions"`     // 失败操作数
	HighRiskActions   int    `json:"high_risk_actions"`  // 高风险操作数
	UniqueIPs         int    `json:"unique_ips"`         // 唯一IP数
	TimeRange         string `json:"time_range"`         // 时间范围
}

// RiskAnalysis 风险分析
type RiskAnalysis struct {
	OverallRisk       string   `json:"overall_risk"`       // 总体风险
	RiskFactors       []string `json:"risk_factors"`       // 风险因素
	Recommendations   []string `json:"recommendations"`    // 建议
	AnomalousPatterns []string `json:"anomalous_patterns"` // 异常模式
	ThreatIndicators  []string `json:"threat_indicators"`  // 威胁指标
}

// NewSecurityService 创建安全功能服务
func NewSecurityService(walletService *WalletService) *SecurityService {
	return &SecurityService{
		securityManager: core.NewAdvancedSecurityManager(),
		walletService:   walletService,
		activeSessions:  make(map[string]*SecuritySessionInfo),
	}
}

// DetectHardwareWallets 检测硬件钱包
func (ss *SecurityService) DetectHardwareWallets(ctx context.Context) ([]*core.HardwareWallet, error) {
	return ss.securityManager.DetectHardwareWallets(ctx)
}

// ProcessHardwareWalletRequest 处理硬件钱包请求
func (ss *SecurityService) ProcessHardwareWalletRequest(ctx context.Context, userAddress string, request *HardwareWalletRequest) (*HardwareWalletResponse, error) {
	response := &HardwareWalletResponse{
		Data: make(map[string]interface{}),
	}

	switch request.Action {
	case "detect":
		wallets, err := ss.DetectHardwareWallets(ctx)
		if err != nil {
			response.Success = false
			response.Message = "检测硬件钱包失败: " + err.Error()
			return response, nil
		}

		response.Success = true
		response.Message = fmt.Sprintf("检测到 %d 个硬件钱包", len(wallets))
		response.Data["wallets"] = wallets

	case "connect":
		// 简化实现：模拟连接成功
		response.Success = true
		response.Message = "硬件钱包连接成功"
		response.Data["status"] = "connected"

	case "get_address":
		// 简化实现：返回示例地址
		if request.DerivationPath == "" {
			request.DerivationPath = "m/44'/60'/0'/0/0"
		}

		address := "0x" + fmt.Sprintf("%040x", time.Now().UnixNano())
		response.Success = true
		response.Message = "地址获取成功"
		response.Data["address"] = address
		response.Data["derivation_path"] = request.DerivationPath

	case "sign":
		// 简化实现：返回模拟签名
		signature := "0x" + fmt.Sprintf("%0128x", time.Now().UnixNano())
		response.Success = true
		response.Message = "签名成功"
		response.Data["signature"] = signature

	default:
		response.Success = false
		response.Message = "不支持的操作: " + request.Action
	}

	return response, nil
}

// CreateMultiSigWallet 创建多签钱包
func (ss *SecurityService) CreateMultiSigWallet(ctx context.Context, userAddress string, request *MultiSigWalletRequest) (*MultiSigWalletResponse, error) {
	// 验证参数
	if request.Threshold < 1 || request.Threshold > len(request.Signers) {
		return nil, fmt.Errorf("无效的签名阈值")
	}

	// 构建签名者列表
	signers := make([]core.MultiSigSigner, len(request.Signers))
	for i, signerReq := range request.Signers {
		if !ss.walletService.IsValidAddress(signerReq.Address) {
			return nil, fmt.Errorf("无效的签名者地址: %s", signerReq.Address)
		}

		signers[i] = core.MultiSigSigner{
			Address:    signerReq.Address,
			Name:       signerReq.Name,
			Role:       signerReq.Role,
			Weight:     signerReq.Weight,
			IsActive:   true,
			JoinedAt:   time.Now(),
			DeviceType: signerReq.DeviceType,
		}
	}

	// 构建配置
	config := &core.MultiSigConfig{
		RequireAllSigs:   request.Configuration.RequireAllSigs,
		TimelockDuration: time.Duration(request.Configuration.TimelockHours) * time.Hour,
		ExpirationTime:   time.Duration(request.Configuration.ExpirationHours) * time.Hour,
		AutoExecute:      false,
		RequireNotes:     true,
	}

	// 解析限额
	if request.Configuration.DailyLimit != "" {
		dailyLimit, ok := new(big.Int).SetString(request.Configuration.DailyLimit, 10)
		if ok {
			config.DailyLimit = dailyLimit
		}
	}

	if request.Configuration.MonthlyLimit != "" {
		monthlyLimit, ok := new(big.Int).SetString(request.Configuration.MonthlyLimit, 10)
		if ok {
			config.MonthlyLimit = monthlyLimit
		}
	}

	// 创建多签钱包
	wallet, err := ss.securityManager.CreateMultiSigWallet(ctx, config, signers, request.Threshold)
	if err != nil {
		return nil, fmt.Errorf("创建多签钱包失败: %w", err)
	}

	// 设置钱包名称和链ID
	wallet.Name = request.Name
	wallet.ChainID = request.ChainID

	// 生成合约地址（简化实现）
	contractAddress := "0x" + fmt.Sprintf("%040x", time.Now().UnixNano())
	wallet.ContractAddress = contractAddress

	response := &MultiSigWalletResponse{
		Wallet:           wallet,
		ContractAddress:  contractAddress,
		DeploymentTxHash: "0x" + fmt.Sprintf("%064x", time.Now().UnixNano()),
	}
	return response, nil
}

// CreateMultiSigTransaction 创建多签交易
func (ss *SecurityService) CreateMultiSigTransaction(ctx context.Context, userAddress string, request *MultiSigTransactionRequest) (*MultiSigTransactionResponse, error) {
	// 验证钱包ID
	// 这里简化实现，实际应该从数据库获取钱包信息

	// 解析金额
	value := big.NewInt(0)
	if request.Value != "" {
		var ok bool
		value, ok = new(big.Int).SetString(request.Value, 10)
		if !ok {
			return nil, fmt.Errorf("无效的金额格式")
		}
	}

	// 创建交易
	txID := ss.generateTransactionID()
	transaction := &core.MultiSigTransaction{
		ID:           txID,
		Title:        request.Title,
		Description:  request.Description,
		To:           request.To,
		Value:        value,
		Status:       "pending",
		RequiredSigs: 2, // 简化实现
		CurrentSigs:  0,
		CreatedBy:    userAddress,
		CreatedAt:    time.Now(),
		Signatures:   make([]core.MultiSigSignature, 0),
	}

	// 设置时间锁
	if request.TimelockHours > 0 {
		timelock := time.Now().Add(time.Duration(request.TimelockHours) * time.Hour)
		transaction.Timelock = &timelock
	}

	// 设置过期时间
	if request.ExpirationHours > 0 {
		expires := time.Now().Add(time.Duration(request.ExpirationHours) * time.Hour)
		transaction.ExpiresAt = &expires
	}

	response := &MultiSigTransactionResponse{
		Transaction:    transaction,
		RequiredSigs:   transaction.RequiredSigs,
		PendingSigners: []string{"0x1111111111111111111111111111111111111111", "0x2222222222222222222222222222222222222222"},
		EstimatedGas:   21000,
	}

	return response, nil
}

// SignMultiSigTransaction 签名多签交易
func (ss *SecurityService) SignMultiSigTransaction(ctx context.Context, request *SignTransactionRequest) error {
	// 验证MFA（如果需要）
	if request.MFACode != "" {
		valid := ss.verifyMFACode(request.SignerAddress, request.MFACode)
		if !valid {
			return fmt.Errorf("MFA验证失败")
		}
	}

	// 执行签名
	err := ss.securityManager.SignMultiSigTransaction(
		ctx,
		request.WalletID,
		request.TransactionID,
		request.SignerAddress,
		request.Signature,
	)

	if err != nil {
		return fmt.Errorf("签名失败: %w", err)
	}

	return nil
}

// SetupMFA 设置多因素认证
func (ss *SecurityService) SetupMFA(ctx context.Context, userAddress string, request *MFASetupRequest) (*MFASetupResponse, error) {
	response := &MFASetupResponse{}

	switch request.MFAType {
	case "TOTP":
		// 生成TOTP密钥
		secret := ss.generateTOTPSecret()
		qrCode := ss.generateQRCode(userAddress, secret)

		response.Secret = secret
		response.QRCode = qrCode
		response.SetupComplete = false

	case "SMS":
		if request.PhoneNumber == "" {
			return nil, fmt.Errorf("手机号码不能为空")
		}
		// 发送验证短信
		response.SetupComplete = false

	case "Email":
		if request.Email == "" {
			return nil, fmt.Errorf("邮箱地址不能为空")
		}
		// 发送验证邮件
		response.SetupComplete = false

	default:
		return nil, fmt.Errorf("不支持的MFA类型: %s", request.MFAType)
	}

	// 生成备用代码
	if request.BackupCodes {
		response.BackupCodes = ss.generateBackupCodes()
	}

	return response, nil
}

// GetSecurityAuditLogs 获取安全审计日志
func (ss *SecurityService) GetSecurityAuditLogs(ctx context.Context, request *SecurityAuditRequest) (*SecurityAuditResponse, error) {
	// 简化实现：返回模拟数据
	logs := []core.AuditLog{
		{
			ID:            "log_001",
			Timestamp:     time.Now().Add(-1 * time.Hour),
			UserAddress:   request.UserAddress,
			Action:        "login",
			Resource:      "auth",
			Result:        "success",
			IPAddress:     "192.168.1.100",
			RiskScore:     0.2,
			SecurityLevel: "medium",
		},
		{
			ID:            "log_002",
			Timestamp:     time.Now().Add(-2 * time.Hour),
			UserAddress:   request.UserAddress,
			Action:        "create_multisig_transaction",
			Resource:      "multisig_wallet",
			Result:        "success",
			IPAddress:     "192.168.1.100",
			RiskScore:     0.5,
			SecurityLevel: "high",
		},
	}

	summary := &AuditSummary{
		TotalActions:      len(logs),
		SuccessfulActions: len(logs),
		FailedActions:     0,
		HighRiskActions:   1,
		UniqueIPs:         1,
		TimeRange:         "last_24h",
	}

	riskAnalysis := &RiskAnalysis{
		OverallRisk:       "low",
		RiskFactors:       []string{"正常使用模式", "可信IP地址"},
		Recommendations:   []string{"继续保持良好的安全习惯", "定期更新密码"},
		AnomalousPatterns: []string{},
		ThreatIndicators:  []string{},
	}

	response := &SecurityAuditResponse{
		Logs:         logs,
		Summary:      summary,
		RiskAnalysis: riskAnalysis,
		Total:        len(logs),
		HasMore:      false,
	}

	return response, nil
}

// 私有方法

// generateTransactionID 生成交易ID
func (ss *SecurityService) generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}

// verifyMFACode 验证MFA代码
func (ss *SecurityService) verifyMFACode(userAddress, code string) bool {
	// 简化实现：总是返回true
	return true
}

// generateTOTPSecret 生成TOTP密钥
func (ss *SecurityService) generateTOTPSecret() string {
	return "JBSWY3DPEHPK3PXP" // 示例密钥
}

// generateQRCode 生成二维码
func (ss *SecurityService) generateQRCode(userAddress, secret string) string {
	// 简化实现：返回示例二维码URL
	return fmt.Sprintf("otpauth://totp/Wallet:%s?secret=%s&issuer=Wallet", userAddress, secret)
}

// generateBackupCodes 生成备用代码
func (ss *SecurityService) generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		codes[i] = fmt.Sprintf("%08d", time.Now().UnixNano()%(100000000))
	}
	return codes
}
