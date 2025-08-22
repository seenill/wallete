package storage

import "wallet/models"

// postgresWalletRepository 是 WalletRepository 的PostgreSQL实现
type postgresWalletRepository struct{}

// NewPostgresWalletRepository 创建一个新的PostgreSQL钱包仓库实例
func NewPostgresWalletRepository() WalletRepository {
	return &postgresWalletRepository{}
}

// Save 将钱包元数据保存到数据库
func (r *postgresWalletRepository) Save(wallet *models.Wallet) error {
	return DB.Create(wallet).Error
}

// FindByAddress 根据地址查找钱包
func (r *postgresWalletRepository) FindByAddress(address string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := DB.Where("address = ?", address).First(&wallet).Error
	return &wallet, err
}

// ListAll 列出所有已保存的钱包
func (r *postgresWalletRepository) ListAll() ([]models.Wallet, error) {
	var wallets []models.Wallet
	err := DB.Find(&wallets).Error
	return wallets, err
}