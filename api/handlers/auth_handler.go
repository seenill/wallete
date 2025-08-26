/*
认证API处理器包

本包实现了用户认证和授权相关的HTTP API处理器，集成数据库进行真实的用户管理。

主要功能：
用户管理：
- 用户注册和邮箱验证
- 用户登录验证和密码检查
- 用户资料获取和更新
- 用户状态管理

会话管理：
- JWT Token生成和验证
- 会话创建和销毁
- Token刷新和延期管理
- 多设备会话支持

API密钥管理：
- API密钥生成和命名
- 密钥有效期管理
- 密钥权限控制
- 密钥撤销和更新

安全特性：
- 密码bcrypt加密存储
- JWT Token签名和验证
- 会话超时管理
- IP地址和设备信息记录
- 用户活动日志记录

学习要点：
1. GORM数据库操作 - 用户查询、创建、更新
2. 密码安全 - bcrypt哈希和盐值
3. JWT认证 - 令牌生成和验证
4. 会话管理 - 数据库存储会话状态
5. 错误处理 - 统一的API错误响应
*/

package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wallet/api/middleware"
	"wallet/database"
	"wallet/models"
	"wallet/pkg/e"
	"wallet/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler 认证和授权相关的HTTP请求处理器
// 负责处理用户注册、登录、会话管理和API密钥生成等功能
type AuthHandler struct{}

// NewAuthHandler 创建新的认证处理器实例
// 返回: 初始化完成的AuthHandler指针
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// =============================================================================
// 请求和响应结构体定义
// =============================================================================

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    int64    `json:"expires_at"`
	User         UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username  *string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
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

// =============================================================================
// 用户注册和认证
// =============================================================================

// Register
// * 用户注册
// * 创建新用户账户，包括密码加密和数据验证
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 检查用户是否已存在
	var existingUser models.User
	result := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser)
	if result.Error == nil {
		// 用户已存在
		if existingUser.Email == req.Email {
			c.JSON(http.StatusConflict, gin.H{
				"code": e.ERROR,
				"msg":  "邮箱已被注册",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusConflict, gin.H{
				"code": e.ERROR,
				"msg":  "用户名已被使用",
				"data": nil,
			})
		}
		return
	}

	// 生成密码哈希
	salt := utils.GenerateSalt()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password+salt), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "密码加密失败",
			"data": nil,
		})
		return
	}

	// 创建新用户
	user := models.User{
		Username:     req.Username,
		Email:        strings.ToLower(req.Email),
		PasswordHash: string(passwordHash),
		Salt:         salt,
		IsActive:     true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建用户失败",
			"data": err.Error(),
		})
		return
	}

	// 记录用户注册活动
	h.logUserActivity(user.ID, "user_register", nil, nil, c)

	// 返回用户信息（不包含密码）
	userInfo := UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "注册成功",
		"data": userInfo,
	})
}

/**
 * 用户登录
 * 验证用户凭据并生成JWT令牌和会话
 */
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查找用户
	var user models.User
	result := database.DB.Where("email = ? AND is_active = ?", strings.ToLower(req.Email), true).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ERROR,
				"msg":  "邮箱或密码错误",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询用户失败",
				"data": nil,
			})
		}
		return
	}

	// 验证密码
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password+user.Salt))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "邮箱或密码错误",
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
	tokenExpiry := 24 * time.Hour

	token, err := authManager.GenerateJWT(strconv.Itoa(int(user.ID)), user.Username, tokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "生成Token失败",
			"data": err.Error(),
		})
		return
	}

	// 生成刷新令牌
	refreshToken := utils.GenerateRandomString(64)

	// 创建用户会话
	session := models.UserSession{
		UserID:       user.ID,
		SessionToken: token,
		RefreshToken: refreshToken,
		DeviceInfo:   h.getDeviceInfo(c),
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
		ExpiresAt:    time.Now().Add(tokenExpiry),
		IsActive:     true,
	}

	if err := database.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建会话失败",
			"data": nil,
		})
		return
	}

	// 更新用户最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	database.DB.Save(&user)

	// 记录登录活动
	h.logUserActivity(user.ID, "user_login", nil, nil, c)

	// 返回登录响应
	userInfo := UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "登录成功",
		"data": LoginResponse{
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(tokenExpiry).Unix(),
			User:         userInfo,
		},
	})
}

/**
 * 用户登出
 * 销毁当前会话
 */
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	// 获取当前会话令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  "缺少认证令牌",
			"data": nil,
		})
		return
	}

	// 提取token
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// 查找并关闭会话
	result := database.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND session_token = ? AND is_active = ?", userID, token, true).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "登出失败",
			"data": nil,
		})
		return
	}

	// 记录登出活动
	h.logUserActivity(userID.(uint), "user_logout", nil, nil, c)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "登出成功",
		"data": nil,
	})
}

/**
 * 刷新Token
 * 使用刷新令牌生成新的访问令牌
 */
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查找有效的会话
	var session models.UserSession
	result := database.DB.Where("refresh_token = ? AND is_active = ? AND expires_at > ?",
		req.RefreshToken, true, time.Now()).First(&session)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "无效的刷新令牌",
			"data": nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.First(&user, session.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取用户信息失败",
			"data": nil,
		})
		return
	}

	// 生成新的JWT token
	authManager := middleware.GetAuthManager()
	if authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "认证服务未初始化",
			"data": nil,
		})
		return
	}

	tokenExpiry := 24 * time.Hour
	newToken, err := authManager.GenerateJWT(strconv.Itoa(int(user.ID)), user.Username, tokenExpiry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "生成Token失败",
			"data": err.Error(),
		})
		return
	}

	// 更新会话
	session.SessionToken = newToken
	session.ExpiresAt = time.Now().Add(tokenExpiry)
	if err := database.DB.Save(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新会话失败",
			"data": nil,
		})
		return
	}

	// 记录活动
	h.logUserActivity(user.ID, "token_refresh", nil, nil, c)

	userInfo := UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "Token刷新成功",
		"data": LoginResponse{
			Token:        newToken,
			RefreshToken: req.RefreshToken,
			ExpiresAt:    time.Now().Add(tokenExpiry).Unix(),
			User:         userInfo,
		},
	})
}

/**
 * 获取用户信息
 * 返回当前登录用户的详细信息
 */
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "用户不存在",
			"data": nil,
		})
		return
	}

	userInfo := UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": userInfo,
	})
}

/**
 * 更新用户资料
 * 允许用户修改用户名和头像
 */
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 获取用户
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "用户不存在",
			"data": nil,
		})
		return
	}

	// 检查用户名是否已被使用
	if req.Username != nil && *req.Username != user.Username {
		var existingUser models.User
		result := database.DB.Where("username = ? AND id != ?", *req.Username, userID).First(&existingUser)
		if result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{
				"code": e.ERROR,
				"msg":  "用户名已被使用",
				"data": nil,
			})
			return
		}
		user.Username = *req.Username
	}

	// 更新头像
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	// 保存更改
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新用户信息失败",
			"data": nil,
		})
		return
	}

	// 记录活动
	h.logUserActivity(user.ID, "profile_update", nil, nil, c)

	userInfo := UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "更新成功",
		"data": userInfo,
	})
}

/**
 * 修改密码
 * 允许用户修改登录密码
 */
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 获取用户
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "用户不存在",
			"data": nil,
		})
		return
	}

	// 验证旧密码
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword+user.Salt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  "旧密码错误",
			"data": nil,
		})
		return
	}

	// 生成新密码哈希
	newSalt := utils.GenerateSalt()
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword+newSalt), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "密码加密失败",
			"data": nil,
		})
		return
	}

	// 更新密码
	user.PasswordHash = string(newPasswordHash)
	user.Salt = newSalt
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新密码失败",
			"data": nil,
		})
		return
	}

	// 关闭所有其他会话（密码修改后需要重新登录）
	database.DB.Model(&models.UserSession{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Update("is_active", false)

	// 记录活动
	h.logUserActivity(user.ID, "password_change", nil, nil, c)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "密码修改成功，请重新登录",
		"data": nil,
	})
}

/**
 * 生成API密钥
 * 为用户生成用于第三方集成的API密钥
 */
func (h *AuthHandler) GenerateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
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

	// 记录API密钥生成活动
	h.logUserActivity(userID.(uint), "api_key_generate", stringPtr("api_key"), &req.Name, c)

	expiresAt := time.Now().Add(expireDuration).Unix()

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "API密钥生成成功",
		"data": GenerateAPIKeyResponse{
			APIKey:    apiKey,
			Name:      req.Name,
			ExpiresAt: expiresAt,
		},
	})
}

// =============================================================================
// 辅助方法
// =============================================================================

/**
 * 获取设备信息
 * 从请求头中提取设备相关信息
 */
func (h *AuthHandler) getDeviceInfo(c *gin.Context) models.JSON {
	return models.JSON{
		"user_agent": c.GetHeader("User-Agent"),
		"platform":   h.extractPlatform(c.GetHeader("User-Agent")),
		"ip":         c.ClientIP(),
		"timestamp":  time.Now().Unix(),
	}
}

/**
 * 提取平台信息
 * 从 User-Agent 中提取平台信息
 */
func (h *AuthHandler) extractPlatform(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	switch {
	case strings.Contains(userAgent, "windows"):
		return "Windows"
	case strings.Contains(userAgent, "mac"):
		return "macOS"
	case strings.Contains(userAgent, "linux"):
		return "Linux"
	case strings.Contains(userAgent, "android"):
		return "Android"
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		return "iOS"
	default:
		return "Unknown"
	}
}

/**
 * 记录用户活动日志
 * 统一记录用户操作日志，用于安全审计
 */
func (h *AuthHandler) logUserActivity(userID uint, action string, resourceType, resourceID *string, c *gin.Context) {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	details := models.JSON{
		"timestamp": time.Now().Unix(),
		"endpoint":  c.FullPath(),
		"method":    c.Request.Method,
	}

	// 尝试记录用户活动日志
	activityLog := models.ActivityLog{
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
		Status:       "success",
	}

	err := database.DB.Create(&activityLog).Error
	if err != nil {
		// 日志记录失败不应该影响正常流程，只记录错误
		fmt.Printf("Failed to log user activity: %v\n", err)
	}
}

/**
 * 字符串指针辅助函数
 * 用于创建字符串指针
 */
func stringPtr(s string) *string {
	return &s
}
