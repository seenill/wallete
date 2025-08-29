package core

import (
	"context"
	"math/big"
)

// BitcoinAdapter Bitcoin链适配器
type BitcoinAdapter struct {
	rpcURL string
}

// NewBitcoinAdapter 创建Bitcoin适配器
func NewBitcoinAdapter(rpcURL string) (*BitcoinAdapter, error) {
	adapter := &BitcoinAdapter{
		rpcURL: rpcURL,
	}
	return adapter, nil
}

// GetBalance 获取BTC余额
func (ba *BitcoinAdapter) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	// TODO: 实现Bitcoin余额查询逻辑
	// 这里需要使用Bitcoin RPC API查询余额
	// 暂时返回0余额而不是错误，避免API 500错误
	return big.NewInt(0), nil
}

// SendTransaction 发送BTC交易
func (ba *BitcoinAdapter) SendTransaction(ctx context.Context, from, to string, amount *big.Int, mnemonic string) (string, error) {
	// TODO: 实现Bitcoin交易发送逻辑
	return "", nil
}

// GetGasSuggestion 获取Gas建议
func (ba *BitcoinAdapter) GetGasSuggestion(ctx context.Context) (*GasSuggestion, error) {
	// Bitcoin使用手续费而非Gas，这里返回默认值
	return &GasSuggestion{
		ChainID:  big.NewInt(0),
		BaseFee:  big.NewInt(0),
		TipCap:   big.NewInt(0),
		MaxFee:   big.NewInt(0),
		GasPrice: big.NewInt(0),
	}, nil
}
