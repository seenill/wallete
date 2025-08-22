package storage

import "wallet/models"

// WalletRepository 定义了钱包数据仓库需要实现的方法
type WalletRepository interface {
	Save(wallet *models.Wallet) error
	FindByAddress(address string) (*models.Wallet, error)
	ListAll() ([]models.Wallet, error)
}