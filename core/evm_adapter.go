/*
EVM适配器核心包

本包实现了与EVM（以太坊虚拟机）兼容区块链的交互功能，包括：

主要功能：
- 原生代币（ETH/MATIC/BNB等）余额查询和转账
- ERC20代币余额查询和转账
- 交易历史查询（支持分页、过滤和排序）
- Gas价格估算和交易费管理
- 智能合约调用和事件监听
- 消息签名（Personal Sign和EIP-712）

支持的区块链：
- 以太坊主网/测试网
- Polygon（MATIC）
- Binance Smart Chain（BSC）
- 其他EVM兼容链

技术特性：
- 支持Legacy和EIP-1559两种交易类型
- 智能 Gas价格调整和限制
- 安全的私钥管理和交易签名
- 完整的错误处理和日志记录
*/
package core

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// EVMAdapter EVM区块链适配器
// 封装了与以太坊及其他EVM兼容链的交互功能
// 通过RPC连接到区块链节点，提供统一的API接口
type EVMAdapter struct {
	client *ethclient.Client // 以太坊客户端，用于与区块链节点通信
}

// NewEVMAdapter 创建新的EVM适配器实例
// 参数: rpcURL - 区块链节点的RPC地址（如: https://eth.llamarpc.com）
// 返回: EVMAdapter实例指针和错误信息
// 注意: 需要确保RPC节点可访问且稳定，建议使用可靠的公共或私有节点
func NewEVMAdapter(rpcURL string) (*EVMAdapter, error) {
	// 连接到以太坊节点
	c, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接以太坊节点失败: %w", err)
	}
	return &EVMAdapter{client: c}, nil
}

// GetBalance 获取指定地址的原生代币余额
// 参数:
//
//	ctx - 上下文对象，用于取消和超时控制
//	address - 目标地址（十六进制格式，0x开头）
//
// 返回: *big.Int类型的余额（wei单位）和错误信息
// 注意: 返回的是最小单位（wei），显示时需要除以10^18转换为ETH
func (a *EVMAdapter) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	// 将字符串地址转换为以太坊地址类型
	addr := common.HexToAddress(address)
	// 查询最新区块的余额（nil表示最新）
	bal, err := a.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}
	return bal, nil
}

// SendTransaction 发送原生代币交易（实现ChainAdapter接口）
func (a *EVMAdapter) SendTransaction(ctx context.Context, from, to string, amount *big.Int, mnemonic string) (string, error) {
	// 使用默认派生路径
	return a.SendETH(ctx, mnemonic, "m/44'/60'/0'/0/0", to, amount)
}

// SendETH 发送原生代币交易（ETH/MATIC/BNB等）
// 使用助记词派生的私钥进行交易签名，支持Legacy Gas模式
// 参数:
//
//	ctx - 上下文对象
//	mnemonic - BIP39助记词
//	derivationPath - BIP44派生路径
//	to - 接收方地址
//	valueWei - 转账金额（wei单位）
//
// 返回: 交易哈希和错误信息
// 注意: 该方法会自动估算Gas限制和价格，并等待短暂时间后返回
func (a *EVMAdapter) SendETH(ctx context.Context, mnemonic, derivationPath, to string, valueWei *big.Int) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}

	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}

	nonce, err := a.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("获取nonce失败: %w", err)
	}

	toAddr := common.HexToAddress(to)

	// 估算 gas limit
	msg := ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: valueWei,
		Data:  nil,
	}
	gasLimit, err := a.client.EstimateGas(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("估算Gas失败: %w", err)
	}

	// 建议 gas price（legacy 简化）
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
	}

	// 构建与签名交易
	tx := types.NewTransaction(nonce, toAddr, valueWei, gasLimit, gasPrice, nil)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 广播交易
	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}

	// 可选：等待打包（简化为轻量等待/立即返回hash）
	_ = a.waitBrief(ctx)

	return signedTx.Hash().Hex(), nil
}

// CallContract 调用智能合约（只读）
// 参数:
//
//	ctx - 上下文对象
//	msg - 调用消息
//	blockNumber - 区块高度（nil表示最新区块）
//
// 返回: 合约返回的数据和错误信息
func (a *EVMAdapter) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return a.client.CallContract(ctx, msg, blockNumber)
}

func (a *EVMAdapter) waitBrief(ctx context.Context) error {
	// 简单小延迟，避免用户端立即查询不到
	t := time.NewTimer(500 * time.Millisecond)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

const erc20ABI = `[{"constant":true,"inputs":[{"name":"owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"},{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"type":"function"}]`

func (a *EVMAdapter) GetERC20Balance(ctx context.Context, tokenAddress, ownerAddress string) (*big.Int, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	token := common.HexToAddress(tokenAddress)
	owner := common.HexToAddress(ownerAddress)

	data, err := parsed.Pack("balanceOf", owner)
	if err != nil {
		return nil, fmt.Errorf("打包balanceOf数据失败: %w", err)
	}

	call := ethereum.CallMsg{To: &token, Data: data}
	out, err := a.client.CallContract(ctx, call, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}

	results, err := parsed.Unpack("balanceOf", out)
	if err != nil || len(results) != 1 {
		return nil, fmt.Errorf("解析balanceOf返回值失败: %w", err)
	}

	bal, ok := results[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("返回值类型错误")
	}
	return bal, nil
}

func (a *EVMAdapter) SendERC20(ctx context.Context, mnemonic, derivationPath, tokenAddress, toAddress string, amount *big.Int) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}
	nonce, err := a.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("获取nonce失败: %w", err)
	}

	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	to := common.HexToAddress(toAddress)
	token := common.HexToAddress(tokenAddress)
	data, err := parsed.Pack("transfer", to, amount)
	if err != nil {
		return "", fmt.Errorf("打包transfer数据失败: %w", err)
	}

	// 估算 gas
	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &token,
		Data: data,
	}
	gasLimit, err := a.client.EstimateGas(ctx, msg)
	if err != nil {
		return "", fmt.Errorf("估算Gas失败: %w", err)
	}

	// legacy gas 简化
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
	}

	// 构建与签名交易（value=0，调用token合约）
	tx := types.NewTransaction(nonce, token, big.NewInt(0), gasLimit, gasPrice, data)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}

	_ = a.waitBrief(ctx)
	return signedTx.Hash().Hex(), nil
}

// GasSuggestion EIP-1559/legacy 的 gas 建议
type GasSuggestion struct {
	ChainID  *big.Int
	BaseFee  *big.Int // EIP-1559 baseFee（有些链可能为 0 或不支持）
	TipCap   *big.Int // 建议的 priority fee (maxPriorityFeePerGas)
	MaxFee   *big.Int // 计算公式：tip + 2*baseFee（常见保守策略）
	GasPrice *big.Int // legacy 模式的建议 gasPrice
}

// GetNonces 获取地址的 nonce（latest 与 pending）
func (a *EVMAdapter) GetNonces(ctx context.Context, address string) (pending uint64, latest uint64, err error) {
	addr := common.HexToAddress(address)
	latest, err = a.client.NonceAt(ctx, addr, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("获取 latest nonce 失败: %w", err)
	}
	pending, err = a.client.PendingNonceAt(ctx, addr)
	if err != nil {
		return 0, 0, fmt.Errorf("获取 pending nonce 失败: %w", err)
	}
	return pending, latest, nil
}

// GetGasSuggestion 获取Gas建议（实现ChainAdapter接口）
func (a *EVMAdapter) GetGasSuggestion(ctx context.Context) (*GasSuggestion, error) {
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取链ID失败: %w", err)
	}

	// 获取基础费用（EIP-1559）
	baseFee, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		baseFee = big.NewInt(0)
	}

	// 获取小费上限（EIP-1559）
	tipCap, err := a.client.SuggestGasTipCap(ctx)
	if err != nil {
		tipCap = big.NewInt(0)
	}

	// 计算最大费用（EIP-1559）
	maxFee := new(big.Int).Add(baseFee, new(big.Int).Mul(tipCap, big.NewInt(2)))

	// 获取传统Gas价格
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		gasPrice = big.NewInt(0)
	}

	return &GasSuggestion{
		ChainID:  chainID,
		BaseFee:  baseFee,
		TipCap:   tipCap,
		MaxFee:   maxFee,
		GasPrice: gasPrice,
	}, nil
}

// EstimateGas 估算交易的 gasLimit（from/to/value/data）
func (a *EVMAdapter) EstimateGas(ctx context.Context, from, to string, value *big.Int, data []byte) (uint64, error) {
	var (
		fromAddr = common.HexToAddress(from)
		call     = ethereum.CallMsg{From: fromAddr, Value: value, Data: data}
	)
	if to != "" {
		toAddr := common.HexToAddress(to)
		call.To = &toAddr
	}
	limit, err := a.client.EstimateGas(ctx, call)
	if err != nil {
		return 0, fmt.Errorf("估算Gas失败: %w", err)
	}
	return limit, nil
}

// BroadcastRawTransaction 广播原始交易（rawTxHex）
func (a *EVMAdapter) BroadcastRawTransaction(ctx context.Context, rawTxHex string) (string, error) {
	raw := strings.TrimPrefix(strings.TrimSpace(rawTxHex), "0x")
	b, err := hexToBytes(raw)
	if err != nil {
		return "", fmt.Errorf("解析原始交易Hex失败: %w", err)
	}

	// 将 RLP 编码的已签名交易解码为 types.Transaction，然后通过 ethclient 广播
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(b); err != nil {
		return "", fmt.Errorf("解析原始交易失败: %w", err)
	}
	if err := a.client.SendTransaction(ctx, tx); err != nil {
		return "", fmt.Errorf("广播原始交易失败: %w", err)
	}
	return tx.Hash().Hex(), nil
}

// hexToBytes 简单的 hex 解码
func hexToBytes(s string) ([]byte, error) {
	if len(s)%2 == 1 {
		// 奇数长度前补 0
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

// GetTransactionReceipt 根据交易哈希查询回执
func (a *EVMAdapter) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	hash := common.HexToHash(txHash)
	receipt, err := a.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("获取交易回执失败: %w", err)
	}
	return receipt, nil
}

// GetRevertReason 当交易失败(status=0)时，尝试在交易所在区块做一次 eth_call 来还原并解析 revert reason
func (a *EVMAdapter) GetRevertReason(ctx context.Context, txHash string) (string, error) {
	hash := common.HexToHash(txHash)

	// 获取交易与回执
	tx, _, err := a.client.TransactionByHash(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("获取交易失败: %w", err)
	}
	receipt, err := a.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return "", fmt.Errorf("获取交易回执失败: %w", err)
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		return "", nil
	}

	// 还原 From（需链ID）
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}
	signer := types.LatestSignerForChainID(chainID)
	from, err := types.Sender(signer, tx)
	if err != nil {
		return "", fmt.Errorf("解析交易发送者失败: %w", err)
	}

	// 构建 call msg 并在该区块模拟执行
	var to *common.Address
	if tx.To() != nil {
		to = tx.To()
	}
	call := ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: tx.Value(),
		Data:  tx.Data(),
	}
	out, callErr := a.client.CallContract(ctx, call, receipt.BlockNumber)
	if callErr == nil {
		// 未出现错误，无法提取原因
		return "", nil
	}
	// 尝试从返回数据解码 revert reason（Error(string)/panic(uint256)）
	reason := decodeRevertReason(out)
	if reason == "" {
		// 兜底返回错误字符串
		reason = callErr.Error()
	}
	return reason, nil
}

// GetERC20Metadata 查询标准 ERC20 元数据（name/symbol/decimals）
func (a *EVMAdapter) GetERC20Metadata(ctx context.Context, tokenAddress string) (string, string, uint8, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", "", 0, fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	token := common.HexToAddress(tokenAddress)

	// name
	dataName, err := parsed.Pack("name")
	if err != nil {
		return "", "", 0, fmt.Errorf("打包name失败: %w", err)
	}
	outName, err := a.client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: dataName}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("调用name失败: %w", err)
	}
	nameVals, err := parsed.Unpack("name", outName)
	if err != nil || len(nameVals) != 1 {
		return "", "", 0, fmt.Errorf("解析name返回值失败: %w", err)
	}
	name, _ := nameVals[0].(string)

	// symbol
	dataSym, err := parsed.Pack("symbol")
	if err != nil {
		return "", "", 0, fmt.Errorf("打包symbol失败: %w", err)
	}
	outSym, err := a.client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: dataSym}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("调用symbol失败: %w", err)
	}
	symVals, err := parsed.Unpack("symbol", outSym)
	if err != nil || len(symVals) != 1 {
		return "", "", 0, fmt.Errorf("解析symbol返回值失败: %w", err)
	}
	symbol, _ := symVals[0].(string)

	// decimals
	dataDec, err := parsed.Pack("decimals")
	if err != nil {
		return "", "", 0, fmt.Errorf("打包decimals失败: %w", err)
	}
	outDec, err := a.client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: dataDec}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("调用decimals失败: %w", err)
	}
	decVals, err := parsed.Unpack("decimals", outDec)
	if err != nil || len(decVals) != 1 {
		return "", "", 0, fmt.Errorf("解析decimals返回值失败: %w", err)
	}
	var decimals uint8
	switch v := decVals[0].(type) {
	case uint8:
		decimals = v
	case *big.Int:
		decimals = uint8(v.Uint64())
	default:
		return "", "", 0, fmt.Errorf("decimals 类型不支持: %T", v)
	}

	return name, symbol, decimals, nil
}

// PersonalSign 对消息做 Ethereum Signed Message 前缀哈希后签名
func (a *EVMAdapter) PersonalSign(_ context.Context, mnemonic, derivationPath, message string) (string, string, error) {
	priv, addr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", "", err
	}
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefix))
	sig, err := crypto.Sign(hash.Bytes(), priv)
	if err != nil {
		return "", "", fmt.Errorf("签名失败: %w", err)
	}
	// v 调整为 27/28
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), addr.Hex(), nil
}

// SignTypedDataV4 对 EIP-712 typed data 进行 v4 签名（typedJSON 为完整 JSON）
func (a *EVMAdapter) SignTypedDataV4(_ context.Context, mnemonic, derivationPath string, typedJSON []byte) (string, string, error) {
	priv, addr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", "", err
	}
	var td apitypes.TypedData
	if err := json.Unmarshal(typedJSON, &td); err != nil {
		return "", "", fmt.Errorf("解析 typed data JSON 失败: %w", err)
	}
	// 计算 EIP-712 摘要: keccak256("\x19\x01" || domainSeparator || hashStruct(message))
	domainSep, err := td.HashStruct("EIP712Domain", td.Domain.Map())
	if err != nil {
		return "", "", fmt.Errorf("计算domainSeparator失败: %w", err)
	}
	msgHash, err := td.HashStruct(td.PrimaryType, td.Message)
	if err != nil {
		return "", "", fmt.Errorf("计算message hash失败: %w", err)
	}
	raw := make([]byte, 0, 2+len(domainSep)+len(msgHash))
	raw = append(raw, 0x19, 0x01)
	raw = append(raw, domainSep...)
	raw = append(raw, msgHash...)
	digest := crypto.Keccak256Hash(raw)

	sig, err := crypto.Sign(digest.Bytes(), priv)
	if err != nil {
		return "", "", fmt.Errorf("签名失败: %w", err)
	}
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), addr.Hex(), nil
}

// TxOptions 用于自定义交易参数（支持 legacy 与 EIP-1559）
type TxOptions struct {
	GasPrice *big.Int // legacy
	TipCap   *big.Int // EIP-1559 maxPriorityFeePerGas
	FeeCap   *big.Int // EIP-1559 maxFeePerGas
	GasLimit uint64   // 为 0 则自动估算
	Nonce    *uint64  // 为空则自动获取 pending
}

// SendETHWithOptions 支持自定义 gas/nonce 的 ETH 发送（自动识别 legacy/EIP-1559）
func (a *EVMAdapter) SendETHWithOptions(ctx context.Context, mnemonic, derivationPath, to string, valueWei *big.Int, opts *TxOptions) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}
	toAddr := common.HexToAddress(to)

	// nonce
	var nonce uint64
	if opts != nil && opts.Nonce != nil {
		nonce = *opts.Nonce
	} else {
		nonce, err = a.client.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			return "", fmt.Errorf("获取nonce失败: %w", err)
		}
	}

	// gasLimit
	gasLimit := uint64(0)
	if opts != nil && opts.GasLimit > 0 {
		gasLimit = opts.GasLimit
	} else {
		msg := ethereum.CallMsg{From: fromAddr, To: &toAddr, Value: valueWei}
		gl, err := a.client.EstimateGas(ctx, msg)
		if err != nil {
			return "", fmt.Errorf("估算Gas失败: %w", err)
		}
		gasLimit = gl
	}

	// 选择费率模式
	var signedTx *types.Transaction
	if opts != nil && (opts.TipCap != nil || opts.FeeCap != nil) {
		// EIP-1559
		tip := opts.TipCap
		fee := opts.FeeCap
		if tip == nil || fee == nil {
			sug, err := a.GetGasSuggestion(ctx)
			if err != nil {
				return "", err
			}
			if tip == nil {
				tip = sug.TipCap
			}
			if fee == nil {
				fee = sug.MaxFee
			}
		}
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &toAddr,
			Value:     valueWei,
			Gas:       gasLimit,
			GasFeeCap: fee,
			GasTipCap: tip,
			Data:      nil,
		})
		signer := types.LatestSignerForChainID(chainID)
		signedTx, err = types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
	} else {
		// legacy
		gp := (*big.Int)(nil)
		if opts != nil && opts.GasPrice != nil {
			gp = opts.GasPrice
		} else {
			gp, err = a.client.SuggestGasPrice(ctx)
			if err != nil {
				return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
			}
		}
		tx := types.NewTransaction(nonce, toAddr, valueWei, gasLimit, gp, nil)
		signer := types.LatestSignerForChainID(chainID)
		signedTx, err = types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
	}

	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}
	_ = a.waitBrief(ctx)
	return signedTx.Hash().Hex(), nil
}

// SendERC20WithOptions 支持自定义 gas/nonce 的 ERC20 发送（自动识别 legacy/EIP-1559）
func (a *EVMAdapter) SendERC20WithOptions(ctx context.Context, mnemonic, derivationPath, tokenAddress, toAddress string, amount *big.Int, opts *TxOptions) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}
	to := common.HexToAddress(toAddress)
	token := common.HexToAddress(tokenAddress)

	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	data, err := parsed.Pack("transfer", to, amount)
	if err != nil {
		return "", fmt.Errorf("打包transfer数据失败: %w", err)
	}

	// nonce
	var nonce uint64
	if opts != nil && opts.Nonce != nil {
		nonce = *opts.Nonce
	} else {
		nonce, err = a.client.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			return "", fmt.Errorf("获取nonce失败: %w", err)
		}
	}

	// gasLimit
	gasLimit := uint64(0)
	if opts != nil && opts.GasLimit > 0 {
		gasLimit = opts.GasLimit
	} else {
		msg := ethereum.CallMsg{From: fromAddr, To: &token, Data: data}
		gl, err := a.client.EstimateGas(ctx, msg)
		if err != nil {
			return "", fmt.Errorf("估算Gas失败: %w", err)
		}
		gasLimit = gl
	}

	// 构建签名（EIP-1559 优先）
	var signedTx *types.Transaction
	if opts != nil && (opts.TipCap != nil || opts.FeeCap != nil) {
		tip := opts.TipCap
		fee := opts.FeeCap
		if tip == nil || fee == nil {
			sug, err := a.GetGasSuggestion(ctx)
			if err != nil {
				return "", err
			}
			if tip == nil {
				tip = sug.TipCap
			}
			if fee == nil {
				fee = sug.MaxFee
			}
		}
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &token,
			Value:     big.NewInt(0),
			Gas:       gasLimit,
			GasFeeCap: fee,
			GasTipCap: tip,
			Data:      data,
		})
		signer := types.LatestSignerForChainID(chainID)
		s, err := types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
		signedTx = s
	} else {
		gp := (*big.Int)(nil)
		if opts != nil && opts.GasPrice != nil {
			gp = opts.GasPrice
		} else {
			gp, err = a.client.SuggestGasPrice(ctx)
			if err != nil {
				return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
			}
		}
		tx := types.NewTransaction(nonce, token, big.NewInt(0), gasLimit, gp, data)
		signer := types.LatestSignerForChainID(chainID)
		s, err := types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
		signedTx = s
	}

	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}
	_ = a.waitBrief(ctx)
	return signedTx.Hash().Hex(), nil
}

// Approve 授权 spender 可花费 amount
func (a *EVMAdapter) Approve(ctx context.Context, mnemonic, derivationPath, tokenAddress, spender string, amount *big.Int, opts *TxOptions) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}
	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}
	token := common.HexToAddress(tokenAddress)
	sp := common.HexToAddress(spender)

	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	data, err := parsed.Pack("approve", sp, amount)
	if err != nil {
		return "", fmt.Errorf("打包approve数据失败: %w", err)
	}

	// nonce
	var nonce uint64
	if opts != nil && opts.Nonce != nil {
		nonce = *opts.Nonce
	} else {
		nonce, err = a.client.PendingNonceAt(ctx, fromAddr)
		if err != nil {
			return "", fmt.Errorf("获取nonce失败: %w", err)
		}
	}

	// gasLimit
	gasLimit := uint64(0)
	if opts != nil && opts.GasLimit > 0 {
		gasLimit = opts.GasLimit
	} else {
		msg := ethereum.CallMsg{From: fromAddr, To: &token, Data: data}
		gl, err := a.client.EstimateGas(ctx, msg)
		if err != nil {
			return "", fmt.Errorf("估算Gas失败: %w", err)
		}
		gasLimit = gl
	}

	// 构建签名并发送
	var signedTx *types.Transaction
	if opts != nil && (opts.TipCap != nil || opts.FeeCap != nil) {
		tip := opts.TipCap
		fee := opts.FeeCap
		if tip == nil || fee == nil {
			sug, err := a.GetGasSuggestion(ctx)
			if err != nil {
				return "", err
			}
			if tip == nil {
				tip = sug.TipCap
			}
			if fee == nil {
				fee = sug.MaxFee
			}
		}
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &token,
			Value:     big.NewInt(0),
			Gas:       gasLimit,
			GasFeeCap: fee,
			GasTipCap: tip,
			Data:      data,
		})
		signer := types.LatestSignerForChainID(chainID)
		s, err := types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
		signedTx = s
	} else {
		gp := (*big.Int)(nil)
		if opts != nil && opts.GasPrice != nil {
			gp = opts.GasPrice
		} else {
			gp, err = a.client.SuggestGasPrice(ctx)
			if err != nil {
				return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
			}
		}
		tx := types.NewTransaction(nonce, token, big.NewInt(0), gasLimit, gp, data)
		signer := types.LatestSignerForChainID(chainID)
		s, err := types.SignTx(tx, signer, priv)
		if err != nil {
			return "", fmt.Errorf("签名交易失败: %w", err)
		}
		signedTx = s
	}

	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}
	_ = a.waitBrief(ctx)
	return signedTx.Hash().Hex(), nil
}

// GetAllowance 查询授权额度
func (a *EVMAdapter) GetAllowance(ctx context.Context, tokenAddress, owner, spender string) (*big.Int, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("解析ERC20 ABI失败: %w", err)
	}
	token := common.HexToAddress(tokenAddress)
	ownerAddr := common.HexToAddress(owner)
	spenderAddr := common.HexToAddress(spender)

	data, err := parsed.Pack("allowance", ownerAddr, spenderAddr)
	if err != nil {
		return nil, fmt.Errorf("打包allowance数据失败: %w", err)
	}
	out, err := a.client.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return nil, fmt.Errorf("调用合约失败: %w", err)
	}
	res, err := parsed.Unpack("allowance", out)
	if err != nil || len(res) != 1 {
		return nil, fmt.Errorf("解析allowance返回失败: %w", err)
	}
	val, ok := res[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("返回值类型错误")
	}
	return val, nil
}

// TransactionInfo 交易信息结构
type TransactionInfo struct {
	Hash        string       `json:"hash"`
	From        string       `json:"from"`
	To          string       `json:"to"`
	Value       string       `json:"value"`
	GasPrice    string       `json:"gas_price"`
	GasUsed     string       `json:"gas_used"`
	GasLimit    string       `json:"gas_limit"`
	Nonce       uint64       `json:"nonce"`
	BlockNumber string       `json:"block_number"`
	BlockHash   string       `json:"block_hash"`
	Timestamp   uint64       `json:"timestamp"`
	Status      uint64       `json:"status"`
	TxType      string       `json:"tx_type"` // "ETH", "ERC20", "CONTRACT"
	TokenInfo   *TokenTxInfo `json:"token_info,omitempty"`
}

// TokenTxInfo ERC20交易信息
type TokenTxInfo struct {
	TokenAddress string `json:"token_address"`
	TokenName    string `json:"token_name"`
	TokenSymbol  string `json:"token_symbol"`
	Decimals     uint8  `json:"decimals"`
	Amount       string `json:"amount"`
	ToAddress    string `json:"to_address"`
}

// TransactionHistoryRequest 交易历史查询请求
type TransactionHistoryRequest struct {
	Address    string `json:"address"`
	Page       int    `json:"page"`        // 页码，从1开始
	Limit      int    `json:"limit"`       // 每页数量，默认20，最大100
	TxType     string `json:"tx_type"`     // "all", "ETH", "ERC20", "CONTRACT"
	StartBlock uint64 `json:"start_block"` // 起始区块
	EndBlock   uint64 `json:"end_block"`   // 结束区块
	SortBy     string `json:"sort_by"`     // "timestamp", "block_number"
	SortOrder  string `json:"sort_order"`  // "asc", "desc"
}

// TransactionHistoryResponse 交易历史查询响应
type TransactionHistoryResponse struct {
	Transactions []TransactionInfo `json:"transactions"`
	Total        int               `json:"total"`
	Page         int               `json:"page"`
	Limit        int               `json:"limit"`
	TotalPages   int               `json:"total_pages"`
}

// GetTransactionHistory 获取地址的交易历史
func (a *EVMAdapter) GetTransactionHistory(ctx context.Context, req *TransactionHistoryRequest) (*TransactionHistoryResponse, error) {
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.SortBy == "" {
		req.SortBy = "timestamp"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// 获取当前最新区块
	latestBlock, err := a.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块失败: %w", err)
	}

	// 设置查询范围
	if req.EndBlock == 0 || req.EndBlock > latestBlock {
		req.EndBlock = latestBlock
	}
	if req.StartBlock == 0 {
		// 默认查询最近1000个区块
		if req.EndBlock > 1000 {
			req.StartBlock = req.EndBlock - 1000
		} else {
			req.StartBlock = 0
		}
	}

	// 收集交易
	transactions, err := a.collectTransactionsInRange(ctx, req.Address, req.StartBlock, req.EndBlock, req.TxType)
	if err != nil {
		return nil, err
	}

	// 排序
	a.sortTransactions(transactions, req.SortBy, req.SortOrder)

	// 分页
	total := len(transactions)
	totalPages := (total + req.Limit - 1) / req.Limit
	start := (req.Page - 1) * req.Limit
	end := start + req.Limit

	if start >= total {
		transactions = []TransactionInfo{}
	} else {
		if end > total {
			end = total
		}
		transactions = transactions[start:end]
	}

	return &TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
		Page:         req.Page,
		Limit:        req.Limit,
		TotalPages:   totalPages,
	}, nil
}

// collectTransactionsInRange 收集指定区块范围内的交易
func (a *EVMAdapter) collectTransactionsInRange(ctx context.Context, address string, startBlock, endBlock uint64, txType string) ([]TransactionInfo, error) {
	var transactions []TransactionInfo
	addr := common.HexToAddress(address)

	// 批量处理区块，避免一次查询太多
	batchSize := uint64(100)
	for current := startBlock; current <= endBlock; current += batchSize {
		batchEnd := current + batchSize - 1
		if batchEnd > endBlock {
			batchEnd = endBlock
		}

		batchTxs, err := a.collectTransactionsInBatch(ctx, addr, current, batchEnd, txType)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, batchTxs...)
	}

	return transactions, nil
}

// collectTransactionsInBatch 收集批量区块中的交易
func (a *EVMAdapter) collectTransactionsInBatch(ctx context.Context, addr common.Address, startBlock, endBlock uint64, txType string) ([]TransactionInfo, error) {
	var transactions []TransactionInfo

	for blockNum := startBlock; blockNum <= endBlock; blockNum++ {
		block, err := a.client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
		if err != nil {
			continue // 跽过获取失败的区块
		}

		for _, tx := range block.Transactions() {
			// 检查交易是否与地址相关
			isRelevant := false
			if tx.To() != nil && tx.To().Hex() == addr.Hex() {
				isRelevant = true
			}

			// 获取交易发送者
			chainID, _ := a.client.NetworkID(ctx)
			signer := types.LatestSignerForChainID(chainID)
			from, err := types.Sender(signer, tx)
			if err == nil && from.Hex() == addr.Hex() {
				isRelevant = true
			}

			if !isRelevant {
				continue
			}

			// 获取交易回执
			receipt, err := a.client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				continue
			}

			// 构建交易信息
			txInfo := a.buildTransactionInfo(tx, receipt, block, addr)

			// 过滤交易类型
			if txType != "all" && txInfo.TxType != txType {
				continue
			}

			transactions = append(transactions, *txInfo)
		}
	}

	return transactions, nil
}

// buildTransactionInfo 构建交易信息
func (a *EVMAdapter) buildTransactionInfo(tx *types.Transaction, receipt *types.Receipt, block *types.Block, userAddr common.Address) *TransactionInfo {
	txInfo := &TransactionInfo{
		Hash:        tx.Hash().Hex(),
		To:          "",
		Value:       tx.Value().String(),
		GasLimit:    new(big.Int).SetUint64(tx.Gas()).String(),
		Nonce:       tx.Nonce(),
		BlockNumber: receipt.BlockNumber.String(),
		BlockHash:   receipt.BlockHash.Hex(),
		Timestamp:   block.Time(),
		Status:      receipt.Status,
		TxType:      "ETH",
	}

	// 设置发送者
	chainID, _ := a.client.NetworkID(context.Background())
	signer := types.LatestSignerForChainID(chainID)
	from, _ := types.Sender(signer, tx)
	txInfo.From = from.Hex()

	// 设置接收者
	if tx.To() != nil {
		txInfo.To = tx.To().Hex()
	}

	// 设置gas相关信息
	if receipt.EffectiveGasPrice != nil {
		txInfo.GasPrice = receipt.EffectiveGasPrice.String()
	} else {
		txInfo.GasPrice = tx.GasPrice().String()
	}
	txInfo.GasUsed = new(big.Int).SetUint64(receipt.GasUsed).String()

	// 检查是否为ERC20交易
	if len(tx.Data()) >= 4 && tx.To() != nil {
		// 检查是否为transfer方法
		methodID := hex.EncodeToString(tx.Data()[:4])
		if methodID == "a9059cbb" { // transfer(address,uint256)
			txInfo.TxType = "ERC20"
			// 尝试解析ERC20交易信息
			if tokenInfo := a.parseERC20Transfer(tx, userAddr); tokenInfo != nil {
				txInfo.TokenInfo = tokenInfo
			}
		} else if len(tx.Data()) > 0 {
			txInfo.TxType = "CONTRACT"
		}
	}

	return txInfo
}

// parseERC20Transfer 解析ERC20转账信息
func (a *EVMAdapter) parseERC20Transfer(tx *types.Transaction, userAddr common.Address) *TokenTxInfo {
	if len(tx.Data()) < 68 { // 4 + 32 + 32
		return nil
	}

	// 解析to地址和amount
	toBytes := tx.Data()[16:36] // 跳过前16个零字节
	toAddr := common.BytesToAddress(toBytes)
	amountBytes := tx.Data()[36:68]
	amount := new(big.Int).SetBytes(amountBytes)

	// 获取代币信息
	ctx := context.Background()
	tokenAddr := tx.To().Hex()
	name, symbol, decimals, err := a.GetERC20Metadata(ctx, tokenAddr)
	if err != nil {
		// 如果获取失败，使用默认值
		name = "Unknown Token"
		symbol = "UNKNOWN"
		decimals = 18
	}

	return &TokenTxInfo{
		TokenAddress: tokenAddr,
		TokenName:    name,
		TokenSymbol:  symbol,
		Decimals:     decimals,
		Amount:       amount.String(),
		ToAddress:    toAddr.Hex(),
	}
}

// sortTransactions 排序交易
func (a *EVMAdapter) sortTransactions(transactions []TransactionInfo, sortBy, sortOrder string) {
	sort.Slice(transactions, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "timestamp":
			less = transactions[i].Timestamp < transactions[j].Timestamp
		case "block_number":
			blockI, _ := new(big.Int).SetString(transactions[i].BlockNumber, 10)
			blockJ, _ := new(big.Int).SetString(transactions[j].BlockNumber, 10)
			less = blockI.Cmp(blockJ) < 0
		default:
			less = transactions[i].Timestamp < transactions[j].Timestamp
		}

		if sortOrder == "desc" {
			return !less
		}
		return less
	})
}

// decodeRevertReason 尝试解码标准的 revert reason
func decodeRevertReason(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	// Error(string): 0x08c379a0
	if len(data) >= 4 && data[0] == 0x08 && data[1] == 0xc3 && data[2] == 0x79 && data[3] == 0xa0 {
		// 4 bytes selector + 32 offset + 32 len + bytes
		if len(data) >= 4+32+32 {
			strLen := new(big.Int).SetBytes(data[4+32 : 4+32+32]).Int64()
			start := 4 + 32 + 32
			end := start + int(strLen)
			if end <= len(data) {
				return string(data[start:end])
			}
		}
		return "execution reverted"
	}
	// Panic(uint256): 0x4e487b71
	if len(data) >= 4 && data[0] == 0x4e && data[1] == 0x48 && data[2] == 0x7b && data[3] == 0x71 {
		if len(data) >= 4+32 {
			code := new(big.Int).SetBytes(data[4 : 4+32])
			return fmt.Sprintf("panic code: 0x%x", code)
		}
		return "panic (malformed data)"
	}
	return ""
}

// SendContractTransaction 发送智能合约交易
// 参数:
//
//	ctx - 上下文对象
//	mnemonic - BIP39助记词
//	derivationPath - BIP44派生路径
//	contractAddr - 合约地址
//	data - 调用数据
//	value - 转账金额（wei单位）
//	gasLimit - Gas限制
//	gasPrice - Gas价格
//
// 返回: 交易哈希和错误信息
func (a *EVMAdapter) SendContractTransaction(ctx context.Context, mnemonic, derivationPath string, contractAddr common.Address, data []byte, value, gasLimit, gasPrice *big.Int) (string, error) {
	priv, fromAddr, err := DerivePrivateKeyFromMnemonic(mnemonic, derivationPath)
	if err != nil {
		return "", err
	}

	chainID, err := a.client.NetworkID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取链ID失败: %w", err)
	}

	nonce, err := a.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("获取nonce失败: %w", err)
	}

	// 如果未指定gasLimit，估算gas
	if gasLimit == nil || gasLimit.Cmp(big.NewInt(0)) == 0 {
		msg := ethereum.CallMsg{
			From:  fromAddr,
			To:    &contractAddr,
			Value: value,
			Data:  data,
		}
		estimatedGas, err := a.client.EstimateGas(ctx, msg)
		if err != nil {
			return "", fmt.Errorf("估算Gas失败: %w", err)
		}
		// 增加20%的安全边际
		gasLimit = new(big.Int).Mul(big.NewInt(int64(estimatedGas)), big.NewInt(120))
		gasLimit = new(big.Int).Div(gasLimit, big.NewInt(100))
	}

	// 如果未指定gasPrice，获取建议gasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice, err = a.client.SuggestGasPrice(ctx)
		if err != nil {
			return "", fmt.Errorf("获取建议GasPrice失败: %w", err)
		}
	}

	// 构建与签名交易
	tx := types.NewTransaction(nonce, contractAddr, value, gasLimit.Uint64(), gasPrice, data)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 广播交易
	if err := a.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}

	// 可选：等待打包（简化为轻量等待/立即返回hash）
	_ = a.waitBrief(ctx)

	return signedTx.Hash().Hex(), nil
}

// GetTokenBalance 获取代币余额（实现TokenSupporter接口）
func (a *EVMAdapter) GetTokenBalance(ctx context.Context, tokenAddress, ownerAddress string) (*big.Int, error) {
	return a.GetERC20Balance(ctx, tokenAddress, ownerAddress)
}

// SendTokenTransaction 发送代币交易（实现TokenSupporter接口）
func (a *EVMAdapter) SendTokenTransaction(ctx context.Context, from, to, tokenAddress string, amount *big.Int, mnemonic string) (string, error) {
	// 使用默认派生路径
	return a.SendERC20(ctx, mnemonic, "m/44'/60'/0'/0/0", tokenAddress, to, amount)
}
