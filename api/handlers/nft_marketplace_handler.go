/*
NFT市场API处理器

本文件实现了NFT市场功能的HTTP接口处理器。
*/
package handlers

import (
	"math/big"
	"net/http"
	"strconv"

	"wallet/core"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NFTMarketplaceHandler NFT市场API处理器
type NFTMarketplaceHandler struct {
	marketplaceService *services.NFTMarketplaceService
}

// NewNFTMarketplaceHandler 创建NFT市场处理器
func NewNFTMarketplaceHandler(marketplaceService *services.NFTMarketplaceService) *NFTMarketplaceHandler {
	return &NFTMarketplaceHandler{
		marketplaceService: marketplaceService,
	}
}

// GetMarketListings 获取市场挂单
// GET /api/v1/nft/marketplace/listings
func (h *NFTMarketplaceHandler) GetMarketListings(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	contract := c.Query("contract")
	platform := c.Query("platform")

	if contract == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "合约地址不能为空",
			"data": nil,
		})
		return
	}

	// 解析分页参数
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	request := &core.MarketListingRequest{
		Platform:  platform,
		Contract:  contract,
		TokenID:   c.Query("token_id"),
		Seller:    c.Query("seller"),
		Currency:  c.Query("currency"),
		Status:    c.DefaultQuery("status", "active"),
		SortBy:    c.DefaultQuery("sort_by", "price"),
		SortOrder: c.DefaultQuery("sort_order", "asc"),
		Limit:     limit,
		Offset:    offset,
	}

	// 解析价格范围
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, ok := new(big.Int).SetString(minPriceStr, 10); ok {
			request.MinPrice = minPrice
		}
	}
	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, ok := new(big.Int).SetString(maxPriceStr, 10); ok {
			request.MaxPrice = maxPrice
		}
	}

	listings, err := h.marketplaceService.GetMarketListings(c.Request.Context(), userAddress, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取市场挂单失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": gin.H{
			"listings": listings,
			"count":    len(listings),
			"request":  request,
		},
	})
}

// GetMarketTransactions 获取市场交易记录
// GET /api/v1/nft/marketplace/transactions
func (h *NFTMarketplaceHandler) GetMarketTransactions(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	contract := c.Query("contract")

	if contract == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "合约地址不能为空",
			"data": nil,
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	request := &core.MarketTransactionRequest{
		Platform:    c.Query("platform"),
		Contract:    contract,
		TokenID:     c.Query("token_id"),
		FromAddress: c.Query("from_address"),
		ToAddress:   c.Query("to_address"),
		Type:        c.Query("type"),
		SortBy:      c.DefaultQuery("sort_by", "timestamp"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
		Limit:       limit,
		Offset:      offset,
	}

	transactions, err := h.marketplaceService.GetMarketTransactions(c.Request.Context(), userAddress, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取交易记录失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": gin.H{
			"transactions": transactions,
			"count":        len(transactions),
		},
	})
}

// GetMarketStats 获取市场统计
// GET /api/v1/nft/marketplace/stats/:contract
func (h *NFTMarketplaceHandler) GetMarketStats(c *gin.Context) {
	contract := c.Param("contract")
	platform := c.Query("platform")

	if contract == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "合约地址不能为空",
			"data": nil,
		})
		return
	}

	stats, err := h.marketplaceService.GetMarketStats(c.Request.Context(), contract, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取市场统计失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": stats,
	})
}

// AnalyzeMarket 市场分析
// POST /api/v1/nft/marketplace/analyze
func (h *NFTMarketplaceHandler) AnalyzeMarket(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")

	var req services.MarketAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	if req.Contract == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "合约地址不能为空",
			"data": nil,
		})
		return
	}

	analysis, err := h.marketplaceService.AnalyzeMarket(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "市场分析失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "分析完成",
		"data": analysis,
	})
}

// SetUserPreferences 设置用户偏好
// POST /api/v1/nft/marketplace/preferences
func (h *NFTMarketplaceHandler) SetUserPreferences(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	var prefs services.UserMarketPrefs
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	h.marketplaceService.SetUserPreferences(userAddress, &prefs)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "偏好设置成功",
		"data": nil,
	})
}

// GetUserPreferences 获取用户偏好
// GET /api/v1/nft/marketplace/preferences
func (h *NFTMarketplaceHandler) GetUserPreferences(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	prefs := h.marketplaceService.GetUserPreferences(userAddress)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": prefs,
	})
}

// AddToWatchlist 添加到关注列表
// POST /api/v1/nft/marketplace/watchlist
func (h *NFTMarketplaceHandler) AddToWatchlist(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	var req struct {
		ListName string `json:"list_name" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Contract string `json:"contract" binding:"required"`
		TokenID  string `json:"token_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	err := h.marketplaceService.AddToWatchlist(userAddress, req.ListName, req.Type, req.Contract, req.TokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "添加失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "添加成功",
		"data": nil,
	})
}

// GetWatchlist 获取关注列表
// GET /api/v1/nft/marketplace/watchlist/:listName
func (h *NFTMarketplaceHandler) GetWatchlist(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	listName := c.Param("listName")

	if userAddress == "" || listName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址和列表名称不能为空",
			"data": nil,
		})
		return
	}

	watchlist, err := h.marketplaceService.GetWatchlist(userAddress, listName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code": e.ERROR,
			"msg":  "获取失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": watchlist,
	})
}

// CreatePriceAlert 创建价格提醒
// POST /api/v1/nft/marketplace/price-alert
func (h *NFTMarketplaceHandler) CreatePriceAlert(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	var req struct {
		Contract    string `json:"contract" binding:"required"`
		TokenID     string `json:"token_id"`
		AlertType   string `json:"alert_type" binding:"required"`
		TargetPrice struct {
			Amount   string `json:"amount" binding:"required"`
			Currency string `json:"currency" binding:"required"`
		} `json:"target_price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 解析目标价格
	amount, ok := new(big.Int).SetString(req.TargetPrice.Amount, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "无效的价格格式",
			"data": nil,
		})
		return
	}

	targetPrice := &core.MarketPrice{
		Amount:   amount,
		Currency: req.TargetPrice.Currency,
		Symbol:   req.TargetPrice.Currency,
	}

	alert, err := h.marketplaceService.CreatePriceAlert(userAddress, req.Contract, req.TokenID, req.AlertType, targetPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建提醒失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": e.SUCCESS,
		"msg":  "价格提醒创建成功",
		"data": alert,
	})
}

// GetPriceAlerts 获取价格提醒
// GET /api/v1/nft/marketplace/price-alerts
func (h *NFTMarketplaceHandler) GetPriceAlerts(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	alerts, err := h.marketplaceService.GetPriceAlerts(userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取提醒失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "获取成功",
		"data": gin.H{
			"alerts": alerts,
			"count":  len(alerts),
		},
	})
}
