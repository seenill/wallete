package core

import (
	"context"
	"math/big"
)

// ChainAdapter 区块链适配器接口
// 定义了所有区块链适配器需要实现的方法
type ChainAdapter interface {
	// GetBalance 获取地址余额
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// SendTransaction 发送交易
	SendTransaction(ctx context.Context, from, to string, amount *big.Int, mnemonic string) (string, error)

	// GetGasSuggestion 获取Gas建议
	GetGasSuggestion(ctx context.Context) (*GasSuggestion, error)
}

// TokenSupporter 支持代币操作的链适配器接口
type TokenSupporter interface {
	ChainAdapter

	// GetTokenBalance 获取代币余额
	GetTokenBalance(ctx context.Context, tokenAddress, ownerAddress string) (*big.Int, error)

	// SendTokenTransaction 发送代币交易
	SendTokenTransaction(ctx context.Context, from, to, tokenAddress string, amount *big.Int, mnemonic string) (string, error)
}
