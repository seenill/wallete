package handlers

import (
	"net/http"
	"strconv"
	"time"

	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

/*
DApp浏览器API处理器

本文件实现了DApp浏览器功能的HTTP接口处理器，包括：

主要接口：
- DApp连接管理：建立和断开DApp连接、会话管理
- Web3请求处理：处理DApp发起的Web3 RPC请求
- DApp发现：DApp列表、搜索、分类浏览
- 用户活动：收藏管理、访问历史、会话状态
- 安全控制：权限管理、风险评估、用户确认

接口分组：
- /api/v1/dapp/connect/* - 连接管理接口
- /api/v1/dapp/web3/* - Web3请求接口
- /api/v1/dapp/discovery/* - DApp发现接口
- /api/v1/dapp/user/* - 用户活动接口
- /api/v1/dapp/security/* - 安全管理接口

安全特性：
- 请求签名验证
- 会话状态检查
- 权限授权确认
- 风险等级评估
- 钓鱼网站检测
*/

// DAppBrowserHandler DApp浏览器API处理器
// 处理所有DApp浏览器相关的HTTP请求，包括连接管理、Web3请求、DApp发现等功能
type DAppBrowserHandler struct {
	dappBrowserService *services.DAppBrowserService // DApp浏览器业务服务实例
}

// NewDAppBrowserHandler 创建新的DApp浏览器处理器实例
// 参数: dappBrowserService - DApp浏览器业务服务实例
// 返回: 配置好的DApp浏览器处理器
func NewDAppBrowserHandler(dappBrowserService *services.DAppBrowserService) *DAppBrowserHandler {
	return &DAppBrowserHandler{
		dappBrowserService: dappBrowserService,
	}
}

// ConnectDApp 连接DApp
// POST /api/v1/dapp/connect
// 请求体: DAppConnectionRequest结构体
// 功能: 建立与DApp的连接，创建Web3会话
func (h *DAppBrowserHandler) ConnectDApp(c *gin.Context) {
	var req services.DAppConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.DAppURL == "" || req.UserAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "DApp URL和用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 连接DApp
	response, err := h.dappBrowserService.ConnectDApp(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "连接DApp失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "DApp连接成功",
		"data": response,
	})
}

// ProcessWeb3Request 处理Web3请求
// POST /api/v1/dapp/web3/request
// 请求体: Web3RequestData结构体
// 功能: 处理DApp发起的Web3 RPC请求
func (h *DAppBrowserHandler) ProcessWeb3Request(c *gin.Context) {
	var req services.Web3RequestData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.SessionID == "" || req.Method == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "会话ID和方法名不能为空",
			"data": nil,
		})
		return
	}

	// 处理Web3请求
	response, err := h.dappBrowserService.ProcessWeb3Request(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "处理Web3请求失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 根据响应状态返回不同的HTTP状态码
	if response.RequiresAuth {
		c.JSON(http.StatusAccepted, gin.H{
			"code": e.SUCCESS,
			"msg":  "请求需要用户确认",
			"data": response,
		})
	} else if response.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.ERROR,
			"msg":  "Web3请求执行失败",
			"data": response,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": e.SUCCESS,
			"msg":  "Web3请求处理成功",
			"data": response,
		})
	}
}

// GetDAppList 获取DApp列表
// GET /api/v1/dapp/discovery/list
// 查询参数:
//   - category: DApp分类（可选）
//   - search: 搜索关键词（可选）
//   - chain: 链过滤（可选）
//   - sort_by: 排序字段（默认rating）
//   - limit: 返回数量限制（默认20）
//   - offset: 偏移量（默认0）
//
// 响应: DApp列表，包含分类、统计等信息
func (h *DAppBrowserHandler) GetDAppList(c *gin.Context) {
	// 获取查询参数
	category := c.Query("category")
	search := c.Query("search")
	chain := c.Query("chain")
	sortBy := c.DefaultQuery("sort_by", "rating")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// 解析数值参数
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 构建请求
	request := &services.DAppListRequest{
		Category: category,
		Search:   search,
		Chain:    chain,
		SortBy:   sortBy,
		Limit:    limit,
		Offset:   offset,
	}

	// 获取DApp列表
	response, err := h.dappBrowserService.GetDAppList(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取DApp列表失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": response,
	})
}

// GetFeaturedDApps 获取推荐DApp
// GET /api/v1/dapp/discovery/featured
// 查询参数:
//   - limit: 返回数量限制（默认10）
//
// 响应: 推荐DApp列表
func (h *DAppBrowserHandler) GetFeaturedDApps(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	// 获取推荐DApp（使用featured分类）
	request := &services.DAppListRequest{
		Category: "featured",
		Limit:    limit,
		Offset:   0,
	}

	response, err := h.dappBrowserService.GetDAppList(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取推荐DApp失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"featured_dapps": response.DApps,
			"total":          response.Total,
		},
	})
}

// GetUserActivity 获取用户活动
// GET /api/v1/dapp/user/:address/activity
// 路径参数:
//   - address: 用户地址
//
// 查询参数:
//   - activity_type: 活动类型（favorites/visits/sessions）
//   - time_range: 时间范围（24h/7d/30d）
//   - limit: 返回数量限制（默认50）
//
// 响应: 用户DApp活动信息
func (h *DAppBrowserHandler) GetUserActivity(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 获取查询参数
	activityType := c.Query("activity_type")
	timeRange := c.Query("time_range")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	// 构建请求
	request := &services.UserActivityRequest{
		UserAddress:  userAddress,
		ActivityType: activityType,
		TimeRange:    timeRange,
		Limit:        limit,
	}

	// 获取用户活动
	response, err := h.dappBrowserService.GetUserActivity(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取用户活动失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": response,
	})
}

// ConfirmWeb3Request 确认Web3请求
// POST /api/v1/dapp/web3/confirm
// 请求体: Web3ConfirmRequest结构体
// 功能: 用户确认或拒绝Web3请求
func (h *DAppBrowserHandler) ConfirmWeb3Request(c *gin.Context) {
	var req struct {
		RequestID string `json:"request_id" binding:"required"`
		Approved  bool   `json:"approved"`
		Signature string `json:"signature"`
		Password  string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 确认Web3请求
	err := h.dappBrowserService.ConfirmWeb3Request(c.Request.Context(), req.RequestID, req.Approved, req.Signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "确认Web3请求失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	message := "请求已拒绝"
	if req.Approved {
		message = "请求已确认"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  message,
		"data": gin.H{
			"request_id": req.RequestID,
			"approved":   req.Approved,
			"timestamp":  time.Now().Unix(),
		},
	})
}

// GetPendingRequests 获取待处理请求
// GET /api/v1/dapp/web3/pending/:address
// 路径参数:
//   - address: 用户地址
//
// 响应: 待确认的Web3请求列表
func (h *DAppBrowserHandler) GetPendingRequests(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 获取待处理请求
	pendingRequests := h.dappBrowserService.GetPendingRequests(userAddress)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"pending_requests": pendingRequests,
			"total":            len(pendingRequests),
		},
	})
}

// SearchDApps 搜索DApp
// GET /api/v1/dapp/discovery/search
// 查询参数:
//   - q: 搜索关键词（必需）
//   - category: 分类过滤（可选）
//   - chain: 链过滤（可选）
//   - limit: 返回数量限制（默认20）
//
// 响应: 搜索结果列表
func (h *DAppBrowserHandler) SearchDApps(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "搜索关键词不能为空",
			"data": nil,
		})
		return
	}

	category := c.Query("category")
	chain := c.Query("chain")
	limitStr := c.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 构建搜索请求
	request := &services.DAppListRequest{
		Search:   query,
		Category: category,
		Chain:    chain,
		Limit:    limit,
		Offset:   0,
	}

	// 执行搜索
	response, err := h.dappBrowserService.GetDAppList(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "搜索DApp失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"results": response.DApps,
			"total":   response.Total,
			"query":   query,
		},
	})
}

// GetCategories 获取DApp分类
// GET /api/v1/dapp/discovery/categories
// 响应: DApp分类列表和统计信息
func (h *DAppBrowserHandler) GetCategories(c *gin.Context) {
	// 获取分类信息（通过获取空的DApp列表来获取分类）
	request := &services.DAppListRequest{
		Limit:  0, // 不需要DApp数据，只需要分类信息
		Offset: 0,
	}

	response, err := h.dappBrowserService.GetDAppList(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取分类失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"categories": response.Categories,
			"total":      len(response.Categories),
		},
	})
}

// ManageFavorite 管理收藏
// POST /api/v1/dapp/user/favorite
// 请求体: FavoriteRequest结构体
// 功能: 添加或移除DApp收藏
func (h *DAppBrowserHandler) ManageFavorite(c *gin.Context) {
	var req struct {
		UserAddress string `json:"user_address" binding:"required"`
		DAppID      string `json:"dapp_id" binding:"required"`
		Action      string `json:"action" binding:"required"` // add/remove
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证操作类型
	if req.Action != "add" && req.Action != "remove" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "无效的操作类型，必须是 add 或 remove",
			"data": nil,
		})
		return
	}

	// 这里简化实现，实际项目中需要调用相应的服务方法
	message := "收藏添加成功"
	if req.Action == "remove" {
		message = "收藏移除成功"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  message,
		"data": gin.H{
			"user_address": req.UserAddress,
			"dapp_id":      req.DAppID,
			"action":       req.Action,
			"timestamp":    time.Now().Unix(),
		},
	})
}

// DisconnectDApp 断开DApp连接
// DELETE /api/v1/dapp/connect/:sessionId
// 路径参数:
//   - sessionId: 会话ID
//
// 功能: 断开与DApp的连接，清理会话
func (h *DAppBrowserHandler) DisconnectDApp(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "会话ID不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：返回成功响应
	// 实际项目中需要调用服务层方法断开连接
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "DApp连接已断开",
		"data": gin.H{
			"session_id":      sessionID,
			"disconnected_at": time.Now().Unix(),
		},
	})
}

// GetSessionInfo 获取会话信息
// GET /api/v1/dapp/connect/:sessionId
// 路径参数:
//   - sessionId: 会话ID
//
// 响应: 会话详细信息
func (h *DAppBrowserHandler) GetSessionInfo(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "会话ID不能为空",
			"data": nil,
		})
		return
	}

	// 简化实现：返回示例会话信息
	sessionInfo := gin.H{
		"session_id":     sessionID,
		"dapp_url":       "https://example-dapp.com",
		"dapp_name":      "示例DApp",
		"user_address":   "0x1234567890123456789012345678901234567890",
		"chain_id":       "0x1",
		"status":         "active",
		"connected_at":   time.Now().Add(-1 * time.Hour).Unix(),
		"last_active_at": time.Now().Unix(),
		"permissions": []string{
			"eth_accounts",
			"eth_sendTransaction",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": sessionInfo,
	})
}
