package handlers

import (
	"math/big"
	"net/http"
	"strings"
	"wallet/core"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NetworkHandler 网络相关的HTTP请求处理器
type NetworkHandler struct {
	multiChain    *core.MultiChainManager
	walletService *services.WalletService // 修复类型错误
}

// NewNetworkHandler 创建新的网络处理器实例
func NewNetworkHandler(multiChain *core.MultiChainManager, walletService *services.WalletService) *NetworkHandler {
	return &NetworkHandler{
		multiChain:    multiChain,
		walletService: walletService,
	}
}

// NetworkInfoResponse 网络信息响应
type NetworkInfoResponse struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	ChainID       int64               `json:"chain_id"`
	Symbol        string              `json:"symbol"`
	Decimals      int                 `json:"decimals"`
	BlockExplorer string              `json:"block_explorer"`
	Testnet       bool                `json:"testnet"`
	LatestBlock   uint64              `json:"latest_block"`
	GasSuggestion *core.GasSuggestion `json:"gas_suggestion"`
	Connected     bool                `json:"connected"`
	ChainType     string              `json:"chain_type"`
}

// ListNetworks 获取网络列表
func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	networks := h.multiChain.GetAvailableNetworks()

	response := make([]NetworkInfoResponse, len(networks))
	for i, network := range networks {
		response[i] = NetworkInfoResponse{
			ID:            network.ID,
			Name:          network.Name,
			ChainID:       network.ChainID,
			Symbol:        network.Symbol,
			Decimals:      network.Decimals,
			BlockExplorer: network.BlockExplorer,
			Testnet:       network.Testnet,
			LatestBlock:   network.LatestBlock,
			GasSuggestion: network.GasSuggestion,
			Connected:     network.Connected,
			ChainType:     network.ChainType,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

// GetNetworkInfo 获取特定网络信息
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

	networkInfo, err := h.multiChain.GetNetworkInfo(networkID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	response := NetworkInfoResponse{
		ID:            networkInfo.ID,
		Name:          networkInfo.Name,
		ChainID:       networkInfo.ChainID,
		Symbol:        networkInfo.Symbol,
		Decimals:      networkInfo.Decimals,
		BlockExplorer: networkInfo.BlockExplorer,
		Testnet:       networkInfo.Testnet,
		LatestBlock:   networkInfo.LatestBlock,
		GasSuggestion: networkInfo.GasSuggestion,
		Connected:     networkInfo.Connected,
		ChainType:     networkInfo.ChainType,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
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
		// 获取所有可用网络的ID
		networkInfos := h.multiChain.GetAvailableNetworks()
		for _, networkInfo := range networkInfos {
			networks = append(networks, networkInfo.ID)
		}
	}

	balances, err := h.multiChain.GetCrossChainBalance(address, networks)
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
		// 获取所有可用网络的ID
		networkInfos := h.multiChain.GetAvailableNetworks()
		for _, networkInfo := range networkInfos {
			networks = append(networks, networkInfo.ID)
		}
	}

	// 使用GetCrossChainBalance方法替代未实现的方法
	balances, err := h.multiChain.GetCrossChainBalance(address, networks)
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
	address := c.Param("address")
	networkID := c.Query("network")

	if address == "" || networkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "地址和网络ID不能为空",
		})
		return
	}

	// 使用钱包服务的方法
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

	// 使用钱包服务的方法
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
			"tx_hash": txHash,
		},
	})
}

// SwitchNetworkRequest 切换网络请求
type SwitchNetworkRequest struct {
	NetworkID string `json:"network_id" binding:"required"`
}

// SwitchNetwork 切换网络
// POST /api/v1/networks/switch
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

	if err := h.multiChain.SwitchNetwork(req.NetworkID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	// 获取切换后的网络信息
	networkInfo, err := h.multiChain.GetNetworkInfo(req.NetworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	response := NetworkInfoResponse{
		ID:            networkInfo.ID,
		Name:          networkInfo.Name,
		ChainID:       networkInfo.ChainID,
		Symbol:        networkInfo.Symbol,
		Decimals:      networkInfo.Decimals,
		BlockExplorer: networkInfo.BlockExplorer,
		Testnet:       networkInfo.Testnet,
		LatestBlock:   networkInfo.LatestBlock,
		GasSuggestion: networkInfo.GasSuggestion,
		Connected:     networkInfo.Connected,
		ChainType:     networkInfo.ChainType,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

// GetCurrentNetwork 获取当前网络信息
// GET /api/v1/networks/current
func (h *NetworkHandler) GetCurrentNetwork(c *gin.Context) {
	currentNetworkID := h.multiChain.GetCurrentNetwork()

	networkInfo, err := h.multiChain.GetNetworkInfo(currentNetworkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  err.Error(),
			"data": nil,
		})
		return
	}

	response := NetworkInfoResponse{
		ID:            networkInfo.ID,
		Name:          networkInfo.Name,
		ChainID:       networkInfo.ChainID,
		Symbol:        networkInfo.Symbol,
		Decimals:      networkInfo.Decimals,
		BlockExplorer: networkInfo.BlockExplorer,
		Testnet:       networkInfo.Testnet,
		LatestBlock:   networkInfo.LatestBlock,
		GasSuggestion: networkInfo.GasSuggestion,
		Connected:     networkInfo.Connected,
		ChainType:     networkInfo.ChainType,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
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
