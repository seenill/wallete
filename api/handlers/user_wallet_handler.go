/**
 * 用户钱包记录API处理器
 *
 * 本包实现了用户钱包记录管理相关的HTTP API处理器，集成数据库进行钱包记录管理。
 *
 * 主要功能：
 * 钱包记录管理：
 * - 记录用户导入/创建的钱包地址
 * - 管理钱包名称和类型信息
 * - 设置主钱包和使用记录
 * - 支持多网络钱包管理
 *
 * 钱包类型：
 * - imported: 导入的钱包（助记词/私钥）
 * - created: 创建的钱包
 * - hardware: 硬件钱包
 *
 * 安全考虑：
 * - 不存储私钥或助记词
 * - 只记录地址和元数据
 * - 用户权限验证
 *
 * 学习要点：
 * 1. 钱包管理 - 用户钱包的元数据管理
 * 2. 数据分离 - 敏感数据与元数据分离
 * 3. 多网络支持 - 不同区块链网络的钱包
 * 4. 排序和分页 - 钱包列表的管理
 * 5. 状态管理 - 主钱包和使用状态
 */
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wallet/database"
	"wallet/models"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserWalletHandler 用户钱包记录相关的HTTP请求处理器
// 负责处理用户钱包记录的增删改查等功能
type UserWalletHandler struct{}

// NewUserWalletHandler 创建新的用户钱包处理器实例
// 返回: 初始化完成的UserWalletHandler指针
func NewUserWalletHandler() *UserWalletHandler {
	return &UserWalletHandler{}
}

// =============================================================================
// 请求和响应结构体定义
// =============================================================================

// AddUserWalletRequest 添加用户钱包请求
type AddUserWalletRequest struct {
	Address        string  `json:"address" binding:"required"`
	WalletName     string  `json:"wallet_name" binding:"required"`
	WalletType     string  `json:"wallet_type" binding:"required,oneof=imported created hardware"`
	DerivationPath *string `json:"derivation_path,omitempty"`
	NetworkID      *int    `json:"network_id,omitempty"`
	IsPrimary      *bool   `json:"is_primary,omitempty"`
}

// UpdateUserWalletRequest 更新用户钱包请求
type UpdateUserWalletRequest struct {
	WalletName *string `json:"wallet_name,omitempty"`
	IsPrimary  *bool   `json:"is_primary,omitempty"`
}

// UserWalletResponse 用户钱包响应
type UserWalletResponse struct {
	ID             uint       `json:"id"`
	Address        string     `json:"address"`
	WalletName     string     `json:"wallet_name"`
	WalletType     string     `json:"wallet_type"`
	DerivationPath *string    `json:"derivation_path,omitempty"`
	NetworkID      int        `json:"network_id"`
	IsPrimary      bool       `json:"is_primary"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// UserWalletListResponse 用户钱包列表响应
type UserWalletListResponse struct {
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
	Wallets  []UserWalletResponse `json:"wallets"`
}

// =============================================================================
// 用户钱包管理
// =============================================================================

/**
 * 添加用户钱包记录
 * 记录用户导入或创建的钱包地址和元数据
 */
func (h *UserWalletHandler) AddUserWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	var req AddUserWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证以太坊地址格式
	if !h.isValidEthereumAddress(req.Address) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "无效的以太坊地址格式",
			"data": nil,
		})
		return
	}

	// 标准化地址格式
	address := strings.ToLower(req.Address)

	// 设置默认值
	networkID := 1 // 默认以太坊主网
	if req.NetworkID != nil {
		networkID = *req.NetworkID
	}

	// 检查是否已存在相同的钱包记录
	var existing models.UserWallet
	result := database.DB.Where("user_id = ? AND address = ? AND network_id = ?",
		userID, address, networkID).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code": e.ERROR,
			"msg":  "该钱包地址已存在",
			"data": nil,
		})
		return
	}

	// 检查是否要设置为主钱包
	isPrimary := false
	if req.IsPrimary != nil {
		isPrimary = *req.IsPrimary
	}

	// 如果设置为主钱包，需要先取消其他主钱包
	if isPrimary {
		database.DB.Model(&models.UserWallet{}).
			Where("user_id = ? AND network_id = ? AND is_primary = ?", userID, networkID, true).
			Update("is_primary", false)
	}

	// 创建钱包记录
	userWallet := models.UserWallet{
		UserID:         userID.(uint),
		Address:        address,
		WalletName:     req.WalletName,
		WalletType:     req.WalletType,
		DerivationPath: req.DerivationPath,
		NetworkID:      networkID,
		IsPrimary:      isPrimary,
	}

	if err := database.DB.Create(&userWallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "添加钱包记录失败",
			"data": err.Error(),
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "user_wallet_add", "user_wallet",
		fmt.Sprintf("%d", userWallet.ID), c)

	// 返回创建的钱包信息
	response := h.convertToUserWalletResponse(userWallet)

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "钱包记录添加成功",
		"data": response,
	})
}

/**
 * 获取用户钱包列表
 * 支持分页、排序和筛选
 */
func (h *UserWalletHandler) GetUserWallets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	networkID := c.Query("network_id")
	walletType := c.Query("wallet_type")
	primaryOnly := c.Query("primary_only")
	sortBy := c.DefaultQuery("sort_by", "is_primary")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 限制分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := database.DB.Model(&models.UserWallet{}).Where("user_id = ?", userID)

	// 网络过滤
	if networkID != "" {
		if netID, err := strconv.Atoi(networkID); err == nil {
			query = query.Where("network_id = ?", netID)
		}
	}

	// 钱包类型过滤
	if walletType != "" {
		query = query.Where("wallet_type = ?", walletType)
	}

	// 主钱包过滤
	if primaryOnly == "true" {
		query = query.Where("is_primary = ?", true)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "查询钱包记录失败",
			"data": nil,
		})
		return
	}

	// 排序
	orderClause := h.buildOrderClause(sortBy, sortOrder)
	query = query.Order(orderClause)

	// 分页查询
	var wallets []models.UserWallet
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&wallets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "查询钱包记录失败",
			"data": nil,
		})
		return
	}

	// 转换响应格式
	responseWallets := make([]UserWalletResponse, len(wallets))
	for i, wallet := range wallets {
		responseWallets[i] = h.convertToUserWalletResponse(wallet)
	}

	response := UserWalletListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Wallets:  responseWallets,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

/**
 * 获取单个钱包记录详情
 */
func (h *UserWalletHandler) GetUserWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	walletID := c.Param("id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少钱包ID",
			"data": nil,
		})
		return
	}

	// 查询钱包记录
	var userWallet models.UserWallet
	result := database.DB.Where("id = ? AND user_id = ?", walletID, userID).First(&userWallet)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": e.ERROR,
				"msg":  "钱包记录不存在",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询钱包记录失败",
				"data": nil,
			})
		}
		return
	}

	// 更新最后使用时间
	now := time.Now()
	userWallet.LastUsedAt = &now
	database.DB.Save(&userWallet)

	// 转换响应格式
	response := h.convertToUserWalletResponse(userWallet)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

/**
 * 更新钱包记录
 * 允许更新钱包名称和主钱包状态
 */
func (h *UserWalletHandler) UpdateUserWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	walletID := c.Param("id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少钱包ID",
			"data": nil,
		})
		return
	}

	var req UpdateUserWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询钱包记录
	var userWallet models.UserWallet
	result := database.DB.Where("id = ? AND user_id = ?", walletID, userID).First(&userWallet)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": e.ERROR,
				"msg":  "钱包记录不存在",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询钱包记录失败",
				"data": nil,
			})
		}
		return
	}

	// 更新字段
	if req.WalletName != nil {
		userWallet.WalletName = *req.WalletName
	}

	// 处理主钱包设置
	if req.IsPrimary != nil {
		if *req.IsPrimary && !userWallet.IsPrimary {
			// 设置为主钱包，需要先取消同网络下其他主钱包
			database.DB.Model(&models.UserWallet{}).
				Where("user_id = ? AND network_id = ? AND is_primary = ? AND id != ?",
					userID, userWallet.NetworkID, true, userWallet.ID).
				Update("is_primary", false)
		}
		userWallet.IsPrimary = *req.IsPrimary
	}

	// 保存更改
	if err := database.DB.Save(&userWallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新钱包记录失败",
			"data": nil,
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "user_wallet_update", "user_wallet", walletID, c)

	// 返回更新后的钱包信息
	response := h.convertToUserWalletResponse(userWallet)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "钱包记录更新成功",
		"data": response,
	})
}

/**
 * 删除钱包记录
 * 软删除钱包记录
 */
func (h *UserWalletHandler) DeleteUserWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	walletID := c.Param("id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少钱包ID",
			"data": nil,
		})
		return
	}

	// 查询并删除钱包记录
	result := database.DB.Where("id = ? AND user_id = ?", walletID, userID).
		Delete(&models.UserWallet{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "删除钱包记录失败",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "钱包记录不存在",
			"data": nil,
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "user_wallet_delete", "user_wallet", walletID, c)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "钱包记录删除成功",
		"data": nil,
	})
}

/**
 * 设置主钱包
 * 将指定钱包设置为主钱包
 */
func (h *UserWalletHandler) SetPrimaryWallet(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	walletID := c.Param("id")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少钱包ID",
			"data": nil,
		})
		return
	}

	// 查询钱包记录
	var userWallet models.UserWallet
	result := database.DB.Where("id = ? AND user_id = ?", walletID, userID).First(&userWallet)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": e.ERROR,
				"msg":  "钱包记录不存在",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询钱包记录失败",
				"data": nil,
			})
		}
		return
	}

	// 开始事务
	tx := database.DB.Begin()

	// 取消同网络下其他主钱包
	if err := tx.Model(&models.UserWallet{}).
		Where("user_id = ? AND network_id = ? AND is_primary = ?",
			userID, userWallet.NetworkID, true).
		Update("is_primary", false).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "设置主钱包失败",
			"data": nil,
		})
		return
	}

	// 设置新的主钱包
	if err := tx.Model(&userWallet).Update("is_primary", true).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "设置主钱包失败",
			"data": nil,
		})
		return
	}

	// 提交事务
	tx.Commit()

	// 更新本地对象
	userWallet.IsPrimary = true

	// 记录活动日志
	h.logUserActivity(userID.(uint), "user_wallet_set_primary", "user_wallet", walletID, c)

	// 返回更新后的钱包信息
	response := h.convertToUserWalletResponse(userWallet)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "主钱包设置成功",
		"data": response,
	})
}

// =============================================================================
// 辅助方法
// =============================================================================

/**
 * 验证以太坊地址格式
 */
func (h *UserWalletHandler) isValidEthereumAddress(address string) bool {
	// 以太坊地址正则表达式：0x开头，后跟40个十六进制字符
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}

	for _, char := range address[2:] {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') ||
			(char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}

/**
 * 构建排序子句
 */
func (h *UserWalletHandler) buildOrderClause(sortBy, sortOrder string) string {
	// 允许的排序字段
	allowedSortFields := map[string]bool{
		"created_at":   true,
		"updated_at":   true,
		"wallet_name":  true,
		"is_primary":   true,
		"last_used_at": true,
		"address":      true,
	}

	// 验证排序字段
	if !allowedSortFields[sortBy] {
		sortBy = "is_primary"
	}

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return fmt.Sprintf("%s %s", sortBy, sortOrder)
}

/**
 * 转换为UserWalletResponse格式
 */
func (h *UserWalletHandler) convertToUserWalletResponse(userWallet models.UserWallet) UserWalletResponse {
	return UserWalletResponse{
		ID:             userWallet.ID,
		Address:        userWallet.Address,
		WalletName:     userWallet.WalletName,
		WalletType:     userWallet.WalletType,
		DerivationPath: userWallet.DerivationPath,
		NetworkID:      userWallet.NetworkID,
		IsPrimary:      userWallet.IsPrimary,
		LastUsedAt:     userWallet.LastUsedAt,
		CreatedAt:      userWallet.CreatedAt,
		UpdatedAt:      userWallet.UpdatedAt,
	}
}

/**
 * 记录用户活动日志
 */
func (h *UserWalletHandler) logUserActivity(userID uint, action string, resourceType, resourceID string, c *gin.Context) {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	details := models.JSON{
		"timestamp": time.Now().Unix(),
		"endpoint":  c.FullPath(),
		"method":    c.Request.Method,
	}

	// 创建活动日志
	activityLog := models.ActivityLog{
		UserID:       &userID,
		Action:       action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
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
