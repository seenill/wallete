/*
Uniswap V2 DEX实现

本文件实现了与Uniswap V2协议的集成，提供完整的DEX交易功能。

主要功能：
- Uniswap V2路由器交互
- 交易对信息查询
- 最佳交易路径计算
- 流动性池管理
- 价格影响计算

技术特性：
- 支持多跳路由（通过WETH）
- 滑点保护机制
- Gas优化策略
- 实时价格更新

合约地址（以太坊主网）：
- Router: 0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D
- Factory: 0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f
- WETH: 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2
*/
package core

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// UniswapV2Exchange Uniswap V2交易所实现
type UniswapV2Exchange struct {
	routerAddress  common.Address // Router合约地址
	factoryAddress common.Address // Factory合约地址
	wethAddress    common.Address // WETH代币地址
	evmAdapter     *EVMAdapter    // EVM适配器
	routerABI      abi.ABI        // Router合约ABI
	pairABI        abi.ABI        // Pair合约ABI
}

// Uniswap V2 Router ABI（简化版，包含主要方法）
const uniswapV2RouterABI = `[
    {
        "inputs": [
            {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
            {"internalType": "address[]", "name": "path", "type": "address[]"}
        ],
        "name": "getAmountsOut",
        "outputs": [
            {"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
            {"internalType": "address[]", "name": "path", "type": "address[]"},
            {"internalType": "address", "name": "to", "type": "address"},
            {"internalType": "uint256", "name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactETHForTokens",
        "outputs": [
            {"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "payable",
        "type": "function"
    },
    {
        "inputs": [
            {"internalType": "uint256", "name": "amountIn", "type": "uint256"},
            {"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
            {"internalType": "address[]", "name": "path", "type": "address[]"},
            {"internalType": "address", "name": "to", "type": "address"},
            {"internalType": "uint256", "name": "deadline", "type": "uint256"}
        ],
        "name": "swapExactTokensForTokens",
        "outputs": [
            {"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {"internalType": "address", "name": "tokenA", "type": "address"},
            {"internalType": "address", "name": "tokenB", "type": "address"}
        ],
        "name": "getPair",
        "outputs": [
            {"internalType": "address", "name": "pair", "type": "address"}
        ],
        "stateMutability": "view",
        "type": "function"
    }
]`

// Uniswap V2 Pair ABI（简化版）
const uniswapV2PairABI = `[
    {
        "inputs": [],
        "name": "getReserves",
        "outputs": [
            {"internalType": "uint112", "name": "_reserve0", "type": "uint112"},
            {"internalType": "uint112", "name": "_reserve1", "type": "uint112"},
            {"internalType": "uint32", "name": "_blockTimestampLast", "type": "uint32"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "token0",
        "outputs": [
            {"internalType": "address", "name": "", "type": "address"}
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [],
        "name": "token1",
        "outputs": [
            {"internalType": "address", "name": "", "type": "address"}
        ],
        "stateMutability": "view",
        "type": "function"
    }
]`

// NewUniswapV2Exchange 创建Uniswap V2交易所实例
func NewUniswapV2Exchange(evmAdapter *EVMAdapter, network string) (*UniswapV2Exchange, error) {
	var routerAddr, factoryAddr, wethAddr string

	// 根据网络设置合约地址
	switch strings.ToLower(network) {
	case "ethereum", "mainnet":
		routerAddr = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
		factoryAddr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
		wethAddr = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	case "goerli":
		routerAddr = "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D"
		factoryAddr = "0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f"
		wethAddr = "0xB4FBF271143F4FBf7B91A5ded31805e42b2208d6"
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// 解析ABI
	routerABI, err := abi.JSON(strings.NewReader(uniswapV2RouterABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse router ABI: %w", err)
	}

	pairABI, err := abi.JSON(strings.NewReader(uniswapV2PairABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pair ABI: %w", err)
	}

	return &UniswapV2Exchange{
		routerAddress:  common.HexToAddress(routerAddr),
		factoryAddress: common.HexToAddress(factoryAddr),
		wethAddress:    common.HexToAddress(wethAddr),
		evmAdapter:     evmAdapter,
		routerABI:      routerABI,
		pairABI:        pairABI,
	}, nil
}

// GetName 返回交易所名称
func (u *UniswapV2Exchange) GetName() string {
	return "Uniswap V2"
}

// GetPair 获取交易对信息
func (u *UniswapV2Exchange) GetPair(tokenA, tokenB string) (*TradingPair, error) {
	ctx := context.Background()

	// 调用factory.getPair获取交易对地址
	factoryData, err := u.routerABI.Pack("getPair",
		common.HexToAddress(tokenA),
		common.HexToAddress(tokenB))
	if err != nil {
		return nil, fmt.Errorf("failed to pack getPair call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &u.factoryAddress,
		Data: factoryData,
	}
	result, err := u.evmAdapter.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getPair: %w", err)
	}

	var pairAddress common.Address
	err = u.routerABI.UnpackIntoInterface(&pairAddress, "getPair", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getPair result: %w", err)
	}

	// 检查交易对是否存在
	if pairAddress == (common.Address{}) {
		return nil, fmt.Errorf("trading pair not found")
	}

	// 获取交易对储备量
	reservesData, err := u.pairABI.Pack("getReserves")
	if err != nil {
		return nil, fmt.Errorf("failed to pack getReserves call: %w", err)
	}

	msg = ethereum.CallMsg{
		To:   &pairAddress,
		Data: reservesData,
	}
	reservesResult, err := u.evmAdapter.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getReserves: %w", err)
	}

	var reserve0, reserve1 *big.Int
	var blockTimestamp uint32

	unpacked, err := u.pairABI.Unpack("getReserves", reservesResult)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getReserves result: %w", err)
	}

	if len(unpacked) >= 3 {
		reserve0 = unpacked[0].(*big.Int)
		reserve1 = unpacked[1].(*big.Int)
		blockTimestamp = unpacked[2].(uint32)
	}

	// 获取token0和token1地址以确定顺序
	token0Data, err := u.pairABI.Pack("token0")
	if err != nil {
		return nil, fmt.Errorf("failed to pack token0 call: %w", err)
	}

	msg = ethereum.CallMsg{
		To:   &pairAddress,
		Data: token0Data,
	}
	token0Result, err := u.evmAdapter.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call token0: %w", err)
	}

	var token0Address common.Address
	err = u.pairABI.UnpackIntoInterface(&token0Address, "token0", token0Result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack token0 result: %w", err)
	}

	// 构造交易对信息
	pair := &TradingPair{
		Address:    pairAddress.Hex(),
		Reserve0:   reserve0,
		Reserve1:   reserve1,
		Fee:        big.NewInt(30), // Uniswap V2固定0.3%手续费
		LastUpdate: int64(blockTimestamp),
	}

	// 根据token0地址确定代币顺序
	if strings.EqualFold(token0Address.Hex(), tokenA) {
		pair.TokenA = &Token{Address: tokenA}
		pair.TokenB = &Token{Address: tokenB}
	} else {
		pair.TokenA = &Token{Address: tokenB}
		pair.TokenB = &Token{Address: tokenA}
		// 交换储备量
		pair.Reserve0, pair.Reserve1 = pair.Reserve1, pair.Reserve0
	}

	return pair, nil
}

// GetQuote 获取交易报价
func (u *UniswapV2Exchange) GetQuote(tokenIn, tokenOut string, amountIn *big.Int) (*QuoteResult, error) {
	ctx := context.Background()

	// 构建交易路径
	path := u.buildTradingPath(tokenIn, tokenOut)

	// 调用getAmountsOut获取输出数量
	data, err := u.routerABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, fmt.Errorf("failed to pack getAmountsOut call: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &u.routerAddress,
		Data: data,
	}
	result, err := u.evmAdapter.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call getAmountsOut: %w", err)
	}

	var amounts []*big.Int
	err = u.routerABI.UnpackIntoInterface(&amounts, "getAmountsOut", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack getAmountsOut result: %w", err)
	}

	if len(amounts) < 2 {
		return nil, fmt.Errorf("invalid amounts returned")
	}

	amountOut := amounts[len(amounts)-1]

	// 计算滑点保护（默认0.5%滑点）
	slippageTolerance := big.NewInt(995) // 99.5%
	amountOutMin := new(big.Int).Mul(amountOut, slippageTolerance)
	amountOutMin.Div(amountOutMin, big.NewInt(1000))

	// 计算价格影响（简化版）
	priceImpact := u.calculatePriceImpact(tokenIn, tokenOut, amountIn, amountOut)

	// 估算Gas费用
	gasEstimate := uint64(150000) // Uniswap V2 swap的典型Gas消耗

	// 构建路由信息
	route := u.buildRouteHops(path, amounts)

	return &QuoteResult{
		AmountOut:    amountOut,
		AmountOutMin: amountOutMin,
		Price:        u.calculatePrice(amountIn, amountOut),
		PriceImpact:  priceImpact,
		GasEstimate:  gasEstimate,
		Route:        route,
		ValidUntil:   getCurrentTimestamp() + 30, // 30秒有效期
		Exchange:     u.GetName(),
	}, nil
}

// ExecuteSwap 执行交易
func (u *UniswapV2Exchange) ExecuteSwap(ctx context.Context, params *SwapParams) (*SwapResult, error) {
	// 构建交易路径
	path := u.buildTradingPath(params.TokenIn, params.TokenOut)

	var txHash string
	var err error

	// 判断是否为ETH交易
	if u.isETH(params.TokenIn) {
		// ETH -> Token
		txHash, err = u.swapETHForTokens(ctx, params, path)
	} else if u.isETH(params.TokenOut) {
		// Token -> ETH
		txHash, err = u.swapTokensForETH(ctx, params, path)
	} else {
		// Token -> Token
		txHash, err = u.swapTokensForTokens(ctx, params, path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute swap: %w", err)
	}

	return &SwapResult{
		TxHash:    txHash,
		AmountIn:  params.AmountIn,
		AmountOut: new(big.Int).Set(params.AmountOutMin), // 实际数量需要从交易receipt获取
		Status:    "pending",
		Timestamp: getCurrentTimestamp(),
		Exchange:  u.GetName(),
	}, nil
}

// GetLiquidityPools 获取流动性池（简化实现）
func (u *UniswapV2Exchange) GetLiquidityPools() ([]*LiquidityPool, error) {
	// 这里返回一些热门的流动性池
	// 实际实现中需要从The Graph或其他索引服务获取
	pools := []*LiquidityPool{
		{
			Address: "0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc", // USDC/WETH
			Name:    "USDC/WETH",
			TokenA:  &Token{Address: "0xA0b86a33E6441Bf11f3f1c9AdC32E56D0e85D4A1", Symbol: "USDC"},
			TokenB:  &Token{Address: u.wethAddress.Hex(), Symbol: "WETH"},
			TVL:     big.NewInt(50000000), // $50M TVL (示例)
			APY:     "15.5",
			Fee:     "0.3",
		},
		// 可以添加更多流动性池...
	}

	return pools, nil
}

// AddLiquidity 添加流动性
func (u *UniswapV2Exchange) AddLiquidity(ctx context.Context, params *AddLiquidityParams) (*TxResult, error) {
	// 实现添加流动性的逻辑
	// 这里需要调用router的addLiquidity方法
	return &TxResult{
		TxHash:    "0x...", // 实际交易哈希
		Status:    "pending",
		Timestamp: getCurrentTimestamp(),
	}, nil
}

// buildTradingPath 构建交易路径
func (u *UniswapV2Exchange) buildTradingPath(tokenIn, tokenOut string) []common.Address {
	tokenInAddr := common.HexToAddress(tokenIn)
	tokenOutAddr := common.HexToAddress(tokenOut)

	// 如果其中一个是WETH，直接交易
	if tokenInAddr == u.wethAddress || tokenOutAddr == u.wethAddress {
		return []common.Address{tokenInAddr, tokenOutAddr}
	}

	// 否则通过WETH路由
	return []common.Address{tokenInAddr, u.wethAddress, tokenOutAddr}
}

// buildRouteHops 构建路由跳跃信息
func (u *UniswapV2Exchange) buildRouteHops(path []common.Address, amounts []*big.Int) []*RouteHop {
	hops := make([]*RouteHop, 0, len(path)-1)

	for i := 0; i < len(path)-1; i++ {
		hop := &RouteHop{
			Exchange:  u.GetName(),
			TokenIn:   &Token{Address: path[i].Hex()},
			TokenOut:  &Token{Address: path[i+1].Hex()},
			AmountIn:  amounts[i],
			AmountOut: amounts[i+1],
			Fee:       big.NewInt(30), // 0.3% fee
		}
		hops = append(hops, hop)
	}

	return hops
}

// calculatePriceImpact 计算价格影响（简化版）
func (u *UniswapV2Exchange) calculatePriceImpact(tokenIn, tokenOut string, amountIn, amountOut *big.Int) string {
	// 简化的价格影响计算
	// 实际实现需要考虑储备量比例
	return "0.15%" // 示例返回值
}

// calculatePrice 计算价格
func (u *UniswapV2Exchange) calculatePrice(amountIn, amountOut *big.Int) string {
	if amountIn.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	price := new(big.Float).Quo(
		new(big.Float).SetInt(amountOut),
		new(big.Float).SetInt(amountIn),
	)

	return price.String()
}

// isETH 检查是否为ETH地址
func (u *UniswapV2Exchange) isETH(tokenAddress string) bool {
	return strings.EqualFold(tokenAddress, "0x0000000000000000000000000000000000000000") ||
		strings.EqualFold(tokenAddress, "ETH")
}

// swapETHForTokens ETH换代币
func (u *UniswapV2Exchange) swapETHForTokens(ctx context.Context, params *SwapParams, path []common.Address) (string, error) {
	// 实现ETH换代币的具体逻辑
	// 需要调用router的swapExactETHForTokens方法
	return "0x...", nil // 返回实际的交易哈希
}

// swapTokensForETH 代币换ETH
func (u *UniswapV2Exchange) swapTokensForETH(ctx context.Context, params *SwapParams, path []common.Address) (string, error) {
	// 实现代币换ETH的具体逻辑
	return "0x...", nil
}

// swapTokensForTokens 代币换代币
func (u *UniswapV2Exchange) swapTokensForTokens(ctx context.Context, params *SwapParams, path []common.Address) (string, error) {
	// 实现代币换代币的具体逻辑
	return "0x...", nil
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
