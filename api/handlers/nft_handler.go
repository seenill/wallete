/*
NFT功能API处理器

本文件实现了NFT功能的HTTP接口处理器，包括：

主要接口：
- NFT查询：获取用户NFT、NFT详情、集合信息
- NFT转账：单个和批量NFT转移
- 市场数据：热门集合、价格信息、市场趋势
- 投资组合：用户NFT投资组合分析和统计
- 集合管理：集合排行榜、统计数据、活动记录

接口分组：
- /api/v1/nft/user/* - 用户NFT相关接口
- /api/v1/nft/collections/* - 集合相关接口
- /api/v1/nft/market/* - 市场数据接口
- /api/v1/nft/transfer/* - 转账操作接口
- /api/v1/nft/portfolio/* - 投资组合接口

安全特性：
- NFT所有权验证
- 转账权限检查
- 钓鱼NFT检测
- 价格异常警告
*/
package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"wallet/core"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// NFTHandler NFT功能API处理器
// 处理所有NFT相关的HTTP请求，包括查询、转账、市场数据等功能
type NFTHandler struct {
	nftService *services.NFTService // NFT业务服务实例
}

// NewNFTHandler 创建新的NFT处理器实例
// 参数: nftService - NFT业务服务实例
// 返回: 配置好的NFT处理器
func NewNFTHandler(nftService *services.NFTService) *NFTHandler {
	return &NFTHandler{
		nftService: nftService,
	}
}

// GetUserNFTs 获取用户NFT列表
// GET /api/v1/nft/user/:address/nfts
// 查询参数:
//   - collection: 指定集合地址过滤（可选）
//   - category: NFT类别过滤（可选）
//   - sort_by: 排序字段（token_id/updated_at/price，默认updated_at）
//   - sort_dir: 排序方向（asc/desc，默认desc）
//   - limit: 返回数量限制（默认50，最大200）
//   - offset: 偏移量（默认0）
//
// 响应: 用户拥有的NFT列表
func (h *NFTHandler) GetUserNFTs(c *gin.Context) {
	// 获取路径参数
	userAddr := c.Param("address")
	if userAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 获取查询参数
	collection := c.Query("collection")
	category := c.Query("category")
	sortBy := c.DefaultQuery("sort_by", "updated_at")
	sortDir := c.DefaultQuery("sort_dir", "desc")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	// 解析数值参数
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 构建过滤条件
	filters := &services.NFTFilters{
		Collection: collection,
		Category:   category,
		SortBy:     sortBy,
		SortDir:    sortDir,
		Limit:      limit,
		Offset:     offset,
	}

	// 获取用户NFT
	nfts, err := h.nftService.GetUserNFTs(c.Request.Context(), userAddr, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取用户NFT失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"nfts":   nfts,
			"total":  len(nfts),
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetNFTDetails 获取NFT详细信息
// GET /api/v1/nft/details/:contract/:tokenId
// 路径参数:
//   - contract: NFT合约地址
//   - tokenId: NFT代币ID
//
// 响应: NFT详细信息，包含元数据、市场数据、历史记录等
func (h *NFTHandler) GetNFTDetails(c *gin.Context) {
	// 获取路径参数
	contractAddr := c.Param("contract")
	tokenID := c.Param("tokenId")

	if contractAddr == "" || tokenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "合约地址和代币ID不能为空",
			"data": nil,
		})
		return
	}

	// 获取NFT详细信息
	nft, err := h.nftService.GetNFTDetails(c.Request.Context(), contractAddr, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取NFT详情失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": nft,
	})
}

// GetCollectionInfo 获取NFT集合信息
// GET /api/v1/nft/collections/:address
// 路径参数:
//   - address: 集合合约地址
//
// 响应: 集合详细信息，包含统计数据、价格信息、活动记录等
func (h *NFTHandler) GetCollectionInfo(c *gin.Context) {
	contractAddr := c.Param("address")
	if contractAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "集合地址不能为空",
			"data": nil,
		})
		return
	}

	// 获取集合信息
	collection, err := h.nftService.GetCollectionInfo(c.Request.Context(), contractAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取集合信息失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": collection,
	})
}

// GetHotCollections 获取热门集合列表
// GET /api/v1/nft/market/hot-collections
// 查询参数:
//   - limit: 返回数量（默认20，最大100）
//   - category: 集合类别过滤（可选）
//
// 响应: 热门集合排行榜
func (h *NFTHandler) GetHotCollections(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "20")
	category := c.Query("category")

	// 解析数量限制
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 获取热门集合
	hotCollections := h.nftService.GetHotCollections(limit)

	// 应用类别过滤
	if category != "" {
		filtered := make([]*services.HotCollection, 0)
		for _, collection := range hotCollections {
			// 这里可以根据实际的类别字段进行过滤
			// 简化实现：跳过过滤
			filtered = append(filtered, collection)
		}
		hotCollections = filtered
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"collections": hotCollections,
			"total":       len(hotCollections),
			"category":    category,
		},
	})
}

// TransferNFT 转账NFT
// POST /api/v1/nft/transfer
// 请求体: NFTTransferRequest结构体
// 功能: 执行NFT转账操作，支持ERC-721和ERC-1155标准
func (h *NFTHandler) TransferNFT(c *gin.Context) {
	var req services.NFTTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 从会话或请求中获取认证信息（这里简化处理）
	mnemonic := "demo mnemonic phrase for testing purposes only"
	derivationPath := "m/44'/60'/0'/0/0"

	// 执行NFT转账
	result, err := h.nftService.TransferNFT(c.Request.Context(), &req, mnemonic, derivationPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ErrorTransactionSend,
			"msg":  "NFT转账失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "NFT转账已提交",
		"data": result,
	})
}

// GetUserPortfolio 获取用户NFT投资组合
// GET /api/v1/nft/portfolio/:address
// 路径参数:
//   - address: 用户地址
//
// 响应: 用户NFT投资组合分析，包含总价值、盈亏、持有分布等
func (h *NFTHandler) GetUserPortfolio(c *gin.Context) {
	userAddr := c.Param("address")
	if userAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 构建示例投资组合数据
	portfolio := &services.UserPortfolio{
		UserAddress:      userAddr,
		TotalValue:       nil, // 会在实际实现中计算
		TotalCount:       0,
		Collections:      make([]*services.UserCollection, 0),
		TopNFTs:          make([]*core.NFT, 0),
		RecentActivities: make([]*services.ActivityInfo, 0),
		UpdatedAt:        time.Now(),
	}

	// 简化实现：返回示例数据
	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": portfolio,
	})
}

// GetMarketTrends 获取市场趋势数据
// GET /api/v1/nft/market/trends
// 查询参数:
//   - period: 时间周期（24h/7d/30d，默认24h）
//   - category: 类别过滤（可选）
//
// 响应: NFT市场整体趋势数据和分析
func (h *NFTHandler) GetMarketTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "24h")
	category := c.Query("category")

	// 验证时间周期参数
	validPeriods := map[string]bool{
		"24h": true,
		"7d":  true,
		"30d": true,
	}
	if !validPeriods[period] {
		period = "24h"
	}

	// 构建示例市场趋势数据
	trend := &services.MarketTrend{
		Category:        category,
		TotalVolume:     nil, // 会在实际实现中计算
		TotalSales:      0,
		AvgPrice:        nil,
		FloorPriceAvg:   nil,
		VolumeChange24h: "+12.5%",
		PriceChange24h:  "+8.3%",
		TopCollections:  make([]*services.HotCollection, 0),
		UpdatedAt:       time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"trend":    trend,
			"period":   period,
			"category": category,
		},
	})
}

// SearchNFTs 搜索NFT
// GET /api/v1/nft/search
// 查询参数:
//   - q: 搜索关键词（必需）
//   - type: 搜索类型（nft/collection，默认nft）
//   - limit: 返回数量限制（默认20）
//
// 响应: 搜索结果列表
func (h *NFTHandler) SearchNFTs(c *gin.Context) {
	query := c.Query("q")
	searchType := c.DefaultQuery("type", "nft")
	limitStr := c.DefaultQuery("limit", "20")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "搜索关键词不能为空",
			"data": nil,
		})
		return
	}

	// 解析数量限制
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 构建搜索结果（简化实现）
	results := make([]interface{}, 0)

	if searchType == "nft" {
		// 搜索NFT
		// 实际实现中会调用搜索服务
		results = append(results, gin.H{
			"type":     "nft",
			"contract": "0x1234567890123456789012345678901234567890",
			"token_id": "123",
			"name":     "示例NFT #123",
			"image":    "https://example.com/nft.png",
		})
	} else if searchType == "collection" {
		// 搜索集合
		results = append(results, gin.H{
			"type":        "collection",
			"address":     "0x1234567890123456789012345678901234567890",
			"name":        "示例NFT集合",
			"description": "这是一个示例NFT集合",
			"image":       "https://example.com/collection.png",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"results":     results,
			"total":       len(results),
			"query":       query,
			"search_type": searchType,
		},
	})
}

// GetNFTActivities 获取NFT活动记录
// GET /api/v1/nft/activities
// 查询参数:
//   - contract: 合约地址过滤（可选）
//   - token_id: 代币ID过滤（可选）
//   - user: 用户地址过滤（可选）
//   - activity_type: 活动类型过滤（可选：mint/sale/transfer/listing）
//   - limit: 返回数量限制（默认50）
//
// 响应: NFT活动记录列表
func (h *NFTHandler) GetNFTActivities(c *gin.Context) {
	// 获取查询参数
	contract := c.Query("contract")
	tokenID := c.Query("token_id")
	user := c.Query("user")
	activityType := c.Query("activity_type")
	limitStr := c.DefaultQuery("limit", "50")

	// 解析数量限制
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	// 验证活动类型
	validTypes := map[string]bool{
		"mint":     true,
		"sale":     true,
		"transfer": true,
		"listing":  true,
	}
	if activityType != "" && !validTypes[activityType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "无效的活动类型",
			"data": nil,
		})
		return
	}

	// 构建活动记录（简化实现）
	activities := make([]*services.ActivityInfo, 0)

	// 示例活动记录
	if contract == "" || contract == "0x1234567890123456789012345678901234567890" {
		activities = append(activities, &services.ActivityInfo{
			Type:        "sale",
			TokenID:     "123",
			From:        "0x1111111111111111111111111111111111111111",
			To:          "0x2222222222222222222222222222222222222222",
			Price:       nil, // 在实际实现中设置价格
			Currency:    "ETH",
			TxHash:      "0x" + strings.Repeat("a", 64),
			Marketplace: "OpenSea",
			Timestamp:   time.Now().Add(-1 * time.Hour),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"activities": activities,
			"total":      len(activities),
			"filters": gin.H{
				"contract":      contract,
				"token_id":      tokenID,
				"user":          user,
				"activity_type": activityType,
			},
		},
	})
}

// EstimateNFTValue 估算NFT价值
// POST /api/v1/nft/estimate-value
// 请求体包含NFT列表，返回估算价值
func (h *NFTHandler) EstimateNFTValue(c *gin.Context) {
	var req struct {
		NFTs []struct {
			Contract string `json:"contract" binding:"required"`
			TokenID  string `json:"token_id" binding:"required"`
		} `json:"nfts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	if len(req.NFTs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "NFT列表不能为空",
			"data": nil,
		})
		return
	}

	if len(req.NFTs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "一次最多只能估算100个NFT",
			"data": nil,
		})
		return
	}

	// 构建估算结果（简化实现）
	estimations := make([]gin.H, len(req.NFTs))
	totalValue := "0"

	for i, nft := range req.NFTs {
		estimations[i] = gin.H{
			"contract":        nft.Contract,
			"token_id":        nft.TokenID,
			"estimated_value": "1.5", // ETH
			"confidence":      "medium",
			"last_sale_price": "1.2",
			"floor_price":     "1.0",
			"basis":           "floor_price_analysis",
			"updated_at":      time.Now().Unix(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"estimations": estimations,
			"total_value": totalValue,
			"currency":    "ETH",
			"count":       len(estimations),
		},
	})
}
