/*
助记词认证API处理器包

本包实现了基于助记词的认证机制，用户可以通过助记词直接访问钱包功能，
无需传统的用户名/密码注册登录流程。

主要功能：
助记词认证：
- 通过助记词生成钱包地址
- 验证助记词有效性
- 生成临时会话用于交易操作

会话管理：
- 临时会话创建和销毁
- 会话超时管理
- 会话安全存储

安全特性：
- 助记词不持久化存储
- 会话临时存储在内存中
- 自动会话清理机制
*/

package handlers

import (
	"net/http"
	"time"
	"wallet/api/middleware"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// MnemonicAuthHandler 助记词认证处理器
type MnemonicAuthHandler struct {
	walletService *services.WalletService
}

// NewMnemonicAuthHandler 创建新的助记词认证处理器实例
func NewMnemonicAuthHandler(walletService *services.WalletService) *MnemonicAuthHandler {
	return &MnemonicAuthHandler{
		walletService: walletService,
	}
}

// =============================================================================
// 请求和响应结构体定义
// =============================================================================

// MnemonicAuthRequest 助记词认证请求
type MnemonicAuthRequest struct {
	Mnemonic       string `json:"mnemonic" binding:"required"`
	DerivationPath string `json:"derivation_path"` // 可选，BIP44派生路径，默认为 m/44'/60'/0'/0/0
	Name           string `json:"name"`            // 可选，钱包显示名称
}

// MnemonicAuthResponse 助记词认证响应
type MnemonicAuthResponse struct {
	SessionID      string `json:"session_id"`
	Address        string `json:"address"`
	DerivationPath string `json:"derivation_path"`
	ExpiresAt      int64  `json:"expires_at"`
}

// CreateWalletResponse 创建钱包响应
type CreateWalletResponse struct {
	Mnemonic string `json:"mnemonic"`
	Address  string `json:"address"`
}

// =============================================================================
// 助记词认证和钱包创建
// =============================================================================

// AuthenticateWithMnemonic
// * 通过助记词认证并创建临时会话
// * 生成会话ID用于后续的交易操作
func (h *MnemonicAuthHandler) AuthenticateWithMnemonic(c *gin.Context) {
	var req MnemonicAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 设置默认派生路径
	derivationPath := req.DerivationPath
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}

	// 通过助记词派生地址
	address, err := h.walletService.ImportMnemonic(req.Mnemonic, derivationPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorWalletImport,
			"msg":  "助记词无效或派生地址失败",
			"data": nil,
		})
		return
	}

	// 创建临时会话（1小时有效期）
	sessionID, err := h.walletService.CreateSession(req.Mnemonic, derivationPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建会话失败",
			"data": nil,
		})
		return
	}

	// 生成JWT token（用于API认证）
	authManager := middleware.GetAuthManager()
	if authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "认证服务未初始化",
			"data": nil,
		})
		return
	}

	// Token有效期1小时
	tokenExpiry := 1 * time.Hour

	token, err := authManager.GenerateJWT(sessionID, "mnemonic_user", tokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "生成Token失败",
			"data": nil,
		})
		return
	}

	// 返回认证响应
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "认证成功",
		"data": MnemonicAuthResponse{
			SessionID:      sessionID,
			Address:        address,
			DerivationPath: derivationPath,
			ExpiresAt:      time.Now().Add(tokenExpiry).Unix(),
		},
		"token": token,
	})
}

// CreateWallet
// * 创建新的助记词钱包
// * 返回助记词和派生地址
func (h *MnemonicAuthHandler) CreateWallet(c *gin.Context) {
	// 生成新的助记词和地址
	mnemonic, address, err := h.walletService.CreateNewWallet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorWalletCreate,
			"msg":  "创建钱包失败",
			"data": nil,
		})
		return
	}

	// 返回创建响应
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "钱包创建成功，请妥善保管助记词",
		"data": CreateWalletResponse{
			Mnemonic: mnemonic,
			Address:  address,
		},
	})
}

// Logout
// * 注销会话
// * 清理会话数据
func (h *MnemonicAuthHandler) Logout(c *gin.Context) {
	// 从上下文获取会话ID
	sessionID, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "未找到有效会话",
			"data": nil,
		})
		return
	}

	// 清理会话
	h.walletService.ClearSession(sessionID.(string))

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "会话已注销",
		"data": nil,
	})
}
