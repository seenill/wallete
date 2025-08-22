package config

import (
	"fmt"
	"github.com/spf13/viper"
)

// Config 结构体用于映射整个配置文件
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Ethereum EthereumConfig
	Keystore KeystoreConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type EthereumConfig struct {
	RPCURL string `mapstructure:"rpc_url"`
}

type KeystoreConfig struct {
	Path string
}

var AppConfig Config

// LoadConfig 加载配置文件
func LoadConfig() {
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
}
