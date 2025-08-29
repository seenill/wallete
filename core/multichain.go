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
	evmAdapters      map[string]*EVMAdapter
	solanaAdapters   map[string]*SolanaAdapter
	bitcoinAdapters  map[string]*BitcoinAdapter
	currentNetwork   string
	currentChainType string // "evm", "solana", "bitcoin"
	mu               sync.RWMutex
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
	ChainType     string         `json:"chain_type"` // 新增字段：链类型 (evm, solana, bitcoin)
}

// NewMultiChainManager 创建多链管理器
func NewMultiChainManager() (*MultiChainManager, error) {
	manager := &MultiChainManager{
		evmAdapters:     make(map[string]*EVMAdapter),
		solanaAdapters:  make(map[string]*SolanaAdapter),
		bitcoinAdapters: make(map[string]*BitcoinAdapter),
	}

	// 初始化所有启用的网络
	enabledNetworks := config.GetEnabledNetworks()
	if len(enabledNetworks) == 0 {
		return nil, fmt.Errorf("没有可用的网络配置")
	}

	for networkID, networkConfig := range enabledNetworks {
		switch networkID {
		case "solana", "solana_devnet":
			// 初始化Solana适配器
			adapter, err := NewSolanaAdapter(networkConfig.RPCURL)
			if err != nil {
				fmt.Printf("警告: 无法创建Solana网络适配器 %s: %v\n", networkID, err)
				continue
			}
			manager.solanaAdapters[networkID] = adapter
		case "bitcoin", "bitcoin_testnet":
			// 初始化Bitcoin适配器
			adapter, err := NewBitcoinAdapter(networkConfig.RPCURL)
			if err != nil {
				fmt.Printf("警告: 无法创建Bitcoin网络适配器 %s: %v\n", networkID, err)
				continue
			}
			manager.bitcoinAdapters[networkID] = adapter
		default:
			// 初始化EVM适配器
			adapter, err := NewEVMAdapter(networkConfig.RPCURL)
			if err != nil {
				// 记录错误但不终止，允许其他网络正常工作
				fmt.Printf("警告: 无法连接到网络 %s: %v\n", networkID, err)
				continue
			}
			manager.evmAdapters[networkID] = adapter
		}
	}

	// 如果没有适配器被成功初始化
	if len(manager.evmAdapters) == 0 && len(manager.solanaAdapters) == 0 && len(manager.bitcoinAdapters) == 0 {
		return nil, fmt.Errorf("无法连接到任何网络")
	}

	// 设置默认网络（优先选择以太坊主网，其次是第一个可用网络）
	if _, exists := manager.evmAdapters["ethereum"]; exists {
		manager.currentNetwork = "ethereum"
		manager.currentChainType = "evm"
	} else if _, exists := manager.evmAdapters["sepolia"]; exists {
		manager.currentNetwork = "sepolia"
		manager.currentChainType = "evm"
	} else if _, exists := manager.solanaAdapters["solana"]; exists {
		manager.currentNetwork = "solana"
		manager.currentChainType = "solana"
	} else if _, exists := manager.bitcoinAdapters["bitcoin"]; exists {
		manager.currentNetwork = "bitcoin"
		manager.currentChainType = "bitcoin"
	} else {
		// 选择第一个可用网络
		for networkID := range manager.evmAdapters {
			manager.currentNetwork = networkID
			manager.currentChainType = "evm"
			break
		}
		if manager.currentNetwork == "" {
			for networkID := range manager.solanaAdapters {
				manager.currentNetwork = networkID
				manager.currentChainType = "solana"
				break
			}
		}
		if manager.currentNetwork == "" {
			for networkID := range manager.bitcoinAdapters {
				manager.currentNetwork = networkID
				manager.currentChainType = "bitcoin"
				break
			}
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

// GetCurrentChainType 获取当前链类型
func (mcm *MultiChainManager) GetCurrentChainType() string {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()
	return mcm.currentChainType
}

// SwitchNetwork 切换网络
func (mcm *MultiChainManager) SwitchNetwork(networkID string) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	// 检查网络是否存在
	if _, exists := mcm.evmAdapters[networkID]; exists {
		mcm.currentNetwork = networkID
		mcm.currentChainType = "evm"
		return nil
	}

	if _, exists := mcm.solanaAdapters[networkID]; exists {
		mcm.currentNetwork = networkID
		mcm.currentChainType = "solana"
		return nil
	}

	if _, exists := mcm.bitcoinAdapters[networkID]; exists {
		mcm.currentNetwork = networkID
		mcm.currentChainType = "bitcoin"
		return nil
	}

	return fmt.Errorf("网络 %s 不存在或未连接", networkID)
}

// GetCurrentAdapter 获取当前网络的适配器
func (mcm *MultiChainManager) GetCurrentAdapter() (ChainAdapter, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	switch mcm.currentChainType {
	case "evm":
		adapter, exists := mcm.evmAdapters[mcm.currentNetwork]
		if !exists {
			return nil, fmt.Errorf("当前网络 %s 的EVM适配器不可用", mcm.currentNetwork)
		}
		return adapter, nil
	case "solana":
		adapter, exists := mcm.solanaAdapters[mcm.currentNetwork]
		if !exists {
			return nil, fmt.Errorf("当前网络 %s 的Solana适配器不可用", mcm.currentNetwork)
		}
		return adapter, nil
	case "bitcoin":
		adapter, exists := mcm.bitcoinAdapters[mcm.currentNetwork]
		if !exists {
			return nil, fmt.Errorf("当前网络 %s 的Bitcoin适配器不可用", mcm.currentNetwork)
		}
		return adapter, nil
	default:
		return nil, fmt.Errorf("不支持的链类型: %s", mcm.currentChainType)
	}
}

// GetAdapter 获取指定网络的适配器
func (mcm *MultiChainManager) GetAdapter(networkID string) (ChainAdapter, error) {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	// 检查EVM适配器
	if adapter, exists := mcm.evmAdapters[networkID]; exists {
		return adapter, nil
	}

	// 检查Solana适配器
	if adapter, exists := mcm.solanaAdapters[networkID]; exists {
		return adapter, nil
	}

	// 检查Bitcoin适配器
	if adapter, exists := mcm.bitcoinAdapters[networkID]; exists {
		return adapter, nil
	}

	return nil, fmt.Errorf("网络 %s 的适配器不可用", networkID)
}

// GetAvailableNetworks 获取所有可用网络
func (mcm *MultiChainManager) GetAvailableNetworks() []NetworkInfo {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	var networks []NetworkInfo

	// 添加EVM网络
	for networkID, adapter := range mcm.evmAdapters {
		networkConfig, err := config.GetNetwork(networkID)
		if err != nil {
			continue
		}

		ctx := context.Background()
		chainID, err := adapter.client.NetworkID(ctx)
		if err != nil {
			chainID = big.NewInt(networkConfig.ChainID)
		}

		latestBlock, err := adapter.client.BlockNumber(ctx)
		if err != nil {
			latestBlock = 0
		}

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

		networks = append(networks, NetworkInfo{
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
			ChainType:     "evm",
		})
	}

	// 添加Solana网络
	for networkID, adapter := range mcm.solanaAdapters {
		networkConfig, err := config.GetNetwork(networkID)
		if err != nil {
			continue
		}

		// 检查适配器是否为nil
		if adapter == nil {
			continue
		}

		networks = append(networks, NetworkInfo{
			ID:            networkID,
			Name:          networkConfig.Name,
			ChainID:       networkConfig.ChainID,
			Symbol:        networkConfig.Symbol,
			Decimals:      networkConfig.Decimals,
			BlockExplorer: networkConfig.BlockExplorer,
			Testnet:       networkConfig.Testnet,
			LatestBlock:   0, // TODO: 获取最新区块
			GasSuggestion: &GasSuggestion{
				ChainID:  big.NewInt(networkConfig.ChainID),
				BaseFee:  big.NewInt(0),
				TipCap:   big.NewInt(0),
				MaxFee:   big.NewInt(0),
				GasPrice: big.NewInt(0),
			},
			Connected: true,
			ChainType: "solana",
		})
	}

	// 添加Bitcoin网络
	for networkID, adapter := range mcm.bitcoinAdapters {
		networkConfig, err := config.GetNetwork(networkID)
		if err != nil {
			continue
		}

		// 检查适配器是否为nil
		if adapter == nil {
			continue
		}

		networks = append(networks, NetworkInfo{
			ID:            networkID,
			Name:          networkConfig.Name,
			ChainID:       networkConfig.ChainID,
			Symbol:        networkConfig.Symbol,
			Decimals:      networkConfig.Decimals,
			BlockExplorer: networkConfig.BlockExplorer,
			Testnet:       networkConfig.Testnet,
			LatestBlock:   0, // TODO: 获取最新区块
			GasSuggestion: &GasSuggestion{
				ChainID:  big.NewInt(networkConfig.ChainID),
				BaseFee:  big.NewInt(0),
				TipCap:   big.NewInt(0),
				MaxFee:   big.NewInt(0),
				GasPrice: big.NewInt(0),
			},
			Connected: true,
			ChainType: "bitcoin",
		})
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

	var chainID *big.Int
	var latestBlock uint64
	var gasSuggestion *GasSuggestion
	var chainType string

	// 根据适配器类型获取信息
	switch a := adapter.(type) {
	case *EVMAdapter:
		chainID, err = a.client.NetworkID(ctx)
		if err != nil {
			chainID = big.NewInt(networkConfig.ChainID)
		}

		latestBlock, err = a.client.BlockNumber(ctx)
		if err != nil {
			latestBlock = 0
		}

		gasSuggestion, err = a.GetGasSuggestion(ctx)
		if err != nil {
			gasSuggestion = &GasSuggestion{
				ChainID:  chainID,
				BaseFee:  big.NewInt(0),
				TipCap:   big.NewInt(0),
				MaxFee:   big.NewInt(0),
				GasPrice: big.NewInt(0),
			}
		}
		chainType = "evm"

	case *SolanaAdapter:
		chainID = big.NewInt(networkConfig.ChainID)
		latestBlock = 0 // TODO: 获取最新区块
		gasSuggestion = &GasSuggestion{
			ChainID:  chainID,
			BaseFee:  big.NewInt(0),
			TipCap:   big.NewInt(0),
			MaxFee:   big.NewInt(0),
			GasPrice: big.NewInt(0),
		}
		chainType = "solana"

	case *BitcoinAdapter:
		chainID = big.NewInt(networkConfig.ChainID)
		latestBlock = 0 // TODO: 获取最新区块
		gasSuggestion = &GasSuggestion{
			ChainID:  chainID,
			BaseFee:  big.NewInt(0),
			TipCap:   big.NewInt(0),
			MaxFee:   big.NewInt(0),
			GasPrice: big.NewInt(0),
		}
		chainType = "bitcoin"

	default:
		return nil, fmt.Errorf("不支持的适配器类型")
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
		ChainType:     chainType,
	}, nil
}

// AddNetwork 添加新网络
func (mcm *MultiChainManager) AddNetwork(networkID string, rpcURL string, chainType string) error {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	// 检查网络是否已存在
	if _, exists := mcm.evmAdapters[networkID]; exists {
		return fmt.Errorf("网络 %s 已存在", networkID)
	}
	if _, exists := mcm.solanaAdapters[networkID]; exists {
		return fmt.Errorf("网络 %s 已存在", networkID)
	}
	if _, exists := mcm.bitcoinAdapters[networkID]; exists {
		return fmt.Errorf("网络 %s 已存在", networkID)
	}

	// 创建新的适配器
	switch chainType {
	case "evm":
		adapter, err := NewEVMAdapter(rpcURL)
		if err != nil {
			return fmt.Errorf("创建EVM网络适配器失败: %w", err)
		}
		mcm.evmAdapters[networkID] = adapter
	case "solana":
		adapter, err := NewSolanaAdapter(rpcURL)
		if err != nil {
			return fmt.Errorf("创建Solana网络适配器失败: %w", err)
		}
		mcm.solanaAdapters[networkID] = adapter
	case "bitcoin":
		adapter, err := NewBitcoinAdapter(rpcURL)
		if err != nil {
			return fmt.Errorf("创建Bitcoin网络适配器失败: %w", err)
		}
		mcm.bitcoinAdapters[networkID] = adapter
	default:
		return fmt.Errorf("不支持的链类型: %s", chainType)
	}

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

	// 移除网络
	if _, exists := mcm.evmAdapters[networkID]; exists {
		delete(mcm.evmAdapters, networkID)
		return nil
	}

	if _, exists := mcm.solanaAdapters[networkID]; exists {
		delete(mcm.solanaAdapters, networkID)
		return nil
	}

	if _, exists := mcm.bitcoinAdapters[networkID]; exists {
		delete(mcm.bitcoinAdapters, networkID)
		return nil
	}

	return fmt.Errorf("网络 %s 不存在", networkID)
}

// CheckNetworkHealth 检查网络健康状态
func (mcm *MultiChainManager) CheckNetworkHealth(networkID string) error {
	adapter, err := mcm.GetAdapter(networkID)
	if err != nil {
		return err
	}

	// TODO: 实现不同链类型的健康检查
	ctx := context.Background()
	switch a := adapter.(type) {
	case *EVMAdapter:
		_, err = a.client.BlockNumber(ctx)
		return err
	case *SolanaAdapter:
		// TODO: 实现Solana健康检查
		return nil
	case *BitcoinAdapter:
		// TODO: 实现Bitcoin健康检查
		return nil
	default:
		return fmt.Errorf("不支持的适配器类型")
	}
}

// CheckAllNetworksHealth 检查所有网络健康状态
func (mcm *MultiChainManager) CheckAllNetworksHealth() map[string]error {
	mcm.mu.RLock()
	defer mcm.mu.RUnlock()

	health := make(map[string]error)

	// 检查EVM网络
	for networkID := range mcm.evmAdapters {
		health[networkID] = mcm.CheckNetworkHealth(networkID)
	}

	// 检查Solana网络
	for networkID := range mcm.solanaAdapters {
		health[networkID] = mcm.CheckNetworkHealth(networkID)
	}

	// 检查Bitcoin网络
	for networkID := range mcm.bitcoinAdapters {
		health[networkID] = mcm.CheckNetworkHealth(networkID)
	}

	return health
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
