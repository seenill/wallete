/*
钱包服务层包

本包提供了钱包相关的业务逻辑封装，做为API层和核心功能层之间的桥梁。

主要功能：
钱包管理：
- HD钱包创建和导入
- 助记词安全管理和加密存储
- 多地址批量生成和管理
- 会话管理（临时存储助记词）

交易处理：
- 原生代币和ERC20代币转账
- 交易费估算和优化
- 交易历史查询和管理
- 原始交易广播

多链支持：
- 动态网络切换
- 跨链余额查询
- 网络状态监控

安全特性：
- 助记词AES-GCM加密存储
- 会话超时自动清理
- 私钥在内存中的安全管理
- 只读钱包模式支持
*/
package services

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"wallet/core"
	"wallet/pkg/crypto"

	// 新增 import
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// WalletService 钱包服务核心类
// 封装了所有钱包相关的业务逻辑，包括HD钱包管理、多链支持和安全存储
type WalletService struct {
	multiChain            *core.MultiChainManager     // 多链管理器，支持动态网络切换
	sessions              map[string]sessionInfo      // 临时会话存储（助记词等敏感信息）
	watchOnly             map[string]struct{}         // 只读钱包地址集合（不包含私钥）
	encryptedWallets      map[string]*EncryptedWallet // 加密存储的钱包信息
	cryptoManager         *crypto.CryptoManager       // 加密管理器，用于助记词加密
	defiService           *DeFiService                // DeFi功能服务实例
	nftService            *NFTService                 // NFT功能服务实例
	dappBrowserService    *DAppBrowserService         // DApp浏览器服务实例
	socialService         *SocialService              // 社交功能服务实例
	securityService       *SecurityService            // 安全功能服务实例
	nftMarketplaceService *NFTMarketplaceService      // NFT市场服务实例
	mu                    sync.RWMutex                // 读写锁，保证并发安全
}

// NewWalletService 创建新的钱包服务实例
// 初始化多链管理器、加密管理器和各种存储映射
// 返回: 完全初始化的WalletService实例指针
// 注意: 如果多链管理器初始化失败会直接panic
func NewWalletService() *WalletService {
	// 初始化多链管理器，从配置文件加载网络信息
	multiChain, err := core.NewMultiChainManager()
	if err != nil {
		panic(fmt.Errorf("初始化多链管理器失败: %w", err))
	}

	// 初始化加密管理器（实际生产环境应该从配置或环境变量获取密码）
	cryptoManager := crypto.NewCryptoManager("wallet_master_key_2024")

	// 初始化DeFi服务
	defiService := NewDeFiService(multiChain)

	// 初始化NFT服务
	nftService, err := NewNFTService(multiChain)
	if err != nil {
		panic(fmt.Errorf("初始化NFT服务失败: %w", err))
	}

	// 初始化DApp浏览器
	dappBrowser := core.NewDAppBrowser(multiChain)
	dappBrowserService := NewDAppBrowserService(dappBrowser, nil) // 将在创建完WalletService后设置

	walletService := &WalletService{
		multiChain:         multiChain,
		sessions:           make(map[string]sessionInfo),
		watchOnly:          make(map[string]struct{}),
		encryptedWallets:   make(map[string]*EncryptedWallet),
		cryptoManager:      cryptoManager,
		defiService:        defiService,
		nftService:         nftService,
		dappBrowserService: dappBrowserService,
	}

	// 设置DApp浏览器服务的钱包服务引用
	dappBrowserService.walletService = walletService

	// 初始化社交服务
	socialService := NewSocialService(walletService)
	walletService.socialService = socialService

	// 初始化安全服务
	securityService := NewSecurityService(walletService)
	walletService.securityService = securityService

	// 初始化NFT市场服务
	nftMarketplaceService := NewNFTMarketplaceService(nftService)
	walletService.nftMarketplaceService = nftMarketplaceService

	return walletService
}

// ImportMnemonic 导入助记词并返回派生地址
// 参数:
//
//	mnemonic - BIP39助记词字符串
//	derivationPath - BIP44派生路径，空则使用默认路径
//
// 返回: 派生的以太坊地址和错误信息
// 注意: 该方法不会持久化存储助记词，仅用于验证和地址生成
func (s *WalletService) ImportMnemonic(mnemonic, derivationPath string) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	addr, err := core.DeriveAddressFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}
	return addr, nil
}

// GetBalance 查询地址余额（wei）
func (s *WalletService) GetBalance(address string) (*big.Int, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetBalance(ctx, address)
}

// SendETH 发送 ETH 交易，返回 txhash（MVP：使用助记词签名，不持久化）
func (s *WalletService) SendETH(mnemonic, derivationPath, to string, valueWei *big.Int) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	ctx := context.Background()
	return adapter.SendETH(ctx, mnemonic, derivationPath, to, valueWei)
}

// GetERC20Balance 查询 ERC20 余额（最小单位）
func (s *WalletService) GetERC20Balance(address, token string) (*big.Int, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetERC20Balance(ctx, token, address)
}

// SendERC20 发送 ERC20 转账
func (s *WalletService) SendERC20(mnemonic, derivationPath, token, to string, amount *big.Int) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	ctx := context.Background()
	return adapter.SendERC20(ctx, mnemonic, derivationPath, token, to, amount)
}

// GetNonces 获取地址的 latest 与 pending nonce
func (s *WalletService) GetNonces(address string) (pending uint64, latest uint64, err error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return 0, 0, err
	}
	ctx := context.Background()
	return adapter.GetNonces(ctx, address)
}

// GetGasSuggestion 获取 EIP-1559/legacy gas 建议
func (s *WalletService) GetGasSuggestion() (*core.GasSuggestion, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetGasSuggestion(ctx)
}

// EstimateGas 估算交易 gasLimit（valueWei 可为 nil 或 0，data 可为 0xHex 或 空）
func (s *WalletService) EstimateGas(from, to string, valueWei *big.Int, data []byte) (uint64, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return 0, err
	}
	ctx := context.Background()
	// 将 data 视为 hex 字符串进行解析
	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return adapter.EstimateGas(ctx, from, to, valueWei, nil)
	}
	raw = strings.TrimPrefix(raw, "0x")
	decoded, err := hexToBytes(raw)
	if err != nil {
		return 0, fmt.Errorf("解析 data(hex) 失败: %w", err)
	}
	return adapter.EstimateGas(ctx, from, to, valueWei, decoded)
}

// hexToBytes 本地解析（与 core 中一致的轻量实现）
func hexToBytes(s string) ([]byte, error) {
	if len(s)%2 == 1 {
		s = "0" + s
	}
	dst := make([]byte, len(s)/2)
	for i := 0; i < len(dst); i++ {
		var b byte
		for j := 0; j < 2; j++ {
			c := s[i*2+j]
			switch {
			case '0' <= c && c <= '9':
				b = b<<4 + (c - '0')
			case 'a' <= c && c <= 'f':
				b = b<<4 + (c - 'a' + 10)
			case 'A' <= c && c <= 'F':
				b = b<<4 + (c - 'A' + 10)
			default:
				return nil, fmt.Errorf("非法hex字符: %c", c)
			}
		}
		dst[i] = b
	}
	return dst, nil
}

// BroadcastRawTx 广播原始交易
func (s *WalletService) BroadcastRawTx(rawTxHex string) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.BroadcastRawTransaction(ctx, rawTxHex)
}

// sessionInfo 临时会话信息结构体
// 用于在内存中短暂存储助记词，并设置过期时间增强安全性
type sessionInfo struct {
	Mnemonic string    // BIP39助记词，明文存储在内存中
	ExpireAt time.Time // 会话过期时间，超过后自动失效
}

// EncryptedWallet 加密钱包信息结构体
// 用于安全存储助记词和钱包元数据，支持JSON序列化
type EncryptedWallet struct {
	ID            string                `json:"id"`             // 钱包唯一标识符（UUID或随机字符串）
	Name          string                `json:"name"`           // 钱包显示名称（用户可自定义）
	EncryptedData *crypto.EncryptedData `json:"encrypted_data"` // AES-GCM加密的助记词数据
	Addresses     []string              `json:"addresses"`      // 已派生的地址列表（为了方便查询）
	CreatedAt     time.Time             `json:"created_at"`     // 钱包创建时间
	UpdatedAt     time.Time             `json:"updated_at"`     // 钱包最后更新时间
}

// WalletInfo 钱包基本信息结构体（不包含敏感数据）
// 用于API响应和前端显示，不包含助记词或私钥等敏感信息
type WalletInfo struct {
	ID        string    `json:"id"`         // 钱包唯一标识符
	Name      string    `json:"name"`       // 钱包显示名称
	Addresses []string  `json:"addresses"`  // 已派生的地址列表
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// -------- 会话管理（助记词仅保存在内存，带过期） --------

func (s *WalletService) newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *WalletService) getSessionMnemonic(sessionID string) (string, error) {
	s.mu.RLock()
	info, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if !ok {
		return "", errors.New("session 不存在")
	}
	if time.Now().After(info.ExpireAt) {
		return "", errors.New("session 已过期")
	}
	return info.Mnemonic, nil
}

// GetSessionMnemonic 获取会话助记词（对外接口）
func (s *WalletService) GetSessionMnemonic(sessionID string) (string, error) {
	return s.getSessionMnemonic(sessionID)
}

// CreateSession 创建会话，返回 session_id 与过期时间
func (s *WalletService) CreateSession(mnemonic string, ttlSeconds int) (string, time.Time, error) {
	if mnemonic == "" {
		return "", time.Time{}, errors.New("mnemonic 不能为空")
	}
	if ttlSeconds <= 0 || ttlSeconds > 86400 {
		ttlSeconds = 900 // 默认 15 分钟
	}
	id, err := s.newSessionID()
	if err != nil {
		return "", time.Time{}, err
	}
	exp := time.Now().Add(time.Duration(ttlSeconds) * time.Second)

	s.mu.Lock()
	s.sessions[id] = sessionInfo{Mnemonic: mnemonic, ExpireAt: exp}
	s.mu.Unlock()

	return id, exp, nil
}

func (s *WalletService) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[sessionID]; !ok {
		return errors.New("session 不存在")
	}
	delete(s.sessions, sessionID)
	return nil
}

// -------- 批量地址派生（支持会话/助记词） --------

func (s *WalletService) DeriveAddressesFromMnemonic(mnemonic, pathPrefix string, start, count int) ([]string, error) {
	if pathPrefix == "" {
		pathPrefix = "m/44'/60'/0'/0"
	}
	return core.DeriveAddressesFromMnemonic(mnemonic, pathPrefix, start, count)
}

func (s *WalletService) DeriveAddressesBySession(sessionID, pathPrefix string, start, count int) ([]string, error) {
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return nil, err
	}
	return s.DeriveAddressesFromMnemonic(mn, pathPrefix, start, count)
}

// -------- 只读钱包（watch-only） --------

func (s *WalletService) AddWatchOnly(address string) (string, error) {
	if address == "" {
		return "", errors.New("address 不能为空")
	}
	checksum := common.HexToAddress(address).Hex()
	s.mu.Lock()
	s.watchOnly[checksum] = struct{}{}
	s.mu.Unlock()
	return checksum, nil
}

func (s *WalletService) RemoveWatchOnly(address string) error {
	checksum := common.HexToAddress(address).Hex()
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.watchOnly[checksum]; !ok {
		return errors.New("address 不存在")
	}
	delete(s.watchOnly, checksum)
	return nil
}

func (s *WalletService) ListWatchOnly() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]string, 0, len(s.watchOnly))
	for a := range s.watchOnly {
		res = append(res, a)
	}
	return res
}

// -------- 基于会话的发送（免提交助记词） --------

func (s *WalletService) SendETHWithSession(sessionID, derivationPath, to string, valueWei *big.Int) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return "", err
	}
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.SendETH(ctx, mn, derivationPath, to, valueWei)
}

func (s *WalletService) SendERC20WithSession(sessionID, derivationPath, token, to string, amount *big.Int) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return "", err
	}
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.SendERC20(ctx, mn, derivationPath, token, to, amount)
}

type TxReceiptDTO struct {
	TxHash            string `json:"tx_hash"`
	Status            uint64 `json:"status"`
	BlockNumber       string `json:"block_number"`
	GasUsed           string `json:"gas_used"`
	EffectiveGasPrice string `json:"effective_gas_price,omitempty"`
	ContractAddress   string `json:"contract_address,omitempty"`
	TransactionIndex  uint   `json:"transaction_index"`
	RevertReason      string `json:"revert_reason,omitempty"`
}

func (s *WalletService) GetReceipt(txHash string) (*TxReceiptDTO, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	receipt, err := adapter.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}
	dto := &TxReceiptDTO{
		TxHash:           txHash,
		Status:           receipt.Status,
		BlockNumber:      receipt.BlockNumber.String(),
		GasUsed:          new(big.Int).SetUint64(receipt.GasUsed).String(),
		TransactionIndex: uint(receipt.TransactionIndex),
	}
	if receipt.EffectiveGasPrice != nil {
		dto.EffectiveGasPrice = receipt.EffectiveGasPrice.String()
	}
	if receipt.ContractAddress != (common.Address{}) {
		// 这里无 CommonAddressZero，直接判断是否为 0 地址
		if receipt.ContractAddress.Hex() != "0x0000000000000000000000000000000000000000" {
			dto.ContractAddress = receipt.ContractAddress.Hex()
		}
	}
	// 失败时尝试提取 revert reason
	if receipt.Status == 0 {
		reason, _ := adapter.GetRevertReason(ctx, txHash)
		dto.RevertReason = reason
	}
	return dto, nil
}

func (s *WalletService) GetTokenMetadata(token string) (name, symbol string, decimals uint8, err error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", "", 0, err
	}
	ctx := context.Background()
	return adapter.GetERC20Metadata(ctx, token)
}

func (s *WalletService) PersonalSign(mnemonic, derivationPath, message string) (sigHex, address string, err error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", "", err
	}
	ctx := context.Background()
	return adapter.PersonalSign(ctx, mnemonic, derivationPath, message)
}

func (s *WalletService) SignTypedDataV4(mnemonic, derivationPath string, typedJSON []byte) (sigHex, address string, err error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", "", err
	}
	ctx := context.Background()
	return adapter.SignTypedDataV4(ctx, mnemonic, derivationPath, typedJSON)
}

// TxOptions 服务层版本，避免 handler 直接依赖 core
type TxOptions struct {
	GasPrice *big.Int
	TipCap   *big.Int
	FeeCap   *big.Int
	GasLimit uint64
	Nonce    *uint64
}

func (s *WalletService) toCoreTxOptions(o *TxOptions) *core.TxOptions {
	if o == nil {
		return nil
	}
	return &core.TxOptions{
		GasPrice: o.GasPrice,
		TipCap:   o.TipCap,
		FeeCap:   o.FeeCap,
		GasLimit: o.GasLimit,
		Nonce:    o.Nonce,
	}
}

// 高级发送 ETH（支持 TxOptions）
func (s *WalletService) SendETHAdvanced(mnemonic, derivationPath, to string, valueWei *big.Int, opts *TxOptions) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.SendETHWithOptions(ctx, mnemonic, derivationPath, to, valueWei, s.toCoreTxOptions(opts))
}

func (s *WalletService) SendETHAdvancedWithSession(sessionID, derivationPath, to string, valueWei *big.Int, opts *TxOptions) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return "", err
	}
	return s.SendETHAdvanced(mn, derivationPath, to, valueWei, opts)
}

// 高级发送 ERC20（支持 TxOptions）
func (s *WalletService) SendERC20Advanced(mnemonic, derivationPath, token, to string, amount *big.Int, opts *TxOptions) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.SendERC20WithOptions(ctx, mnemonic, derivationPath, token, to, amount, s.toCoreTxOptions(opts))
}

func (s *WalletService) SendERC20AdvancedWithSession(sessionID, derivationPath, token, to string, amount *big.Int, opts *TxOptions) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return "", err
	}
	return s.SendERC20Advanced(mn, derivationPath, token, to, amount, opts)
}

// ERC20 授权 approve
func (s *WalletService) ApproveToken(mnemonic, derivationPath, token, spender string, amount *big.Int, opts *TxOptions) (string, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	return adapter.Approve(ctx, mnemonic, derivationPath, token, spender, amount, s.toCoreTxOptions(opts))
}

func (s *WalletService) ApproveTokenWithSession(sessionID, derivationPath, token, spender string, amount *big.Int, opts *TxOptions) (string, error) {
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	mn, err := s.getSessionMnemonic(sessionID)
	if err != nil {
		return "", err
	}
	return s.ApproveToken(mn, derivationPath, token, spender, amount, opts)
}

// ERC20 allowance 读取
func (s *WalletService) GetAllowance(token, owner, spender string) (*big.Int, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetAllowance(ctx, token, owner, spender)
}

// GetTransactionHistory 获取交易历史
func (s *WalletService) GetTransactionHistory(req *core.TransactionHistoryRequest) (*core.TransactionHistoryResponse, error) {
	adapter, err := s.multiChain.GetCurrentAdapter()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetTransactionHistory(ctx, req)
}

// -------- 多链管理 --------

// GetCurrentNetwork 获取当前网络
func (s *WalletService) GetCurrentNetwork() string {
	return s.multiChain.GetCurrentNetwork()
}

// SwitchNetwork 切换网络
func (s *WalletService) SwitchNetwork(networkID string) error {
	return s.multiChain.SwitchNetwork(networkID)
}

// GetAvailableNetworks 获取所有可用网络
func (s *WalletService) GetAvailableNetworks() []string {
	return s.multiChain.GetAvailableNetworks()
}

// GetNetworkInfo 获取网络信息
func (s *WalletService) GetNetworkInfo(networkID string) (*core.NetworkInfo, error) {
	return s.multiChain.GetNetworkInfo(networkID)
}

// GetAllNetworksInfo 获取所有网络信息
func (s *WalletService) GetAllNetworksInfo() (map[string]*core.NetworkInfo, error) {
	networks := s.multiChain.GetAvailableNetworks()
	networksInfo := make(map[string]*core.NetworkInfo)

	for _, networkID := range networks {
		info, err := s.multiChain.GetNetworkInfo(networkID)
		if err != nil {
			// 记录错误但继续处理其他网络
			continue
		}
		networksInfo[networkID] = info
	}

	return networksInfo, nil
}

// GetCrossChainBalance 获取跨链余额
func (s *WalletService) GetCrossChainBalance(address string, networks []string) (map[string]*big.Int, error) {
	return s.multiChain.GetCrossChainBalance(address, networks)
}

// GetCrossChainTokenBalance 获取跨链代币余额
func (s *WalletService) GetCrossChainTokenBalance(address, tokenAddress string, networks []string) (map[string]*big.Int, error) {
	return s.multiChain.GetCrossChainTokenBalance(address, tokenAddress, networks)
}

// GetBalanceOnNetwork 获取指定网络上的余额
func (s *WalletService) GetBalanceOnNetwork(address, networkID string) (*big.Int, error) {
	adapter, err := s.multiChain.GetAdapter(networkID)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return adapter.GetBalance(ctx, address)
}

// SendETHOnNetwork 在指定网络上发送ETH
func (s *WalletService) SendETHOnNetwork(networkID, mnemonic, derivationPath, to string, valueWei *big.Int) (string, error) {
	adapter, err := s.multiChain.GetAdapter(networkID)
	if err != nil {
		return "", err
	}
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	ctx := context.Background()
	return adapter.SendETH(ctx, mnemonic, derivationPath, to, valueWei)
}

// SendERC20OnNetwork 在指定网络上发送ERC20
func (s *WalletService) SendERC20OnNetwork(networkID, mnemonic, derivationPath, token, to string, amount *big.Int) (string, error) {
	adapter, err := s.multiChain.GetAdapter(networkID)
	if err != nil {
		return "", err
	}
	if derivationPath == "" {
		derivationPath = "m/44'/60'/0'/0/0"
	}
	ctx := context.Background()
	return adapter.SendERC20(ctx, mnemonic, derivationPath, token, to, amount)
}

// -------- 加密钱包管理 --------

// CreateEncryptedWallet 创建加密钱包
func (s *WalletService) CreateEncryptedWallet(name, password string, addressCount int) (*WalletInfo, error) {
	if addressCount <= 0 {
		addressCount = 1
	}

	// 生成助记词
	mnemonic, err := core.GenerateMnemonic(128)
	if err != nil {
		return nil, fmt.Errorf("生成助记词失败: %w", err)
	}

	// 加密助记词
	encryptedData, err := s.cryptoManager.EncryptWithPassword(mnemonic, password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	// 派生地址
	addresses, err := core.DeriveAddressesFromMnemonic(mnemonic, "m/44'/60'/0'/0", 0, addressCount)
	if err != nil {
		return nil, fmt.Errorf("派生地址失败: %w", err)
	}

	// 生成钱包ID
	walletID, err := s.newSessionID() // 复用已有的随机ID生成方法
	if err != nil {
		return nil, fmt.Errorf("生成钱包ID失败: %w", err)
	}

	now := time.Now()
	encryptedWallet := &EncryptedWallet{
		ID:            walletID,
		Name:          name,
		EncryptedData: encryptedData,
		Addresses:     addresses,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 存储加密钱包
	s.mu.Lock()
	s.encryptedWallets[walletID] = encryptedWallet
	s.mu.Unlock()

	return &WalletInfo{
		ID:        walletID,
		Name:      name,
		Addresses: addresses,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ImportEncryptedWallet 导入助记词并创建加密钱包
func (s *WalletService) ImportEncryptedWallet(name, mnemonic, password string, addressCount int) (*WalletInfo, error) {
	if addressCount <= 0 {
		addressCount = 1
	}

	// 验证助记词
	_, err := core.DeriveAddressFromMnemonic(mnemonic, "m/44'/60'/0'/0/0")
	if err != nil {
		return nil, fmt.Errorf("无效的助记词: %w", err)
	}

	// 加密助记词
	encryptedData, err := s.cryptoManager.EncryptWithPassword(mnemonic, password)
	if err != nil {
		return nil, fmt.Errorf("加密助记词失败: %w", err)
	}

	// 派生地址
	addresses, err := core.DeriveAddressesFromMnemonic(mnemonic, "m/44'/60'/0'/0", 0, addressCount)
	if err != nil {
		return nil, fmt.Errorf("派生地址失败: %w", err)
	}

	// 生成钱包ID
	walletID, err := s.newSessionID()
	if err != nil {
		return nil, fmt.Errorf("生成钱包ID失败: %w", err)
	}

	now := time.Now()
	encryptedWallet := &EncryptedWallet{
		ID:            walletID,
		Name:          name,
		EncryptedData: encryptedData,
		Addresses:     addresses,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 存储加密钱包
	s.mu.Lock()
	s.encryptedWallets[walletID] = encryptedWallet
	s.mu.Unlock()

	return &WalletInfo{
		ID:        walletID,
		Name:      name,
		Addresses: addresses,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetEncryptedWallet 获取加密钱包信息
func (s *WalletService) GetEncryptedWallet(walletID string) (*WalletInfo, error) {
	s.mu.RLock()
	encWallet, exists := s.encryptedWallets[walletID]
	s.mu.RUnlock()

	if !exists {
		return nil, errors.New("钱包不存在")
	}

	return &WalletInfo{
		ID:        encWallet.ID,
		Name:      encWallet.Name,
		Addresses: encWallet.Addresses,
		CreatedAt: encWallet.CreatedAt,
		UpdatedAt: encWallet.UpdatedAt,
	}, nil
}

// ListEncryptedWallets 列出所有加密钱包
func (s *WalletService) ListEncryptedWallets() ([]*WalletInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallets := make([]*WalletInfo, 0, len(s.encryptedWallets))
	for _, encWallet := range s.encryptedWallets {
		wallets = append(wallets, &WalletInfo{
			ID:        encWallet.ID,
			Name:      encWallet.Name,
			Addresses: encWallet.Addresses,
			CreatedAt: encWallet.CreatedAt,
			UpdatedAt: encWallet.UpdatedAt,
		})
	}

	return wallets, nil
}

// UnlockWallet 解锁钱包获取助记词
func (s *WalletService) UnlockWallet(walletID, password string) (string, error) {
	s.mu.RLock()
	encWallet, exists := s.encryptedWallets[walletID]
	s.mu.RUnlock()

	if !exists {
		return "", errors.New("钱包不存在")
	}

	// 解密助记词
	mnemonic, err := s.cryptoManager.DecryptWithPassword(encWallet.EncryptedData, password)
	if err != nil {
		return "", fmt.Errorf("密码错误或解密失败: %w", err)
	}

	return mnemonic, nil
}

// DeleteEncryptedWallet 删除加密钱包
func (s *WalletService) DeleteEncryptedWallet(walletID, password string) error {
	// 验证密码
	_, err := s.UnlockWallet(walletID, password)
	if err != nil {
		return err
	}

	// 删除钱包
	s.mu.Lock()
	delete(s.encryptedWallets, walletID)
	s.mu.Unlock()

	return nil
}

// SendETHWithEncryptedWallet 使用加密钱包发送ETH
func (s *WalletService) SendETHWithEncryptedWallet(walletID, password, derivationPath, to string, valueWei *big.Int) (string, error) {
	// 解锁钱包
	mnemonic, err := s.UnlockWallet(walletID, password)
	if err != nil {
		return "", err
	}

	// 发送交易
	return s.SendETH(mnemonic, derivationPath, to, valueWei)
}

// SendERC20WithEncryptedWallet 使用加密钱包发送ERC20
func (s *WalletService) SendERC20WithEncryptedWallet(walletID, password, derivationPath, token, to string, amount *big.Int) (string, error) {
	// 解锁钱包
	mnemonic, err := s.UnlockWallet(walletID, password)
	if err != nil {
		return "", err
	}

	// 发送交易
	return s.SendERC20(mnemonic, derivationPath, token, to, amount)
}

// CreateNewWallet 生成新的助记词和地址
func (s *WalletService) CreateNewWallet() (mnemonic, address string, err error) {
	mnemonic, err = core.GenerateMnemonic(128)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}
	address, err = core.DeriveAddressFromMnemonic(mnemonic, "m/44'/60'/0'/0/0")
	if err != nil {
		return "", "", fmt.Errorf("failed to derive address: %w", err)
	}
	return mnemonic, address, nil
}

// -------- 服务实例获取 --------

// GetDeFiService 获取DeFi服务实例
func (s *WalletService) GetDeFiService() *DeFiService {
	return s.defiService
}

// GetNFTService 获取NFT服务实例
func (s *WalletService) GetNFTService() *NFTService {
	return s.nftService
}

// GetDAppBrowserService 获取DApp浏览器服务实例
func (s *WalletService) GetDAppBrowserService() *DAppBrowserService {
	return s.dappBrowserService
}

// GetSocialService 获取社交服务实例
func (s *WalletService) GetSocialService() *SocialService {
	return s.socialService
}

// GetSecurityService 获取安全服务实例
func (s *WalletService) GetSecurityService() *SecurityService {
	return s.securityService
}

// GetNFTMarketplaceService 获取NFT市场服务实例
func (s *WalletService) GetNFTMarketplaceService() *NFTMarketplaceService {
	return s.nftMarketplaceService
}

// IsValidAddress 验证地址格式
func (s *WalletService) IsValidAddress(address string) bool {
	return common.IsHexAddress(address)
}
