/*
认证API处理器包

本包实现了用户认证和授权相关的HTTP API处理器。

主要功能：
认证管理：
- 用户登录验证和JWT Token生成
- Token刷新和延期管理
- 用户会话状态管理
- 用户资料获取和更新

API密钥管理：
- API密钥生成和命名
- 密钥有效期管理
- 密钥权限控制
- 密钥撤销和更新

安全特性：
- 密码安全验证
- JWT Token签名和验证
- 会话超时管理
- 防暴力破解机制

注意事项：
- 当前实现为简化版本，仅作为演示
- 生产环境应该集成正式的用户管理系统
- 密码应该使用安全的哈希算法存储
*/
package handlers

import (
	"net/http"
	"time"
	"wallet/api/middleware"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证和授权相关的HTTP请求处理器
// 负责处理用户登录、Token管理和API密钥生成等功能
type AuthHandler struct{}

// NewAuthHandler 创建新的认证处理器实例
// 返回: 初始化完成的AuthHandler指针
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
}

// GenerateAPIKeyRequest 生成API密钥请求
type GenerateAPIKeyRequest struct {
	Name       string `json:"name" binding:"required"`
	ExpireDays int    `json:"expire_days"` // 过期天数，默认30天
}

// GenerateAPIKeyResponse 生成API密钥响应
type GenerateAPIKeyResponse struct {
	APIKey    string `json:"api_key"`
	Name      string `json:"name"`
	ExpiresAt int64  `json:"expires_at"`
}

// Login 用户登录（简化版本，实际生产环境需要对接用户系统）
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	// 简化的用户验证（实际应该从数据库验证）
	// 这里仅作为演示，使用固定的用户名密码
	if req.Username != "admin" || req.Password != "wallet123" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户名或密码错误",
			"data": nil,
		})
		return
	}

	// 生成JWT token
	authManager := middleware.GetAuthManager()
	if authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "认证服务未初始化",
			"data": nil,
		})
		return
	}

	// Token有效期24小时
	expireDuration := 24 * time.Hour
	token, err := authManager.GenerateJWT("user_001", req.Username, expireDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "生成Token失败",
			"data": err.Error(),
		})
		return
	}

	expiresAt := time.Now().Add(expireDuration).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			UserID:    "user_001",
			Username:  req.Username,
		},
	})
}

// GenerateAPIKey 生成API密钥
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": err.Error(),
		})
		return
	}

	if req.ExpireDays <= 0 {
		req.ExpireDays = 30 // 默认30天
	}

	authManager := middleware.GetAuthManager()
	if authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "认证服务未初始化",
			"data": nil,
		})
		return
	}

	expireDuration := time.Duration(req.ExpireDays) * 24 * time.Hour
	apiKey, err := authManager.GenerateAPIKey(req.Name, expireDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "生成API密钥失败",
			"data": err.Error(),
		})
		return
	}

	expiresAt := time.Now().Add(expireDuration).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": GenerateAPIKeyResponse{
			APIKey:    apiKey,
			Name:      req.Name,
			ExpiresAt: expiresAt,
		},
	})
}

// RefreshToken 刷新Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	username, _ := c.Get("username")

	authManager := middleware.GetAuthManager()
	if authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "认证服务未初始化",
			"data": nil,
		})
		return
	}

	// 生成新的Token
	expireDuration := 24 * time.Hour
	token, err := authManager.GenerateJWT(userID.(string), username.(string), expireDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "刷新Token失败",
			"data": err.Error(),
		})
		return
	}

	expiresAt := time.Now().Add(expireDuration).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": LoginResponse{
			Token:     token,
			ExpiresAt: expiresAt,
			UserID:    userID.(string),
			Username:  username.(string),
		},
	})
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": gin.H{
			"user_id":  userID,
			"username": username,
		},
	})
}
