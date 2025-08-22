package models

import "gorm.io/gorm"

// Wallet 代表数据库中的钱包记录
type Wallet struct {
	gorm.Model
	Address      string `gorm:"uniqueIndex;not null"` // 钱包地址
	Name         string // 钱包名称
	KeystorePath string `gorm:"not null"`           // 指向加密keystore文件的路径
}