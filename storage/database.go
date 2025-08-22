package storage

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"wallet/config"
	"wallet/models"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.AppConfig.Database.Host,
		config.AppConfig.Database.User,
		config.AppConfig.Database.Password,
		config.AppConfig.Database.DBName,
		config.AppConfig.Database.Port,
		config.AppConfig.Database.SSLMode,
	)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 自动迁移数据库模型
	err = DB.AutoMigrate(&models.Wallet{})
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}