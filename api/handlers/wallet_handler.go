/*
钱包API处理器包

本包实现了所有钱包相关的HTTP API处理器，负责处理客户端请求并调用业务服务层。

主要功能模块：
钱包管理：
- 钱包创建和导入（支持助记词和Keystore）
- 地址生成和管理（HD钱包派生）
- 钱包信息查询和更新
- 会话管理和助记词临时存储

余额查询：
- 原生代币余额查询（ETH/MATIC/BNB等）
- ERC20代币余额查询
- 跨链余额聚合查询
- 批量地址余额查询

交易处理：
- 原生代币转账（支持Legacy和EIP-1559）
- ERC20代币转账
- 交易费估算和优化
- 原始交易广播
- 交易历史查询和过滤

高级功能：
- 智能合约交互
- 消息签名（Personal Sign和EIP-712）
- 代币授权管理
- Gas价格建议

安全特性：
- 请求参数验证和清理
- 错误信息脱敏处理
- 会话超时管理
- 交易权限检查
*/
package handlers

import (
	"math/big"
	"net/http"
	"wallet/core"
	"wallet/pkg/e"
	"wallet/services"

	// 需要导入 strings 包
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// WalletHandler 钱包相关的HTTP请求处理器
// 封装了所有钱包操作的API接口，提供RESTful风格的服务
type WalletHandler struct {
	walletService *services.WalletService // 钱包业务服务实例，用于处理具体的业务逻辑
}

// NewWalletHandler 创建新的钱包处理器实例
// 参数: walletService - 钱包业务服务实例
// 返回: 初始化完成的WalletHandler指针
func NewWalletHandler(walletService *services.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// ImportMnemonicRequest 导入助记词的请求参数
// 支持通过BIP39助记词导入已存在的钱包
type ImportMnemonicRequest struct {
	Name           string `json:"name"`                        // 钱包显示名称（可选，MVP版本不入库）
	Mnemonic       string `json:"mnemonic" binding:"required"` // BIP39助记词（必填，12-24个单词）
	DerivationPath string `json:"derivation_path" binding:"-"` // BIP44派生路径（默认: m/44'/60'/0'/0/0）
}

// ImportMnemonic 导入助记词并返回派生地址
// POST /api/v1/wallets/import-mnemonic
// 功能: 通过BIP39助记词导入钱包并生成以太坊地址
// 参数: ImportMnemonicRequest JSON请求体
// 返回: 包含生成地址的JSON响应
// 注意: MVP版本不会持久化存储助记词，仅用于验证和地址生成
func (h *WalletHandler) ImportMnemonic(c *gin.Context) {
	var req ImportMnemonicRequest
	// 绑定JSON请求参数并验证必填字段
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	// 调用业务服务层导入助记词并生成地址
	addr, err := h.walletService.ImportMnemonic(req.Mnemonic, req.DerivationPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorWalletImport,
			"msg":  e.GetMsg(e.ErrorWalletImport),
			"data": err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{"address": addr},
	})
}

// GetBalance 查询指定地址的原生代币余额
// GET /api/v1/wallets/:address/balance
// 功能: 获取以太坊地址的ETH余额（或其他网络的原生代币）
// 参数: address - 路径参数，以太坊地址（0x开头）
// 返回: 包含余额信息的JSON响应（wei单位）
func (h *WalletHandler) GetBalance(c *gin.Context) {
	// 获取路径参数中的地址
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "钱包地址不能为空",
		})
		return
	}

	// 调用业务服务层获取余额
	bal, err := h.walletService.GetBalance(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorGetBalance,
			"msg":  e.GetMsg(e.ErrorGetBalance),
			"data": err.Error(),
		})
		return
	}

	// 返回成功响应，余额以wei为单位
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{"address": address, "balance_wei": bal.String()},
	})
}

// SendTransactionRequest 发送原生代币交易的请求参数
// 支持会话模式和直接助记词模式两种认证方式
type SendTransactionRequest struct {
	SessionID      string `json:"session_id"`                   // 会话 ID（与 mnemonic 二选一）
	Mnemonic       string `json:"mnemonic"`                     // BIP39助记词（与 session_id 二选一）
	DerivationPath string `json:"derivation_path"`              // BIP44派生路径（默认: m/44'/60'/0'/0/0）
	To             string `json:"to" binding:"required"`        // 接收方地址（必填）
	ValueWei       string `json:"value_wei" binding:"required"` // 转账金额（wei单位的十进制字符串）
}

// SendTransaction 发送 ETH 交易
func (h *WalletHandler) SendTransaction(c *gin.Context) {
	var req SendTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

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
	if req.SessionID != "" {
		txHash, err = h.walletService.SendETHWithSession(req.SessionID, req.DerivationPath, req.To, val)
	} else if req.Mnemonic != "" {
		txHash, err = h.walletService.SendETH(req.Mnemonic, req.DerivationPath, req.To, val)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}

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
		"data": gin.H{"tx_hash": txHash},
	})
}

// GetERC20Balance 查询指定地址的某个 ERC20 余额
func (h *WalletHandler) GetERC20Balance(c *gin.Context) {
	address := c.Param("address")
	token := c.Param("tokenAddress")
	if address == "" || token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "address 或 tokenAddress 不能为空",
		})
		return
	}
	bal, err := h.walletService.GetERC20Balance(address, token)
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
		"data": gin.H{"address": address, "token": token, "balance": bal.String()},
	})
}

// SendERC20Request 发送 ERC20 请求
type SendERC20Request struct {
	SessionID      string `json:"session_id"`      // 新增
	Mnemonic       string `json:"mnemonic"`        // 可选（与 session 二选一）
	DerivationPath string `json:"derivation_path"` // 默认为 m/44'/60'/0'/0/0
	Token          string `json:"token" binding:"required"`
	To             string `json:"to" binding:"required"`
	Amount         string `json:"amount" binding:"required"` // token 最小单位，十进制字符串
}

// SendERC20 发送 ERC20 转账
func (h *WalletHandler) SendERC20(c *gin.Context) {
	var req SendERC20Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}
	amount := new(big.Int)
	if _, ok := amount.SetString(req.Amount, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "amount 需要是十进制数字字符串",
		})
		return
	}

	var (
		txHash string
		err    error
	)
	if req.SessionID != "" {
		txHash, err = h.walletService.SendERC20WithSession(req.SessionID, req.DerivationPath, req.Token, req.To, amount)
	} else if req.Mnemonic != "" {
		txHash, err = h.walletService.SendERC20(req.Mnemonic, req.DerivationPath, req.Token, req.To, amount)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}

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
		"data": gin.H{"tx_hash": txHash},
	})
}

// GetNonces 获取地址的 latest 与 pending nonce
func (h *WalletHandler) GetNonces(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "地址不能为空"})
		return
	}
	pending, latest, err := h.walletService.GetNonces(address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorNonceGet, "msg": e.GetMsg(e.ErrorNonceGet), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{"address": address, "nonce_latest": latest, "nonce_pending": pending},
	})
}

// GetGasSuggestion 获取 gas 建议（EIP-1559 + legacy）
func (h *WalletHandler) GetGasSuggestion(c *gin.Context) {
	sug, err := h.walletService.GetGasSuggestion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorGasSuggestion, "msg": e.GetMsg(e.ErrorGasSuggestion), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"chain_id":  sug.ChainID.String(),
			"base_fee":  sug.BaseFee.String(),
			"tip_cap":   sug.TipCap.String(),
			"max_fee":   sug.MaxFee.String(),
			"gas_price": sug.GasPrice.String(),
		},
	})
}

// EstimateTransaction 估算交易 gasLimit
type EstimateTxRequest struct {
	From     string `json:"from" binding:"required"`
	To       string `json:"to"`        // 合约交互可为空（仅 data）
	ValueWei string `json:"value_wei"` // 可选，十进制字符串
	DataHex  string `json:"data"`      // 可选，0x 开头或纯 hex
}

func (h *WalletHandler) EstimateTransaction(c *gin.Context) {
	var req EstimateTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	val := big.NewInt(0)
	if req.ValueWei != "" {
		if _, ok := val.SetString(req.ValueWei, 10); !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "value_wei 需要是十进制数字字符串"})
			return
		}
	}
	// data 解析在服务层处理也可，这里直接传原始 hex 字符串由服务层解析为 bytes
	limit, err := h.walletService.EstimateGas(req.From, req.To, val, []byte(req.DataHex))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionBuild, "msg": e.GetMsg(e.ErrorTransactionBuild), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"gas_limit": limit}})
}

// BroadcastRawTransaction 广播原始交易
type BroadcastTxRequest struct {
	RawTx string `json:"raw_tx" binding:"required"` // 0x 开头或纯十六进制
}

func (h *WalletHandler) BroadcastRawTransaction(c *gin.Context) {
	var req BroadcastTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	txHash, err := h.walletService.BroadcastRawTx(req.RawTx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorBroadcastRawTx, "msg": e.GetMsg(e.ErrorBroadcastRawTx), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"tx_hash": txHash}})
}

// -------- 新增：会话与批量派生、只读钱包 --------

type CreateSessionRequest struct {
	Mnemonic   string `json:"mnemonic" binding:"required"`
	TTLSeconds int    `json:"ttl_seconds"` // 可选，默认 900，最大 86400
}

func (h *WalletHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	id, exp, err := h.walletService.CreateSession(req.Mnemonic, req.TTLSeconds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorWalletImport, "msg": e.GetMsg(e.ErrorWalletImport), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{
		"session_id": id,
		"expire_at":  exp.Unix(),
	}})
}

func (h *WalletHandler) CloseSession(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "session_id 不能为空"})
		return
	}
	if err := h.walletService.CloseSession(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": "ok"})
}

type DeriveAddressesRequest struct {
	SessionID  string `json:"session_id"`  // 可选：优先使用 session
	Mnemonic   string `json:"mnemonic"`    // 可选：未提供 session_id 时使用
	PathPrefix string `json:"path_prefix"` // 默认 "m/44'/60'/0'/0"
	Start      int    `json:"start"`       // 默认 0
	Count      int    `json:"count"`       // 默认 5
}

func (h *WalletHandler) DeriveAddresses(c *gin.Context) {
	var req DeriveAddressesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.Count <= 0 {
		req.Count = 5
	}
	var (
		addrs []string
		err   error
	)
	if req.SessionID != "" {
		addrs, err = h.walletService.DeriveAddressesBySession(req.SessionID, req.PathPrefix, req.Start, req.Count)
	} else if req.Mnemonic != "" {
		addrs, err = h.walletService.DeriveAddressesFromMnemonic(req.Mnemonic, req.PathPrefix, req.Start, req.Count)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorWalletImport, "msg": e.GetMsg(e.ErrorWalletImport), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"addresses": addrs}})
}

type WatchOnlyAddRequest struct {
	Address string `json:"address" binding:"required"`
}

func (h *WalletHandler) AddWatchOnly(c *gin.Context) {
	var req WatchOnlyAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	addr, err := h.walletService.AddWatchOnly(req.Address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"address": addr}})
}

func (h *WalletHandler) ListWatchOnly(c *gin.Context) {
	addrs := h.walletService.ListWatchOnly()
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"addresses": addrs}})
}

func (h *WalletHandler) RemoveWatchOnly(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "address 不能为空"})
		return
	}
	if err := h.walletService.RemoveWatchOnly(address); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": "ok"})
}

// -------- 新增：交易回执 / token元数据 / 签名 --------

func (h *WalletHandler) GetTxReceipt(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "hash 不能为空"})
		return
	}
	dto, err := h.walletService.GetReceipt(hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionBuild, "msg": e.GetMsg(e.ErrorTransactionBuild), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": dto})
}

func (h *WalletHandler) GetTokenMetadata(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "token 地址不能为空"})
		return
	}
	name, symbol, decimals, err := h.walletService.GetTokenMetadata(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorGetBalance, "msg": e.GetMsg(e.ErrorGetBalance), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{
		"name": name, "symbol": symbol, "decimals": decimals,
	}})
}

type SendTransactionAdvanced struct {
	SessionID      string `json:"session_id"`
	Mnemonic       string `json:"mnemonic"`
	DerivationPath string `json:"derivation_path"`
	To             string `json:"to" binding:"required"`
	ValueWei       string `json:"value_wei" binding:"required"`

	// gas & nonce（十进制字符串）
	GasPrice             string `json:"gas_price"`                // legacy
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"` // EIP-1559
	MaxFeePerGas         string `json:"max_fee_per_gas"`          // EIP-1559
	GasLimit             string `json:"gas_limit"`                // 可选
	Nonce                string `json:"nonce"`                    // 可选
}

// AdvancedERC20SendRequest 高级 ERC20 发送
type AdvancedERC20SendRequest struct {
	SessionID      string `json:"session_id"`
	Mnemonic       string `json:"mnemonic"`
	DerivationPath string `json:"derivation_path"`
	Token          string `json:"token" binding:"required"`
	To             string `json:"to" binding:"required"`
	Amount         string `json:"amount" binding:"required"`

	GasPrice             string `json:"gas_price"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	GasLimit             string `json:"gas_limit"`
	Nonce                string `json:"nonce"`
}

// ApproveRequest 授权
type ApproveRequest struct {
	SessionID      string `json:"session_id"`
	Mnemonic       string `json:"mnemonic"`
	DerivationPath string `json:"derivation_path"`
	Spender        string `json:"spender" binding:"required"`
	Amount         string `json:"amount" binding:"required"`

	GasPrice             string `json:"gas_price"`
	MaxPriorityFeePerGas string `json:"max_priority_fee_per_gas"`
	MaxFeePerGas         string `json:"max_fee_per_gas"`
	GasLimit             string `json:"gas_limit"`
	Nonce                string `json:"nonce"`
}

func (h *WalletHandler) SendTransactionAdvanced(c *gin.Context) {
	var req SendTransactionAdvanced
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	val := new(big.Int)
	if _, ok := val.SetString(req.ValueWei, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "value_wei 需要十进制字符串"})
		return
	}
	opts, err := parseTxOptions(req.GasPrice, req.MaxPriorityFeePerGas, req.MaxFeePerGas, req.GasLimit, req.Nonce)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.DerivationPath == "" {
		req.DerivationPath = "m/44'/60'/0'/0/0"
	}
	var (
		txHash string
	)
	if req.SessionID != "" {
		txHash, err = h.walletService.SendETHAdvancedWithSession(req.SessionID, req.DerivationPath, req.To, val, opts)
	} else if req.Mnemonic != "" {
		txHash, err = h.walletService.SendETHAdvanced(req.Mnemonic, req.DerivationPath, req.To, val, opts)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionSend, "msg": e.GetMsg(e.ErrorTransactionSend), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"tx_hash": txHash}})
}

func (h *WalletHandler) SendERC20Advanced(c *gin.Context) {
	var req AdvancedERC20SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	amount := new(big.Int)
	if _, ok := amount.SetString(req.Amount, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "amount 需要十进制字符串"})
		return
	}
	opts, err := parseTxOptions(req.GasPrice, req.MaxPriorityFeePerGas, req.MaxFeePerGas, req.GasLimit, req.Nonce)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.DerivationPath == "" {
		req.DerivationPath = "m/44'/60'/0'/0/0"
	}
	var txHash string
	if req.SessionID != "" {
		txHash, err = h.walletService.SendERC20AdvancedWithSession(req.SessionID, req.DerivationPath, req.Token, req.To, amount, opts)
	} else if req.Mnemonic != "" {
		txHash, err = h.walletService.SendERC20Advanced(req.Mnemonic, req.DerivationPath, req.Token, req.To, amount, opts)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionSend, "msg": e.GetMsg(e.ErrorTransactionSend), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"tx_hash": txHash}})
}

// ApproveToken 授权
func (h *WalletHandler) ApproveToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "token 不能为空"})
		return
	}
	var req ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	amt := new(big.Int)
	if _, ok := amt.SetString(req.Amount, 10); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "amount 需要十进制字符串"})
		return
	}
	opts, err := parseTxOptions(req.GasPrice, req.MaxPriorityFeePerGas, req.MaxFeePerGas, req.GasLimit, req.Nonce)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.DerivationPath == "" {
		req.DerivationPath = "m/44'/60'/0'/0/0"
	}
	var txHash string
	if req.SessionID != "" {
		txHash, err = h.walletService.ApproveTokenWithSession(req.SessionID, req.DerivationPath, token, req.Spender, amt, opts)
	} else if req.Mnemonic != "" {
		txHash, err = h.walletService.ApproveToken(req.Mnemonic, req.DerivationPath, token, req.Spender, amt, opts)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "需要提供 session_id 或 mnemonic"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionSend, "msg": e.GetMsg(e.ErrorTransactionSend), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{"tx_hash": txHash}})
}

// GetAllowance 查询授权额度
func (h *WalletHandler) GetAllowance(c *gin.Context) {
	token := c.Param("token")
	owner := c.Query("owner")
	spender := c.Query("spender")
	if token == "" || owner == "" || spender == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": "token/owner/spender 不能为空"})
		return
	}
	val, err := h.walletService.GetAllowance(token, owner, spender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorContractCall, "msg": e.GetMsg(e.ErrorContractCall), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{
		"token": token, "owner": owner, "spender": spender, "allowance": val.String(),
	}})
}

// 解析通用高级交易选项
func parseTxOptions(gasPrice, tip, fee, gasLimit, nonce string) (*services.TxOptions, error) {
	opts := &services.TxOptions{}
	// gasPrice legacy

	if strings.TrimSpace(gasPrice) != "" {
		gp := new(big.Int)
		if _, ok := gp.SetString(gasPrice, 10); !ok {
			return nil, fmt.Errorf("gas_price 需要十进制字符串")
		}
		opts.GasPrice = gp
	}
	// EIP-1559
	if strings.TrimSpace(tip) != "" {
		t := new(big.Int)
		if _, ok := t.SetString(tip, 10); !ok {
			return nil, fmt.Errorf("max_priority_fee_per_gas 需要十进制字符串")
		}
		opts.TipCap = t
	}
	if strings.TrimSpace(fee) != "" {
		f := new(big.Int)
		if _, ok := f.SetString(fee, 10); !ok {
			return nil, fmt.Errorf("max_fee_per_gas 需要十进制字符串")
		}
		opts.FeeCap = f
	}
	// gasLimit
	if strings.TrimSpace(gasLimit) != "" {
		gl := new(big.Int)
		if _, ok := gl.SetString(gasLimit, 10); !ok {
			return nil, fmt.Errorf("gas_limit 需要十进制字符串")
		}
		opts.GasLimit = gl.Uint64()
	}
	// nonce
	if strings.TrimSpace(nonce) != "" {
		n := new(big.Int)
		if _, ok := n.SetString(nonce, 10); !ok {
			return nil, fmt.Errorf("nonce 需要十进制字符串")
		}
		u := n.Uint64()
		opts.Nonce = &u
	}
	return opts, nil
}

type SignMessageRequest struct {
	Mnemonic       string `json:"mnemonic" binding:"required"`
	DerivationPath string `json:"derivation_path"` // 默认 m/44'/60'/0'/0/0
	Message        string `json:"message" binding:"required"`
}

func (h *WalletHandler) PersonalSign(c *gin.Context) {
	var req SignMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.DerivationPath == "" {
		req.DerivationPath = "m/44'/60'/0'/0/0"
	}
	sig, addr, err := h.walletService.PersonalSign(req.Mnemonic, req.DerivationPath, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionBuild, "msg": e.GetMsg(e.ErrorTransactionBuild), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{
		"address": addr, "signature": sig,
	}})
}

type SignTypedRequest struct {
	Mnemonic       string          `json:"mnemonic" binding:"required"`
	DerivationPath string          `json:"derivation_path"`               // 默认 m/44'/60'/0'/0/0
	TypedData      json.RawMessage `json:"typed_data" binding:"required"` // 完整 EIP-712 JSON
}

func (h *WalletHandler) SignTypedDataV4(c *gin.Context) {
	var req SignTypedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": e.InvalidParams, "msg": e.GetMsg(e.InvalidParams), "data": err.Error()})
		return
	}
	if req.DerivationPath == "" {
		req.DerivationPath = "m/44'/60'/0'/0/0"
	}
	sig, addr, err := h.walletService.SignTypedDataV4(req.Mnemonic, req.DerivationPath, req.TypedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": e.ErrorTransactionBuild, "msg": e.GetMsg(e.ErrorTransactionBuild), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": e.SUCCESS, "msg": e.GetMsg(e.SUCCESS), "data": gin.H{
		"address": addr, "signature": sig,
	}})
}

// GetTransactionHistory 获取交易历史
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "地址不能为空",
		})
		return
	}

	// 解析查询参数
	req := &core.TransactionHistoryRequest{
		Address:   address,
		Page:      1,
		Limit:     20,
		TxType:    "all",
		SortBy:    "timestamp",
		SortOrder: "desc",
	}

	// 解析查询参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if txType := c.Query("tx_type"); txType != "" {
		if txType == "all" || txType == "ETH" || txType == "ERC20" || txType == "CONTRACT" {
			req.TxType = txType
		}
	}

	if startBlockStr := c.Query("start_block"); startBlockStr != "" {
		if startBlock, err := strconv.ParseUint(startBlockStr, 10, 64); err == nil {
			req.StartBlock = startBlock
		}
	}

	if endBlockStr := c.Query("end_block"); endBlockStr != "" {
		if endBlock, err := strconv.ParseUint(endBlockStr, 10, 64); err == nil {
			req.EndBlock = endBlock
		}
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		if sortBy == "timestamp" || sortBy == "block_number" {
			req.SortBy = sortBy
		}
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		if sortOrder == "asc" || sortOrder == "desc" {
			req.SortOrder = sortOrder
		}
	}

	// 查询交易历史
	resp, err := h.walletService.GetTransactionHistory(req)
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
		"data": resp,
	})
}

// CreateWallet godoc
// @Summary      Create a new wallet
// @Description  Generates a new 12-word mnemonic and derives the first address.
// @Tags         Wallets
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /wallets/new [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	mnemonic, address, err := h.walletService.CreateNewWallet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mnemonic": mnemonic,
		"address":  address,
	})
}
