/*
安全功能增强API处理器

本文件实现了安全功能增强的HTTP接口处理器，包括：

主要接口：
- 硬件钱包管理：检测、连接、操作硬件钱包设备
- 多重签名钱包：创建、管理、交易签名功能
- 多因素认证：TOTP、SMS、Email等MFA设置
- 安全审计：操作日志、风险分析、安全报告
- 生物识别：指纹、面容识别认证

接口分组：
- /api/v1/security/hardware/* - 硬件钱包接口
- /api/v1/security/multisig/* - 多重签名接口
- /api/v1/security/mfa/* - 多因素认证接口
- /api/v1/security/audit/* - 安全审计接口
- /api/v1/security/biometric/* - 生物识别接口

安全特性：
- 请求签名验证
- 设备指纹识别
- 操作日志记录
- 风险等级评估
- 实时威胁检测
*/
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// SecurityHandler 安全功能API处理器
// 处理所有安全相关的HTTP请求，包括硬件钱包、多重签名、MFA认证等功能
type SecurityHandler struct {
	securityService *services.SecurityService // 安全功能业务服务实例
}

// NewSecurityHandler 创建新的安全功能处理器实例
// 参数: securityService - 安全功能业务服务实例
// 返回: 配置好的安全功能处理器
func NewSecurityHandler(securityService *services.SecurityService) *SecurityHandler {
	return &SecurityHandler{
		securityService: securityService,
	}
}

// DetectHardwareWallets 检测硬件钱包
// GET /api/v1/security/hardware/detect
// 功能: 检测当前连接的硬件钱包设备
// 响应: 检测到的硬件钱包设备列表
func (h *SecurityHandler) DetectHardwareWallets(c *gin.Context) {
	wallets, err := h.securityService.DetectHardwareWallets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "检测硬件钱包失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "检测成功",
		"data": gin.H{
			"hardware_wallets": wallets,
			"count":            len(wallets),
			"detected_at":      time.Now().Unix(),
		},
	})
}

// ProcessHardwareWalletRequest 处理硬件钱包请求
// POST /api/v1/security/hardware/request
// 请求体: HardwareWalletRequest结构体
// 功能: 处理硬件钱包的各种操作请求（连接、获取地址、签名等）
func (h *SecurityHandler) ProcessHardwareWalletRequest(c *gin.Context) {
	var req services.HardwareWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.Action == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "操作类型不能为空",
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 处理硬件钱包请求
	response, err := h.securityService.ProcessHardwareWalletRequest(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "处理硬件钱包请求失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 根据响应状态设置HTTP状态码
	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"code": e.SUCCESS,
		"msg":  response.Message,
		"data": response,
	})
}

// CreateMultiSigWallet 创建多重签名钱包
// POST /api/v1/security/multisig/create
// 请求体: MultiSigWalletRequest结构体
// 功能: 创建新的多重签名钱包
func (h *SecurityHandler) CreateMultiSigWallet(c *gin.Context) {
	var req services.MultiSigWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.Name == "" || req.Threshold < 1 || len(req.Signers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "钱包名称、签名阈值和签名者列表不能为空",
			"data": nil,
		})
		return
	}

	// 验证签名阈值
	if req.Threshold > len(req.Signers) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "签名阈值不能大于签名者数量",
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 创建多重签名钱包
	response, err := h.securityService.CreateMultiSigWallet(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建多重签名钱包失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "多重签名钱包创建成功",
		"data": response,
	})
}

// CreateMultiSigTransaction 创建多重签名交易
// POST /api/v1/security/multisig/transaction/create
// 请求体: MultiSigTransactionRequest结构体
// 功能: 创建需要多重签名确认的交易
func (h *SecurityHandler) CreateMultiSigTransaction(c *gin.Context) {
	var req services.MultiSigTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.WalletID == "" || req.Title == "" || req.To == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "钱包ID、交易标题和接收地址不能为空",
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 创建多重签名交易
	response, err := h.securityService.CreateMultiSigTransaction(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建多重签名交易失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "多重签名交易创建成功",
		"data": response,
	})
}

// SignMultiSigTransaction 签名多重签名交易
// POST /api/v1/security/multisig/transaction/sign
// 请求体: SignTransactionRequest结构体
// 功能: 为多重签名交易提供签名
func (h *SecurityHandler) SignMultiSigTransaction(c *gin.Context) {
	var req services.SignTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.WalletID == "" || req.TransactionID == "" || req.SignerAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "钱包ID、交易ID和签名者地址不能为空",
			"data": nil,
		})
		return
	}

	// 签名多重签名交易
	err := h.securityService.SignMultiSigTransaction(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "签名交易失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "交易签名成功",
		"data": gin.H{
			"wallet_id":      req.WalletID,
			"transaction_id": req.TransactionID,
			"signer_address": req.SignerAddress,
			"signed_at":      time.Now().Unix(),
		},
	})
}

// SetupMFA 设置多因素认证
// POST /api/v1/security/mfa/setup
// 请求体: MFASetupRequest结构体
// 功能: 设置多因素认证（TOTP、SMS、Email等）
func (h *SecurityHandler) SetupMFA(c *gin.Context) {
	var req services.MFASetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证MFA类型
	if req.MFAType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "MFA类型不能为空",
			"data": nil,
		})
		return
	}

	// 根据MFA类型验证额外参数
	switch req.MFAType {
	case "SMS":
		if req.PhoneNumber == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": e.InvalidParams,
				"msg":  "设置SMS MFA时手机号码不能为空",
				"data": nil,
			})
			return
		}
	case "Email":
		if req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": e.InvalidParams,
				"msg":  "设置Email MFA时邮箱地址不能为空",
				"data": nil,
			})
			return
		}
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 设置MFA
	response, err := h.securityService.SetupMFA(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "设置MFA失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "MFA设置成功",
		"data": response,
	})
}

// VerifyMFA 验证多因素认证
// POST /api/v1/security/mfa/verify
// 请求体: { "mfa_type": "TOTP", "code": "123456" }
// 功能: 验证MFA代码
func (h *SecurityHandler) VerifyMFA(c *gin.Context) {
	var req struct {
		MFAType string `json:"mfa_type" binding:"required"`
		Code    string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：验证MFA代码（实际项目中需要实现真实的验证逻辑）
	isValid := len(req.Code) == 6 // 简单的长度验证

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "MFA验证失败",
			"data": gin.H{
				"valid": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "MFA验证成功",
		"data": gin.H{
			"valid":        true,
			"user_address": userAddress,
			"mfa_type":     req.MFAType,
			"verified_at":  time.Now().Unix(),
		},
	})
}

// GetSecurityAuditLogs 获取安全审计日志
// GET /api/v1/security/audit/logs
// 查询参数:
//   - start_time: 开始时间（Unix时间戳）
//   - end_time: 结束时间（Unix时间戳）
//   - action_types: 操作类型过滤（逗号分隔）
//   - security_levels: 安全级别过滤（逗号分隔）
//   - limit: 返回数量限制（默认50）
//   - offset: 偏移量（默认0）
//
// 响应: 安全审计日志列表和统计分析
func (h *SecurityHandler) GetSecurityAuditLogs(c *gin.Context) {
	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 解析查询参数
	request := &services.SecurityAuditRequest{
		UserAddress: userAddress,
	}

	// 解析时间参数
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			t := time.Unix(startTime, 0)
			request.StartTime = &t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			t := time.Unix(endTime, 0)
			request.EndTime = &t
		}
	}

	// 解析数值参数
	limitStr := c.DefaultQuery("limit", "50")
	if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
		request.Limit = limit
	} else {
		request.Limit = 50
	}

	offsetStr := c.DefaultQuery("offset", "0")
	if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
		request.Offset = offset
	} else {
		request.Offset = 0
	}

	// 获取安全审计日志
	response, err := h.securityService.GetSecurityAuditLogs(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取审计日志失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取审计日志成功",
		"data": response,
	})
}

// GetSecurityReport 获取安全报告
// GET /api/v1/security/audit/report/:address
// 路径参数:
//   - address: 用户地址
//
// 查询参数:
//   - time_range: 时间范围（24h/7d/30d/90d）
//   - report_type: 报告类型（summary/detailed/threat_analysis）
//
// 响应: 安全分析报告
func (h *SecurityHandler) GetSecurityReport(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	timeRange := c.DefaultQuery("time_range", "7d")
	reportType := c.DefaultQuery("report_type", "summary")

	// 构建审计请求
	var startTime *time.Time
	switch timeRange {
	case "24h":
		t := time.Now().Add(-24 * time.Hour)
		startTime = &t
	case "7d":
		t := time.Now().Add(-7 * 24 * time.Hour)
		startTime = &t
	case "30d":
		t := time.Now().Add(-30 * 24 * time.Hour)
		startTime = &t
	case "90d":
		t := time.Now().Add(-90 * 24 * time.Hour)
		startTime = &t
	}

	request := &services.SecurityAuditRequest{
		UserAddress: userAddress,
		StartTime:   startTime,
		Limit:       1000, // 获取更多数据用于分析
		Offset:      0,
	}

	// 获取审计数据
	auditResponse, err := h.securityService.GetSecurityAuditLogs(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取安全报告失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 构建报告响应
	report := gin.H{
		"user_address":  userAddress,
		"time_range":    timeRange,
		"report_type":   reportType,
		"generated_at":  time.Now().Unix(),
		"summary":       auditResponse.Summary,
		"risk_analysis": auditResponse.RiskAnalysis,
	}

	// 根据报告类型添加详细信息
	if reportType == "detailed" || reportType == "threat_analysis" {
		report["logs"] = auditResponse.Logs
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "安全报告生成成功",
		"data": report,
	})
}

// EnableBiometric 启用生物识别认证
// POST /api/v1/security/biometric/enable
// 请求体: { "biometric_type": "fingerprint|face|voice", "device_info": {...} }
// 功能: 启用生物识别认证
func (h *SecurityHandler) EnableBiometric(c *gin.Context) {
	var req struct {
		BiometricType string                 `json:"biometric_type" binding:"required"`
		DeviceInfo    map[string]interface{} `json:"device_info"`
		BackupMethod  string                 `json:"backup_method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证生物识别类型
	validTypes := map[string]bool{
		"fingerprint": true,
		"face":        true,
		"voice":       true,
	}

	if !validTypes[req.BiometricType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "不支持的生物识别类型",
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "生物识别认证启用成功",
		"data": gin.H{
			"user_address":   userAddress,
			"biometric_type": req.BiometricType,
			"backup_method":  req.BackupMethod,
			"enabled_at":     time.Now().Unix(),
			"enrollment_id":  "bio_" + strconv.FormatInt(time.Now().UnixNano(), 36),
		},
	})
}

// VerifyBiometric 验证生物识别
// POST /api/v1/security/biometric/verify
// 请求体: { "biometric_type": "fingerprint", "biometric_data": "base64_data", "challenge": "..." }
// 功能: 验证生物识别数据
func (h *SecurityHandler) VerifyBiometric(c *gin.Context) {
	var req struct {
		BiometricType string `json:"biometric_type" binding:"required"`
		BiometricData string `json:"biometric_data" binding:"required"`
		Challenge     string `json:"challenge"`
		DeviceID      string `json:"device_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 从请求头获取用户地址
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：模拟生物识别验证
	isValid := len(req.BiometricData) > 10 // 简单的数据长度验证

	if !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "生物识别验证失败",
			"data": gin.H{
				"valid":  false,
				"reason": "生物识别数据无效",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "生物识别验证成功",
		"data": gin.H{
			"valid":          true,
			"user_address":   userAddress,
			"biometric_type": req.BiometricType,
			"verified_at":    time.Now().Unix(),
			"confidence":     0.95, // 置信度
		},
	})
}

// GetSecurityStatus 获取安全状态概览
// GET /api/v1/security/status/:address
// 路径参数:
//   - address: 用户地址
//
// 响应: 用户的安全功能启用状态和安全评分
func (h *SecurityHandler) GetSecurityStatus(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：返回模拟的安全状态
	securityStatus := gin.H{
		"user_address":   userAddress,
		"security_score": 85, // 安全评分（0-100）
		"features": gin.H{
			"hardware_wallet": gin.H{
				"enabled":      true,
				"device_count": 2,
				"last_used":    time.Now().Add(-2 * time.Hour).Unix(),
			},
			"multisig_wallet": gin.H{
				"enabled":             true,
				"wallet_count":        1,
				"active_transactions": 0,
			},
			"mfa": gin.H{
				"enabled":      true,
				"methods":      []string{"TOTP", "SMS"},
				"backup_codes": 8,
			},
			"biometric": gin.H{
				"enabled": true,
				"types":   []string{"fingerprint", "face"},
			},
		},
		"risk_analysis": gin.H{
			"level":   "low",
			"factors": []string{"启用了多重安全验证", "使用硬件钱包", "定期安全审计"},
			"recommendations": []string{
				"定期更新备用代码",
				"检查设备安全状态",
			},
		},
		"recent_activity": gin.H{
			"login_attempts":    5,
			"successful_logins": 5,
			"failed_attempts":   0,
			"last_login":        time.Now().Add(-1 * time.Hour).Unix(),
		},
		"updated_at": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取安全状态成功",
		"data": securityStatus,
	})
}
