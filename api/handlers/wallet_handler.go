package handlers

import (
	"net/http"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// WalletHandler 处理钱包相关的HTTP请求
type WalletHandler struct {
	walletService *services.WalletService
}

// NewWalletHandler 创建一个新的钱包处理器实例
func NewWalletHandler(walletService *services.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// CreateWalletRequest 创建钱包的请求结构体
type CreateWalletRequest struct {
	Name     string `json:"name" binding:"required"`     // 钱包名称
	Password string `json:"password" binding:"required"` // 钱包密码
}

// CreateWalletResponse 创建钱包的响应结构体
type CreateWalletResponse struct {
	Address string `json:"address"` // 钱包地址
	Name    string `json:"name"`    // 钱包名称
}

// CreateWallet 处理创建钱包的HTTP请求
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	// 调用服务层创建钱包
	wallet, err := h.walletService.CreateWallet(req.Name, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorWalletCreate,
			"msg":  e.GetMsg(e.ErrorWalletCreate),
			"data": err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": CreateWalletResponse{
			Address: wallet.Address,
			Name:    wallet.Name,
		},
	})
}

// GetWalletInfo 获取钱包信息
func (h *WalletHandler) GetWalletInfo(c *gin.Context) {
	address := c.Param("address")

	// 这里可以添加地址格式验证
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "钱包地址不能为空",
		})
		return
	}

	// TODO: 实现获取钱包信息的逻辑
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"address": address,
			"message": "获取钱包信息功能待实现",
		},
	})
}