package handlers

import (
	"math/big"
	"net/http"
	"strings"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NetworkHandler 网络管理处理器
type NetworkHandler struct {
	walletService *services.WalletService
}

// NewNetworkHandler 创建网络处理器
func NewNetworkHandler(walletService *services.WalletService) *NetworkHandler {
	return &NetworkHandler{
		walletService: walletService,
	}
}

// GetCurrentNetwork 获取当前网络
func (h *NetworkHandler) GetCurrentNetwork(c *gin.Context) {
	currentNetwork := h.walletService.GetCurrentNetwork()

	// 获取网络详细信息
	networkInfo, err := h.walletService.GetNetworkInfo(currentNetwork)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取网络信息失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": networkInfo,
	})
}

// SwitchNetworkRequest 切换网络请求
type SwitchNetworkRequest struct {
	NetworkID string `json:"network_id" binding:"required"`
}

// SwitchNetwork 切换网络
func (h *NetworkHandler) SwitchNetwork(c *gin.Context) {
	var req SwitchNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	err := h.walletService.SwitchNetwork(req.NetworkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "切换网络失败",
			"data": err.Error(),
		})
		return
	}

	// 获取切换后的网络信息
	networkInfo, err := h.walletService.GetNetworkInfo(req.NetworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取网络信息失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": networkInfo,
	})
}

// ListNetworks 列出所有可用网络
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	networksInfo, err := h.walletService.GetAllNetworksInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取网络列表失败",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"networks":        networksInfo,
			"current_network": h.walletService.GetCurrentNetwork(),
		},
	})
}

// GetNetworkInfo 获取指定网络信息
func (h *NetworkHandler) GetNetworkInfo(c *gin.Context) {
	networkID := c.Param("networkId")
	if networkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "网络ID不能为空",
		})
		return
	}

	networkInfo, err := h.walletService.GetNetworkInfo(networkID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.InvalidParams,
			"msg":  "网络不存在",
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": networkInfo,
	})
}

// GetCrossChainBalance 获取跨链余额
func (h *NetworkHandler) GetCrossChainBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "地址不能为空",
		})
		return
	}

	// 获取网络列表（从查询参数或使用所有可用网络）
	networksParam := c.Query("networks")
	var networks []string
	if networksParam != "" {
		networks = strings.Split(networksParam, ",")
	} else {
		networks = h.walletService.GetAvailableNetworks()
	}

	balances, err := h.walletService.GetCrossChainBalance(address, networks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorGetBalance,
			"msg":  e.GetMsg(e.ErrorGetBalance),
			"data": err.Error(),
		})
		return
	}

	// 转换为字符串格式
	balanceStrings := make(map[string]string)
	for network, balance := range balances {
		balanceStrings[network] = balance.String()
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"address":  address,
			"balances": balanceStrings,
		},
	})
}

// GetCrossChainTokenBalance 获取跨链代币余额
func (h *NetworkHandler) GetCrossChainTokenBalance(c *gin.Context) {
	address := c.Param("address")
	tokenAddress := c.Param("tokenAddress")

	if address == "" || tokenAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "地址和代币地址不能为空",
		})
		return
	}

	// 获取网络列表
	networksParam := c.Query("networks")
	var networks []string
	if networksParam != "" {
		networks = strings.Split(networksParam, ",")
	} else {
		networks = h.walletService.GetAvailableNetworks()
	}

	balances, err := h.walletService.GetCrossChainTokenBalance(address, tokenAddress, networks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorGetBalance,
			"msg":  e.GetMsg(e.ErrorGetBalance),
			"data": err.Error(),
		})
		return
	}

	// 转换为字符串格式
	balanceStrings := make(map[string]string)
	for network, balance := range balances {
		balanceStrings[network] = balance.String()
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"address":       address,
			"token_address": tokenAddress,
			"balances":      balanceStrings,
		},
	})
}

// GetBalanceOnNetwork 获取指定网络上的余额
func (h *NetworkHandler) GetBalanceOnNetwork(c *gin.Context) {
	networkID := c.Param("networkId")
	address := c.Param("address")

	if networkID == "" || address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "网络ID和地址不能为空",
		})
		return
	}

	balance, err := h.walletService.GetBalanceOnNetwork(address, networkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorGetBalance,
			"msg":  e.GetMsg(e.ErrorGetBalance),
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"network_id":  networkID,
			"address":     address,
			"balance_wei": balance.String(),
		},
	})
}

// SendETHOnNetworkRequest 在指定网络发送ETH请求
type SendETHOnNetworkRequest struct {
	NetworkID      string `json:"network_id" binding:"required"`
	SessionID      string `json:"session_id"`
	Mnemonic       string `json:"mnemonic"`
	DerivationPath string `json:"derivation_path"`
	To             string `json:"to" binding:"required"`
	ValueWei       string `json:"value_wei" binding:"required"`
}

// SendETHOnNetwork 在指定网络发送ETH
func (h *NetworkHandler) SendETHOnNetwork(c *gin.Context) {
	var req SendETHOnNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	// 解析金额
	val := new(big.Int)
	if _, ok := val.SetString(req.ValueWei, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "value_wei 需要是十进制数字字符串",
		})
		return
	}

	var (
		txHash string
		err    error
	)

	// 获取助记词
	var mnemonic string
	if req.SessionID != "" {
		mnemonic, err = h.walletService.GetSessionMnemonic(req.SessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.InvalidParams,
				"msg":  "Session无效",
				"data": err.Error(),
			})
			return
		}
	} else if req.Mnemonic != "" {
		mnemonic = req.Mnemonic
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "需要提供 session_id 或 mnemonic",
		})
		return
	}

	// 发送交易
	txHash, err = h.walletService.SendETHOnNetwork(req.NetworkID, mnemonic, req.DerivationPath, req.To, val)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorTransactionSend,
			"msg":  e.GetMsg(e.ErrorTransactionSend),
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"tx_hash":    txHash,
			"network_id": req.NetworkID,
		},
	})
}
