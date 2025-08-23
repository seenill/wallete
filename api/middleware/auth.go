/*
认证中间件包

本包实现了钱包服务的认证和授权中间件，提供多层次的安全防护。

认证机制：
JWT Token认证：
- 基于JWT标准的无状态认证
- 支持用户信息和权限台纳
- 自动过期检查和更新
- 支持Token刷新和撤销

API密钥认证：
- 长期有效的API访问密钥
- 支持密钥命名和管理
- 灵活的过期时间设置
- 密钥启用/禁用控制

安全特性：
- HMAC-SHA256签名算法
- 随机生成的安全密钥
- 多种认证方式兼容
- 请求来源验证和记录

中间件类型：
- JWTAuth: 强制JWT认证
- APIKeyAuth: 强制API密钥认证
- OptionalAuth: 可选认证（支持多种方式）
*/
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT令牌声明结构体
// 包含用户身份信息和JWT标准声明字段
type JWTClaims struct {
	UserID               string `json:"user_id"`  // 用户唯一标识符
	Username             string `json:"username"` // 用户名
	jwt.RegisteredClaims        // JWT标准声明（过期时间、发布者等）
}

// AuthManager 认证管理器
// 统一管理JWT和API密钥的生成、验证和存储
type AuthManager struct {
	jwtSecret []byte                // JWT签名密钥（HMAC-SHA256）
	apiKeys   map[string]APIKeyInfo // API密钥存储映射（内存存储）
}

// APIKeyInfo API密钥详细信息
// 用于管理和验证API密钥的有效性和权限
type APIKeyInfo struct {
	Name      string    `json:"name"`       // 密钥显示名称（用于标识和管理）
	CreatedAt time.Time `json:"created_at"` // 密钥创建时间
	ExpiresAt time.Time `json:"expires_at"` // 密钥过期时间
	Enabled   bool      `json:"enabled"`    // 密钥是否启用（支持禁用操作）
}

// authManager 全局认证管理器实例
// 在应用启动时初始化，为所有认证中间件提供服务
var authManager *AuthManager

// InitAuth 初始化认证管理器
func InitAuth(jwtSecret string) {
	secret := []byte(jwtSecret)
	if len(secret) == 0 {
		// 生成随机密钥
		secret = make([]byte, 32)
		rand.Read(secret)
	}

	authManager = &AuthManager{
		jwtSecret: secret,
		apiKeys:   make(map[string]APIKeyInfo),
	}
}

// GenerateJWT 生成JWT token
func (am *AuthManager) GenerateJWT(userID, username string, expireDuration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.jwtSecret)
}

// ValidateJWT 验证JWT token
func (am *AuthManager) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateAPIKey 生成API密钥
func (am *AuthManager) GenerateAPIKey(name string, expireDuration time.Duration) (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}

	apiKey := "wapi_" + hex.EncodeToString(keyBytes)

	am.apiKeys[apiKey] = APIKeyInfo{
		Name:      name,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expireDuration),
		Enabled:   true,
	}

	return apiKey, nil
}

// ValidateAPIKey 验证API密钥
func (am *AuthManager) ValidateAPIKey(apiKey string) (bool, string) {
	info, exists := am.apiKeys[apiKey]
	if !exists {
		return false, "API key not found"
	}

	if !info.Enabled {
		return false, "API key disabled"
	}

	if time.Now().After(info.ExpiresAt) {
		return false, "API key expired"
	}

	return true, ""
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ErrorAuth,
				"msg":  "Authentication not initialized",
				"data": nil,
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ErrorAuth,
				"msg":  "Authorization header required",
				"data": nil,
			})
			c.Abort()
			return
		}

		// 检查Bearer token
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := authManager.ValidateJWT(tokenString)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": e.ErrorAuth,
					"msg":  "Invalid JWT token",
					"data": err.Error(),
				})
				c.Abort()
				return
			}

			// 将用户信息添加到上下文
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ErrorAuth,
				"msg":  "Invalid authorization format",
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyAuth API密钥认证中间件
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authManager == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ErrorAuth,
				"msg":  "Authentication not initialized",
				"data": nil,
			})
			c.Abort()
			return
		}

		// 从header或query参数中获取API key
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ErrorAuth,
				"msg":  "API key required",
				"data": nil,
			})
			c.Abort()
			return
		}

		valid, errMsg := authManager.ValidateAPIKey(apiKey)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ErrorAuth,
				"msg":  errMsg,
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth 可选认证中间件（允许未认证的访问，但会设置认证状态）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authManager == nil {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		authenticated := false

		// 尝试JWT认证
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := authManager.ValidateJWT(tokenString); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("username", claims.Username)
				authenticated = true
			}
		}

		// 尝试API Key认证
		if !authenticated && apiKey != "" {
			if valid, _ := authManager.ValidateAPIKey(apiKey); valid {
				authenticated = true
			}
		}

		c.Set("authenticated", authenticated)
		c.Next()
	}
}

// GetAuthManager 获取认证管理器实例
func GetAuthManager() *AuthManager {
	return authManager
}
