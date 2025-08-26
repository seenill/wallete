/**
 * 观察地址管理API处理器
 *
 * 本包实现了观察地址管理相关的HTTP API处理器，集成数据库进行地址管理。
 *
 * 主要功能：
 * 地址管理：
 * - 添加观察地址并验证地址格式
 * - 查询用户的观察地址列表
 * - 更新地址标签、备注等信息
 * - 删除观察地址
 *
 * 地址监控：
 * - 地址余额缓存和历史记录
 * - 支持多网络（以太坊、Polygon等）
 * - 地址类型识别（EOA、合约、多签）
 * - 收藏和通知设置
 *
 * 数据验证：
 * - 以太坊地址格式验证
 * - 用户权限验证
 * - 重复地址检查
 *
 * 学习要点：
 * 1. 地址验证 - 以太坊地址格式和校验和验证
 * 2. 分页查询 - 数据库分页和排序
 * 3. 条件过滤 - 动态查询条件构建
 * 4. 数据关联 - GORM的关联查询
 * 5. 错误处理 - 业务逻辑错误处理
 */
package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"wallet/database"
	"wallet/models"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WatchAddressHandler 观察地址管理相关的HTTP请求处理器
// 负责处理观察地址的增删改查等功能
type WatchAddressHandler struct{}

// NewWatchAddressHandler 创建新的观察地址处理器实例
// 返回: 初始化完成的WatchAddressHandler指针
func NewWatchAddressHandler() *WatchAddressHandler {
	return &WatchAddressHandler{}
}

// =============================================================================
// 请求和响应结构体定义
// =============================================================================

// AddWatchAddressRequest 添加观察地址请求
type AddWatchAddressRequest struct {
	Address             string            `json:"address" binding:"required"`
	Label               *string           `json:"label,omitempty"`
	NetworkID           *int              `json:"network_id,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
	Notes               *string           `json:"notes,omitempty"`
	IsFavorite          *bool             `json:"is_favorite,omitempty"`
	NotificationEnabled *bool             `json:"notification_enabled,omitempty"`
}

// UpdateWatchAddressRequest 更新观察地址请求
type UpdateWatchAddressRequest struct {
	Label               *string           `json:"label,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
	Notes               *string           `json:"notes,omitempty"`
	IsFavorite          *bool             `json:"is_favorite,omitempty"`
	NotificationEnabled *bool             `json:"notification_enabled,omitempty"`
}

// WatchAddressResponse 观察地址响应
type WatchAddressResponse struct {
	ID                  uint                            `json:"id"`
	Address             string                          `json:"address"`
	Label               *string                         `json:"label,omitempty"`
	NetworkID           int                             `json:"network_id"`
	AddressType         string                          `json:"address_type"`
	Tags                models.JSON                     `json:"tags,omitempty"`
	Notes               *string                         `json:"notes,omitempty"`
	IsFavorite          bool                            `json:"is_favorite"`
	NotificationEnabled bool                            `json:"notification_enabled"`
	BalanceCache        *string                         `json:"balance_cache,omitempty"`
	LastActivityAt      *time.Time                      `json:"last_activity_at,omitempty"`
	CreatedAt           time.Time                       `json:"created_at"`
	UpdatedAt           time.Time                       `json:"updated_at"`
	BalanceHistory      []AddressBalanceHistoryResponse `json:"balance_history,omitempty"`
}

// AddressBalanceHistoryResponse 地址余额历史响应
type AddressBalanceHistoryResponse struct {
	ID           uint      `json:"id"`
	Balance      string    `json:"balance"`
	TokenAddress *string   `json:"token_address,omitempty"`
	TokenSymbol  *string   `json:"token_symbol,omitempty"`
	BlockNumber  *uint64   `json:"block_number,omitempty"`
	RecordedAt   time.Time `json:"recorded_at"`
}

// WatchAddressListResponse 观察地址列表响应
type WatchAddressListResponse struct {
	Total     int64                  `json:"total"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"page_size"`
	Addresses []WatchAddressResponse `json:"addresses"`
}

// =============================================================================
// 观察地址管理
// =============================================================================

/**
 * 添加观察地址
 * 创建新的观察地址记录，包括地址验证和重复检查
 */
func (h *WatchAddressHandler) AddWatchAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	var req AddWatchAddressRequest
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

	// 标准化地址格式（转换为小写）
	address := strings.ToLower(req.Address)

	// 检查是否已存在
	networkID := 1 // 默认以太坊主网
	if req.NetworkID != nil {
		networkID = *req.NetworkID
	}

	var existing models.WatchAddress
	result := database.DB.Where("user_id = ? AND address = ? AND network_id = ?",
		userID, address, networkID).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{
			"code": e.ERROR,
			"msg":  "该地址已在观察列表中",
			"data": nil,
		})
		return
	}

	// 准备标签数据
	var tags models.JSON
	if req.Tags != nil {
		// 将map[string]string转换为models.JSON
		tagsMap := make(map[string]interface{})
		for k, v := range req.Tags {
			tagsMap[k] = v
		}
		tags = models.JSON(tagsMap)
	} else {
		tags = models.JSON{}
	}

	// 检测地址类型
	addressType := h.detectAddressType(address)

	// 创建观察地址记录
	watchAddress := models.WatchAddress{
		UserID:              userID.(uint),
		Address:             address,
		Label:               req.Label,
		NetworkID:           networkID,
		AddressType:         addressType,
		Tags:                tags,
		Notes:               req.Notes,
		IsFavorite:          h.getBoolWithDefault(req.IsFavorite, false),
		NotificationEnabled: h.getBoolWithDefault(req.NotificationEnabled, true),
	}

	if err := database.DB.Create(&watchAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "添加观察地址失败",
			"data": err.Error(),
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "watch_address_add", "watch_address",
		fmt.Sprintf("%d", watchAddress.ID), c)

	// 返回创建的地址信息
	response := h.convertToWatchAddressResponse(watchAddress)

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "观察地址添加成功",
		"data": response,
	})
}

/**
 * 获取观察地址列表
 * 支持分页、排序和筛选
 */
func (h *WatchAddressHandler) GetWatchAddresses(c *gin.Context) {
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
	favorite := c.Query("favorite")
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")

	// 限制分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建查询
	query := database.DB.Model(&models.WatchAddress{}).Where("user_id = ?", userID)

	// 网络过滤
	if networkID != "" {
		if netID, err := strconv.Atoi(networkID); err == nil {
			query = query.Where("network_id = ?", netID)
		}
	}

	// 收藏过滤
	if favorite == "true" {
		query = query.Where("is_favorite = ?", true)
	} else if favorite == "false" {
		query = query.Where("is_favorite = ?", false)
	}

	// 搜索过滤
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("address ILIKE ? OR label ILIKE ? OR notes ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "查询观察地址失败",
			"data": nil,
		})
		return
	}

	// 排序
	orderClause := h.buildOrderClause(sortBy, sortOrder)
	query = query.Order(orderClause)

	// 分页查询
	var addresses []models.WatchAddress
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&addresses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "查询观察地址失败",
			"data": nil,
		})
		return
	}

	// 转换响应格式
	responseAddresses := make([]WatchAddressResponse, len(addresses))
	for i, addr := range addresses {
		responseAddresses[i] = h.convertToWatchAddressResponse(addr)
	}

	response := WatchAddressListResponse{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		Addresses: responseAddresses,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

/**
 * 获取单个观察地址详情
 * 包含余额历史记录
 */
func (h *WatchAddressHandler) GetWatchAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少地址ID",
			"data": nil,
		})
		return
	}

	// 查询观察地址
	var watchAddress models.WatchAddress
	result := database.DB.Where("id = ? AND user_id = ?", addressID, userID).
		Preload("BalanceHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("recorded_at DESC").Limit(50) // 最近50条记录
		}).First(&watchAddress)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": e.ERROR,
				"msg":  "观察地址不存在",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询观察地址失败",
				"data": nil,
			})
		}
		return
	}

	// 转换响应格式
	response := h.convertToWatchAddressResponse(watchAddress)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": response,
	})
}

/**
 * 更新观察地址
 * 允许更新标签、备注、收藏状态等
 */
func (h *WatchAddressHandler) UpdateWatchAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少地址ID",
			"data": nil,
		})
		return
	}

	var req UpdateWatchAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 查询观察地址
	var watchAddress models.WatchAddress
	result := database.DB.Where("id = ? AND user_id = ?", addressID, userID).First(&watchAddress)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code": e.ERROR,
				"msg":  "观察地址不存在",
				"data": nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  "查询观察地址失败",
				"data": nil,
			})
		}
		return
	}

	// 更新字段
	if req.Label != nil {
		watchAddress.Label = req.Label
	}
	if req.Notes != nil {
		watchAddress.Notes = req.Notes
	}
	if req.IsFavorite != nil {
		watchAddress.IsFavorite = *req.IsFavorite
	}
	if req.NotificationEnabled != nil {
		watchAddress.NotificationEnabled = *req.NotificationEnabled
	}
	// 更新标签
	if req.Tags != nil {
		// 将map[string]string转换为models.JSON
		tagsMap := make(map[string]interface{})
		for k, v := range req.Tags {
			tagsMap[k] = v
		}
		watchAddress.Tags = models.JSON(tagsMap)
	}

	// 保存更改
	if err := database.DB.Save(&watchAddress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新观察地址失败",
			"data": nil,
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "watch_address_update", "watch_address", addressID, c)

	// 返回更新后的地址信息
	response := h.convertToWatchAddressResponse(watchAddress)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "观察地址更新成功",
		"data": response,
	})
}

/**
 * 删除观察地址
 * 软删除地址记录
 */
func (h *WatchAddressHandler) DeleteWatchAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": e.ERROR,
			"msg":  "用户未认证",
			"data": nil,
		})
		return
	}

	addressID := c.Param("id")
	if addressID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少地址ID",
			"data": nil,
		})
		return
	}

	// 查询并删除观察地址
	result := database.DB.Where("id = ? AND user_id = ?", addressID, userID).
		Delete(&models.WatchAddress{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "删除观察地址失败",
			"data": nil,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "观察地址不存在",
			"data": nil,
		})
		return
	}

	// 记录活动日志
	h.logUserActivity(userID.(uint), "watch_address_delete", "watch_address", addressID, c)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "观察地址删除成功",
		"data": nil,
	})
}

// =============================================================================
// 辅助方法
// =============================================================================

/**
 * 验证以太坊地址格式
 * 检查地址是否符合以太坊地址规范
 */
func (h *WatchAddressHandler) isValidEthereumAddress(address string) bool {
	// 以太坊地址正则表达式：0x开头，后跟40个十六进制字符
	pattern := `^0x[a-fA-F0-9]{40}$`
	matched, err := regexp.MatchString(pattern, address)
	return err == nil && matched
}

/**
 * 检测地址类型
 * 简单的地址类型检测，实际应该通过区块链查询
 */
func (h *WatchAddressHandler) detectAddressType(address string) string {
	// 这里是简化实现，实际应该查询区块链
	// 检查是否是合约地址需要调用eth_getCode
	return "EOA" // 默认为外部账户
}

/**
 * 获取布尔值或默认值
 */
func (h *WatchAddressHandler) getBoolWithDefault(value *bool, defaultValue bool) bool {
	if value != nil {
		return *value
	}
	return defaultValue
}

/**
 * 构建排序子句
 */
func (h *WatchAddressHandler) buildOrderClause(sortBy, sortOrder string) string {
	// 允许的排序字段
	allowedSortFields := map[string]bool{
		"created_at":       true,
		"updated_at":       true,
		"address":          true,
		"label":            true,
		"is_favorite":      true,
		"last_activity_at": true,
	}

	// 验证排序字段
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}

	// 验证排序方向
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return fmt.Sprintf("%s %s", sortBy, sortOrder)
}

/**
 * 转换为WatchAddressResponse格式
 */
func (h *WatchAddressHandler) convertToWatchAddressResponse(watchAddress models.WatchAddress) WatchAddressResponse {
	response := WatchAddressResponse{
		ID:                  watchAddress.ID,
		Address:             watchAddress.Address,
		Label:               watchAddress.Label,
		NetworkID:           watchAddress.NetworkID,
		AddressType:         watchAddress.AddressType,
		Tags:                watchAddress.Tags,
		Notes:               watchAddress.Notes,
		IsFavorite:          watchAddress.IsFavorite,
		NotificationEnabled: watchAddress.NotificationEnabled,
		BalanceCache:        watchAddress.BalanceCache,
		LastActivityAt:      watchAddress.LastActivityAt,
		CreatedAt:           watchAddress.CreatedAt,
		UpdatedAt:           watchAddress.UpdatedAt,
	}

	// 转换余额历史
	if len(watchAddress.BalanceHistory) > 0 {
		history := make([]AddressBalanceHistoryResponse, len(watchAddress.BalanceHistory))
		for i, h := range watchAddress.BalanceHistory {
			history[i] = AddressBalanceHistoryResponse{
				ID:           h.ID,
				Balance:      h.Balance,
				TokenAddress: h.TokenAddress,
				TokenSymbol:  h.TokenSymbol,
				BlockNumber:  h.BlockNumber,
				RecordedAt:   h.RecordedAt,
			}
		}
		response.BalanceHistory = history
	}

	return response
}

/**
 * 记录用户活动日志
 * 统一记录用户操作日志，用于安全审计
 */
func (h *WatchAddressHandler) logUserActivity(userID uint, action string, resourceType, resourceID string, c *gin.Context) {
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
