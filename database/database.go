/**
 * 数据库配置和连接管理
 * 
 * 这个包负责数据库的初始化、连接管理和迁移
 * 
 * 后端学习要点：
 * 1. GORM配置 - 数据库连接和配置
 * 2. 自动迁移 - 根据模型自动创建/更新表结构
 * 3. 连接池 - 管理数据库连接的生命周期
 * 4. 日志配置 - 数据库操作的日志记录
 * 5. 环境配置 - 不同环境使用不同的数据库设置
 */
package database

import (
	"fmt"
	"log"
	"os"
	"time"
	"wallet/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Driver   string `yaml:"driver" mapstructure:"driver"`     // postgres, mysql, sqlite
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	Database string `yaml:"database" mapstructure:"database"`
	SSLMode  string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	Timezone string `yaml:"timezone" mapstructure:"timezone"`
	
	// 连接池配置
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"`
	
	// 日志配置
	LogLevel logger.LogLevel `yaml:"log_level" mapstructure:"log_level"`
}

// 全局数据库实例
var DB *gorm.DB

/**
 * 初始化数据库连接
 * 
 * @param config 数据库配置
 * @return error 错误信息
 */
func InitDatabase(config DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector
	
	// 根据驱动类型选择方言
	switch config.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			config.Host, config.Username, config.Password, config.Database, 
			config.Port, config.SSLMode, config.Timezone)
		dialector = postgres.Open(dsn)
		
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database)
		dialector = mysql.Open(dsn)
		
	case "sqlite":
		dialector = sqlite.Open(config.Database)
		
	default:
		return fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
	
	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// 禁用外键约束检查（开发阶段）
		DisableForeignKeyConstraintWhenMigrating: true,
	}
	
	// 建立连接
	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// 配置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// 设置连接池参数
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}
	
	log.Printf("✅ Database connected successfully using %s driver", config.Driver)
	return nil
}

/**
 * 自动迁移数据库表结构
 * 根据模型定义自动创建/更新表
 */
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	
	log.Println("🔄 Starting database migration...")
	
	// 按依赖顺序迁移表
	err := DB.AutoMigrate(
		// 用户相关表
		&models.User{},
		&models.UserSession{},
		&models.UserPreference{},
		
		// 地址和钱包相关表
		&models.WatchAddress{},
		&models.UserWallet{},
		&models.AddressBalanceHistory{},
		
		// 日志表
		&models.ActivityLog{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	// 创建额外的索引（如果需要）
	if err := createAdditionalIndexes(); err != nil {
		log.Printf("⚠️ Warning: failed to create additional indexes: %v", err)
	}
	
	log.Println("✅ Database migration completed successfully")
	return nil
}

/**
 * 创建额外的索引
 * 根据查询需求创建复合索引
 */
func createAdditionalIndexes() error {
	// 用户表索引
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, is_active) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}
	
	// 观察地址复合索引
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_watch_addresses_user_network_active ON watch_addresses(user_id, network_id, is_favorite) WHERE deleted_at IS NULL").Error; err != nil {
		return err
	}
	
	// 会话表索引
	if err := DB.Exec("CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON user_sessions(user_id, is_active, expires_at)").Error; err != nil {
		return err
	}
	
	return nil
}

/**
 * 健康检查
 * 检查数据库连接状态
 */
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	return sqlDB.Ping()
}

/**
 * 关闭数据库连接
 * 优雅关闭数据库连接
 */
func CloseDatabase() error {
	if DB == nil {
		return nil
	}
	
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	
	log.Println("🔌 Closing database connection...")
	return sqlDB.Close()
}

/**
 * 获取默认配置
 * 提供开发环境的默认数据库配置
 */
func GetDefaultConfig() DatabaseConfig {
	// 从环境变量获取配置，如果没有则使用默认值
	driver := getEnv("DB_DRIVER", "sqlite")
	host := getEnv("DB_HOST", "localhost")
	database := getEnv("DB_NAME", "wallet.db")
	username := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	
	config := DatabaseConfig{
		Driver:   driver,
		Host:     host,
		Database: database,
		Username: username,
		Password: password,
		SSLMode:  "disable",
		Timezone: "UTC",
		
		// 连接池默认配置
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		
		// 开发环境显示详细日志
		LogLevel: logger.Info,
	}
	
	// 根据驱动设置默认端口
	switch driver {
	case "postgres":
		config.Port = 5432
		config.SSLMode = "disable"
	case "mysql":
		config.Port = 3306
	case "sqlite":
		// SQLite不需要端口
	}
	
	return config
}

/**
 * 辅助函数：获取环境变量
 */
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

/**
 * 开发辅助函数：重置数据库
 * 警告：这会删除所有数据！只用于开发环境
 */
func ResetDatabase() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	
	// 确保只在开发环境使用
	env := getEnv("GO_ENV", "development")
	if env == "production" {
		return fmt.Errorf("cannot reset database in production environment")
	}
	
	log.Println("⚠️ WARNING: Resetting database - all data will be lost!")
	
	// 删除所有表
	tables := []string{
		"address_balance_histories",
		"activity_logs", 
		"user_wallets",
		"watch_addresses",
		"user_preferences",
		"user_sessions",
		"users",
	}
	
	for _, table := range tables {
		if err := DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: failed to drop table %s: %v", table, err)
		}
	}
	
	// 重新运行迁移
	return AutoMigrate()
}package database
