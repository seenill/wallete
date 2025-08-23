/*
配置管理包

本包负责加载和管理应用程序的所有配置信息，包括：
- 服务器配置（端口等）
- 数据库配置（连接参数）
- 多链网络配置（RPC地址、链 ID等）
- 安全配置（JWT密钥、加密密钥等）
- Keystore配置（文件路径）

配置文件格式为YAML，支持环境变量覆盖。
*/
package config

import (
	"fmt"
	"math/big"

	"github.com/spf13/viper"
)

// Config 主配置结构体，映射整个配置文件的内容
// 包含服务器、数据库、网络、安全和Keystore配置
type Config struct {
	Server   ServerConfig             // 服务器配置
	Database DatabaseConfig           // 数据库配置
	Networks map[string]NetworkConfig `mapstructure:"networks"` // 网络配置映射
	Security SecurityConfig           // 安全配置
	Keystore KeystoreConfig           // 密钥库配置
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port int // 服务器监听端口
}

// DatabaseConfig 数据库连接配置
// 目前支持PostgreSQL数据库
type DatabaseConfig struct {
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名
	SSLMode  string // SSL连接模式（disable/require/verify-ca/verify-full）
}

// NetworkConfig 区块链网络配置
// 支持多个区块链网络，包括以太坊主网、测试网、Polygon、BSC等
type NetworkConfig struct {
	Name             string `mapstructure:"name"`              // 网络显示名称
	RPCURL           string `mapstructure:"rpc_url"`           // RPC节点地址
	ChainID          int64  `mapstructure:"chain_id"`          // 区块链链 ID（EIP-155）
	Symbol           string `mapstructure:"symbol"`            // 网络原生代币符号（如ETH、MATIC等）
	Decimals         int    `mapstructure:"decimals"`          // 网络原生代币小数位数（通常为18）
	BlockExplorer    string `mapstructure:"block_explorer"`    // 区块浏览器地址（如Etherscan）
	Enabled          bool   `mapstructure:"enabled"`           // 是否启用该网络
	Testnet          bool   `mapstructure:"testnet"`           // 是否为测试网络
	MaxGasPrice      string `mapstructure:"max_gas_price"`     // 最大gas价格限制（wei单位）
	MinConfirmations int    `mapstructure:"min_confirmations"` // 交易最小确认数
}

// SecurityConfig 安全相关配置
// 包括JWT认证、数据加密和速率限制配置
type SecurityConfig struct {
	JWTSecret     string          `mapstructure:"jwt_secret"`     // JWT签名密钥（生产环境应使用强密码）
	EncryptionKey string          `mapstructure:"encryption_key"` // 数据加密密钥（用于助记词加密）
	RateLimit     RateLimitConfig `mapstructure:"rate_limit"`     // 速率限制配置
}

// RateLimitConfig API速率限制配置
// 为不同API类型设置不同的限制策略
type RateLimitConfig struct {
	General     int `mapstructure:"general"`     // 通用API限制（每分钟请求数）
	Transaction int `mapstructure:"transaction"` // 交易API限制（每分钟请求数）
	Auth        int `mapstructure:"auth"`        // 认证API限制（每分钟请求数）
}

// KeystoreConfig 密钥库存储配置
type KeystoreConfig struct {
	Path string // 密钥库文件存储路径
}

// AppConfig 全局配置实例
// 应用启动时加载，全局可访问
var AppConfig Config

// LoadConfig 加载配置文件
// 从./config/config.yaml文件中加载配置，支持环境变量覆盖
// 加载成功后会验证配置的合法性并设置默认值
func LoadConfig() {
	// 设置配置文件名称和类型
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config") // 指定配置文件路径

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// 将配置解析到结构体中
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %w", err))
	}

	// 验证配置的合法性
	validateConfig()
}

// validateConfig 验证配置的合法性并设置默认值
// 检查必要的配置项是否存在，为缺失的配置设置合理默认值
func validateConfig() {
	// 检查是否至少配置了一个网络
	if len(AppConfig.Networks) == 0 {
		panic("至少需要配置一个网络")
	}

	// 逐个验证网络配置的完整性
	for name, network := range AppConfig.Networks {
		if network.RPCURL == "" {
			panic(fmt.Sprintf("网络 %s 的 RPC URL 不能为空", name))
		}
		if network.ChainID <= 0 {
			panic(fmt.Sprintf("网络 %s 的 Chain ID 必须大于 0", name))
		}
		if network.Symbol == "" {
			panic(fmt.Sprintf("网络 %s 的 Symbol 不能为空", name))
		}
	}

	// 为速率限制设置默认值
	if AppConfig.Security.RateLimit.General == 0 {
		AppConfig.Security.RateLimit.General = 100 // 默认每分钟100次请求
	}
	if AppConfig.Security.RateLimit.Transaction == 0 {
		AppConfig.Security.RateLimit.Transaction = 10 // 默认每分钟10次交易
	}
	if AppConfig.Security.RateLimit.Auth == 0 {
		AppConfig.Security.RateLimit.Auth = 5 // 默认每分钟5次认证请求
	}
}

// GetNetwork 获取指定网络的配置信息
// 参数: networkID - 网络标识符（如"ethereum"、"polygon"等）
// 返回: 网络配置指针和错误信息
// 注意: 只返回已启用的网络
func GetNetwork(networkID string) (*NetworkConfig, error) {
	network, exists := AppConfig.Networks[networkID]
	if !exists {
		return nil, fmt.Errorf("网络 %s 不存在", networkID)
	}
	if !network.Enabled {
		return nil, fmt.Errorf("网络 %s 未启用", networkID)
	}
	return &network, nil
}

// GetEnabledNetworks 获取所有已启用的网络配置
// 返回: 网络标识符到配置的映射
// 用于显示可用网络列表或网络切换
func GetEnabledNetworks() map[string]NetworkConfig {
	enabledNetworks := make(map[string]NetworkConfig)
	for name, network := range AppConfig.Networks {
		if network.Enabled {
			enabledNetworks[name] = network
		}
	}
	return enabledNetworks
}

// GetMainnetNetworks 获取所有已启用的主网络配置
// 返回: 主网络标识符到配置的映射
// 用于区分主网和测试网，主网交易需要更高的安全级别
func GetMainnetNetworks() map[string]NetworkConfig {
	mainnetNetworks := make(map[string]NetworkConfig)
	for name, network := range AppConfig.Networks {
		if network.Enabled && !network.Testnet {
			mainnetNetworks[name] = network
		}
	}
	return mainnetNetworks
}

// GetTestnetNetworks 获取所有已启用的测试网络配置
// 返回: 测试网络标识符到配置的映射
// 用于开发和测试环境，测试网的代币没有价值
func GetTestnetNetworks() map[string]NetworkConfig {
	testnetNetworks := make(map[string]NetworkConfig)
	for name, network := range AppConfig.Networks {
		if network.Enabled && network.Testnet {
			testnetNetworks[name] = network
		}
	}
	return testnetNetworks
}

// GetMaxGasPrice 获取网络的最大gas价格限制
// 返回: *big.Int 类型的gas价格（wei单位），如果未设置则返回nil
// 用于防止gas价格过高导致的意外损失
func (nc *NetworkConfig) GetMaxGasPrice() *big.Int {
	if nc.MaxGasPrice == "" {
		return nil // 未设置限制
	}
	maxGas := new(big.Int)
	maxGas.SetString(nc.MaxGasPrice, 10) // 十进制字符串转换为big.Int
	return maxGas
}
