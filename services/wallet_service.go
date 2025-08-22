package services

import (
	"fmt"
	"wallet/config"
	"wallet/core"
	"wallet/models"
	"wallet/storage"
)

// WalletService 封装了钱包相关的业务逻辑
type WalletService struct {
	repo storage.WalletRepository
}

// NewWalletService 创建一个新的 WalletService 实例
func NewWalletService(repo storage.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

// CreateWallet 创建一个新钱包
// name: 钱包的自定义名称
// password: 用于加密钱包的密码
// 返回: 创建的钱包模型和可能发生的错误
func (s *WalletService) CreateWallet(name, password string) (*models.Wallet, error) {
	// 1. 调用核心层创建加密的keystore
	keystoreDir := config.AppConfig.Keystore.Path
	account, err := core.CreateKeystore(password, keystoreDir)
	if err != nil {
		return nil, fmt.Errorf("核心层创建keystore失败: %w", err)
	}

	// 2. 准备要存入数据库的模型
	wallet := &models.Wallet{
		Address:      account.Address.Hex(),
		Name:         name,
		KeystorePath: account.URL.Path, // account.URL.Path 包含了keystore文件的完整路径
	}

	// 3. 调用存储层将钱包元数据保存到数据库
	if err := s.repo.Save(wallet); err != nil {
		return nil, fmt.Errorf("存储钱包元数据失败: %w", err)
	}

	fmt.Printf("钱包元数据已成功存入数据库, 地址: %s\n", wallet.Address)
	return wallet, nil
}