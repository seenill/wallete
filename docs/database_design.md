# 钱包应用数据库设计

## 概述
本文档定义了钱包应用的数据库结构，包括用户管理、观察地址、钱包记录等核心功能的数据模型。

## 数据库选型
- **主数据库**: PostgreSQL (推荐) 或 MySQL
- **缓存**: Redis (用于会话管理和缓存)
- **连接方式**: GORM (Go语言ORM框架)

## 表结构设计

### 1. 用户表 (users)
存储用户基本信息和认证数据

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL -- 软删除
);

-- 索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_created_at ON users(created_at);
```

### 2. 用户会话表 (user_sessions)
管理用户登录会话和JWT令牌

```sql
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    refresh_token VARCHAR(255) UNIQUE NOT NULL,
    device_info JSONB, -- 设备信息
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
```

### 3. 观察地址表 (watch_addresses)
存储用户添加的观察地址

```sql
CREATE TABLE watch_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL, -- 以太坊地址
    label VARCHAR(100), -- 用户自定义标签
    network_id INTEGER DEFAULT 1, -- 网络ID (1=以太坊主网)
    address_type VARCHAR(20) DEFAULT 'EOA', -- EOA, Contract, MultiSig
    tags JSONB, -- 地址标签 ["DeFi", "NFT", "Exchange"]
    notes TEXT, -- 用户备注
    is_favorite BOOLEAN DEFAULT false,
    notification_enabled BOOLEAN DEFAULT true,
    balance_cache DECIMAL(36,18), -- 缓存的余额
    last_activity_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE(user_id, address, network_id) -- 同一用户同一网络下地址唯一
);

-- 索引
CREATE INDEX idx_watch_addresses_user_id ON watch_addresses(user_id);
CREATE INDEX idx_watch_addresses_address ON watch_addresses(address);
CREATE INDEX idx_watch_addresses_network ON watch_addresses(network_id);
CREATE INDEX idx_watch_addresses_user_network ON watch_addresses(user_id, network_id);
```

### 4. 用户钱包记录表 (user_wallets)
记录用户导入/创建的钱包(不存储私钥)

```sql
CREATE TABLE user_wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    address VARCHAR(42) NOT NULL,
    wallet_name VARCHAR(100) NOT NULL,
    wallet_type VARCHAR(20) DEFAULT 'imported', -- imported, created, hardware
    derivation_path VARCHAR(100), -- HD钱包派生路径
    network_id INTEGER DEFAULT 1,
    is_primary BOOLEAN DEFAULT false, -- 是否为主钱包
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE(user_id, address, network_id)
);

-- 索引
CREATE INDEX idx_user_wallets_user_id ON user_wallets(user_id);
CREATE INDEX idx_user_wallets_address ON user_wallets(address);
```

### 5. 地址余额历史表 (address_balance_history)
记录观察地址的余额变化历史

```sql
CREATE TABLE address_balance_history (
    id BIGSERIAL PRIMARY KEY,
    watch_address_id BIGINT NOT NULL REFERENCES watch_addresses(id) ON DELETE CASCADE,
    balance DECIMAL(36,18) NOT NULL,
    token_address VARCHAR(42), -- NULL表示主币(ETH)
    token_symbol VARCHAR(20),
    block_number BIGINT,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_balance_history_watch_id ON address_balance_history(watch_address_id);
CREATE INDEX idx_balance_history_recorded ON address_balance_history(recorded_at);
CREATE INDEX idx_balance_history_token ON address_balance_history(token_address);
```

### 6. 用户偏好设置表 (user_preferences)
存储用户的个性化设置

```sql
CREATE TABLE user_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    default_currency VARCHAR(10) DEFAULT 'USD',
    theme VARCHAR(20) DEFAULT 'light', -- light, dark, auto
    language VARCHAR(10) DEFAULT 'zh-CN',
    notifications JSONB DEFAULT '{}', -- 通知设置
    display_settings JSONB DEFAULT '{}', -- 显示设置
    privacy_settings JSONB DEFAULT '{}', -- 隐私设置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 7. 操作日志表 (user_activity_logs)
记录用户重要操作，用于安全审计

```sql
CREATE TABLE user_activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- login, logout, add_address, remove_address
    resource_type VARCHAR(50), -- user, wallet, address
    resource_id VARCHAR(100),
    details JSONB, -- 详细信息
    ip_address INET,
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success', -- success, failed, pending
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_activity_logs_user_id ON user_activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON user_activity_logs(action);
CREATE INDEX idx_activity_logs_created_at ON user_activity_logs(created_at);
```

## 数据关系说明

1. **users** ← **user_sessions**: 一对多，一个用户可以有多个活跃会话
2. **users** ← **watch_addresses**: 一对多，一个用户可以观察多个地址
3. **users** ← **user_wallets**: 一对多，一个用户可以有多个钱包
4. **watch_addresses** ← **address_balance_history**: 一对多，每个地址有历史余额记录
5. **users** ← **user_preferences**: 一对一，每个用户有一套偏好设置
6. **users** ← **user_activity_logs**: 一对多，记录用户所有操作

## 安全考虑

1. **密码安全**: 使用bcrypt加密，加盐存储
2. **会话管理**: JWT + 刷新令牌机制
3. **软删除**: 重要数据使用软删除，保留审计记录
4. **数据加密**: 敏感字段考虑加密存储
5. **访问控制**: 数据库层面的权限控制

## 性能优化

1. **索引策略**: 在查询频繁的字段上建立索引
2. **分区表**: 对于大量数据的日志表考虑分区
3. **缓存策略**: 热点数据使用Redis缓存
4. **读写分离**: 高并发时考虑主从架构