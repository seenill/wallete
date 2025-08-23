package core

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"wallet/config"
)

// MultiChainManager 多链管理器
type MultiChainManager struct {
	adapters       map[string]*EVMAdapter
	currentNetwork string
	mu             sync.RWMutex
}

// NewMultiChainManager 创建多链管理器
func NewMultiChainManager() (*MultiChainManager, error) {
	manager := &MultiChainManager{
		adapters: make(map[string]*EVMAdapter),
	}

	// 初始化所有启用的网络
	enabledNetworks := config.GetEnabledNetworks()
	if len(enabledNetworks) == 0 {
		return nil, fmt.Errorf("没有可用的网络配置")
	}

	for networkID, networkConfig := range enabledNetworks {
		adapter, err := NewEVMAdapter(networkConfig.RPCURL)
		if err != nil {
			// 记录错误但不终止，允许其他网络正常工作
			fmt.Printf("警告: 无法连接到网络 %s: %v\n", networkID, err)
			continue
		}
		manager.adapters[networkID] = adapter
	}

	if len(manager.adapters) == 0 {
		return nil, fmt.Errorf("无法连接到任何网络")
	}

	// 设置默认网络（优先选择以太坊主网，其次是第一个可用网络）
	if _, exists := manager.adapters["ethereum"]; exists {
		manager.currentNetwork = "ethereum"
	} else if _, exists := manager.adapters["sepolia"]; exists {
		manager.currentNetwork = "sepolia"
	} else {
		// 选择第一个可用网络
		for networkID := range manager.adapters {
			manager.currentNetwork = networkID
			break
		}
	}

	return manager, nil
}

// GetCurrentNetwork 获取当前网络ID
func (mcm *MultiChainManager) GetCurrentNetwork() string {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.currentNetwork
}

// SwitchNetwork 切换网络
func (mcm *MultiChainManager) SwitchNetwork(networkID string) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	if _, exists := mcm.adapters[networkID]; !exists {
		return fmt.Errorf("网络 %s 不存在或未连接", networkID)
	}

	mcm.currentNetwork = networkID
	return nil
}

// GetCurrentAdapter 获取当前网络的适配器
func (mcm *MultiChainManager) GetCurrentAdapter() (*EVMAdapter, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	adapter, exists := mcm.adapters[mcm.currentNetwork]
	if !exists {
		return nil, fmt.Errorf("当前网络 %s 的适配器不可用", mcm.currentNetwork)
	}

	return adapter, nil
}

// GetAdapter 获取指定网络的适配器
func (mcm *MultiChainManager) GetAdapter(networkID string) (*EVMAdapter, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	adapter, exists := mcm.adapters[networkID]
	if !exists {
		return nil, fmt.Errorf("网络 %s 的适配器不可用", networkID)
	}

	return adapter, nil
}

// GetAvailableNetworks 获取所有可用网络
func (mcm *MultiChainManager) GetAvailableNetworks() []string {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	networks := make([]string, 0, len(mcm.adapters))
	for networkID := range mcm.adapters {
		networks = append(networks, networkID)
	}

	return networks
}

// GetNetworkInfo 获取网络信息
func (mcm *MultiChainManager) GetNetworkInfo(networkID string) (*NetworkInfo, error) {
	// 获取配置
	networkConfig, err := config.GetNetwork(networkID)
	if err != nil {
		return nil, err
	}

	// 获取适配器
	adapter, err := mcm.GetAdapter(networkID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// 获取链ID和最新区块
	chainID, err := adapter.client.NetworkID(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取链ID失败: %w", err)
	}

	latestBlock, err := adapter.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块失败: %w", err)
	}

	// 获取gas建议
	gasSuggestion, err := adapter.GetGasSuggestion(ctx)
	if err != nil {
		gasSuggestion = &GasSuggestion{
			ChainID:  chainID,
			BaseFee:  big.NewInt(0),
			TipCap:   big.NewInt(0),
			MaxFee:   big.NewInt(0),
			GasPrice: big.NewInt(0),
		}
	}

	return &NetworkInfo{
		ID:            networkID,
		Name:          networkConfig.Name,
		ChainID:       chainID.Int64(),
		Symbol:        networkConfig.Symbol,
		Decimals:      networkConfig.Decimals,
		BlockExplorer: networkConfig.BlockExplorer,
		Testnet:       networkConfig.Testnet,
		LatestBlock:   latestBlock,
		GasSuggestion: gasSuggestion,
		Connected:     true,
	}, nil
}

// AddNetwork 添加新网络
func (mcm *MultiChainManager) AddNetwork(networkID string, rpcURL string) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	// 检查网络是否已存在
	if _, exists := mcm.adapters[networkID]; exists {
		return fmt.Errorf("网络 %s 已存在", networkID)
	}

	// 创建新的适配器
	adapter, err := NewEVMAdapter(rpcURL)
	if err != nil {
		return fmt.Errorf("创建网络适配器失败: %w", err)
	}

	mcm.adapters[networkID] = adapter
	return nil
}

// RemoveNetwork 移除网络
func (mcm *MultiChainManager) RemoveNetwork(networkID string) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	// 不能移除当前网络
	if mcm.currentNetwork == networkID {
		return fmt.Errorf("不能移除当前正在使用的网络")
	}

	delete(mcm.adapters, networkID)
	return nil
}

// CheckNetworkHealth 检查网络健康状态
func (mcm *MultiChainManager) CheckNetworkHealth(networkID string) error {
	adapter, err := mcm.GetAdapter(networkID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = adapter.client.BlockNumber(ctx)
	return err
}

// CheckAllNetworksHealth 检查所有网络健康状态
func (mcm *MultiChainManager) CheckAllNetworksHealth() map[string]error {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	health := make(map[string]error)
	for networkID := range mcm.adapters {
		health[networkID] = mcm.CheckNetworkHealth(networkID)
	}

	return health
}

// NetworkInfo 网络信息
type NetworkInfo struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	ChainID       int64          `json:"chain_id"`
	Symbol        string         `json:"symbol"`
	Decimals      int            `json:"decimals"`
	BlockExplorer string         `json:"block_explorer"`
	Testnet       bool           `json:"testnet"`
	LatestBlock   uint64         `json:"latest_block"`
	GasSuggestion *GasSuggestion `json:"gas_suggestion"`
	Connected     bool           `json:"connected"`
}

// CrossChainBalance 跨链余额查询
func (mcm *MultiChainManager) GetCrossChainBalance(address string, networks []string) (map[string]*big.Int, error) {
	balances := make(map[string]*big.Int)

	for _, networkID := range networks {
		adapter, err := mcm.GetAdapter(networkID)
		if err != nil {
			balances[networkID] = big.NewInt(0)
			continue
		}

		ctx := context.Background()
		balance, err := adapter.GetBalance(ctx, address)
		if err != nil {
			balances[networkID] = big.NewInt(0)
		} else {
			balances[networkID] = balance
		}
	}

	return balances, nil
}

// CrossChainTokenBalance 跨链代币余额查询
func (mcm *MultiChainManager) GetCrossChainTokenBalance(address, tokenAddress string, networks []string) (map[string]*big.Int, error) {
	balances := make(map[string]*big.Int)

	for _, networkID := range networks {
		adapter, err := mcm.GetAdapter(networkID)
		if err != nil {
			balances[networkID] = big.NewInt(0)
			continue
		}

		ctx := context.Background()
		balance, err := adapter.GetERC20Balance(ctx, tokenAddress, address)
		if err != nil {
			balances[networkID] = big.NewInt(0)
		} else {
			balances[networkID] = balance
		}
	}

	return balances, nil
}
