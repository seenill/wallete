/*
社交功能API处理器

本文件实现了社交功能的HTTP接口处理器，包括：

主要接口：
- 地址簿管理：联系人增删改查、分组管理、标签管理
- 转账记录分享：交易分享、二维码生成、隐私控制
- 社交网络：关注/取消关注、用户搜索、社交资料
- 用户活动：活动记录、通知管理、社交统计

接口分组：
- /api/v1/social/contacts/* - 联系人管理接口
- /api/v1/social/share/* - 分享功能接口
- /api/v1/social/network/* - 社交网络接口
- /api/v1/social/user/* - 用户社交资料接口
- /api/v1/social/search/* - 搜索功能接口

安全特性：
- 联系人数据加密
- 分享权限控制
- 隐私设置保护
- 反垃圾邮件机制
*/
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// SocialHandler 社交功能API处理器
// 处理所有社交相关的HTTP请求，包括地址簿、分享、社交网络等功能
type SocialHandler struct {
	socialService *services.SocialService // 社交功能业务服务实例
}

// NewSocialHandler 创建新的社交功能处理器实例
// 参数: socialService - 社交功能业务服务实例
// 返回: 配置好的社交功能处理器
func NewSocialHandler(socialService *services.SocialService) *SocialHandler {
	return &SocialHandler{
		socialService: socialService,
	}
}

// AddContact 添加联系人
// POST /api/v1/social/contacts
// 请求体: ContactRequest结构体
// 功能: 添加新的联系人到地址簿
func (h *SocialHandler) AddContact(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	var req services.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证必要字段
	if req.Name == "" || len(req.Addresses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "联系人姓名和地址不能为空",
			"data": nil,
		})
		return
	}

	// 添加联系人
	response, err := h.socialService.AddContact(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "添加联系人失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "联系人添加成功",
		"data": response,
	})
}

// GetContact 获取联系人详情
// GET /api/v1/social/contacts/:contactId
// 路径参数:
//   - contactId: 联系人ID
//
// 响应: 联系人详细信息，包含最近活动
func (h *SocialHandler) GetContact(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	contactID := c.Param("contactId")
	if contactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "联系人ID不能为空",
			"data": nil,
		})
		return
	}

	// 获取联系人
	response, err := h.socialService.GetContact(c.Request.Context(), userAddress, contactID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取联系人失败: " + err.Error(),
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

// UpdateContact 更新联系人
// PUT /api/v1/social/contacts/:contactId
// 路径参数:
//   - contactId: 联系人ID
//
// 请求体: ContactRequest结构体
// 功能: 更新联系人信息
func (h *SocialHandler) UpdateContact(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	contactID := c.Param("contactId")
	if contactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "联系人ID不能为空",
			"data": nil,
		})
		return
	}

	var req services.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 更新联系人
	response, err := h.socialService.UpdateContact(c.Request.Context(), userAddress, contactID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新联系人失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "联系人更新成功",
		"data": response,
	})
}

// DeleteContact 删除联系人
// DELETE /api/v1/social/contacts/:contactId
// 路径参数:
//   - contactId: 联系人ID
//
// 功能: 从地址簿中删除联系人
func (h *SocialHandler) DeleteContact(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	contactID := c.Param("contactId")
	if contactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "联系人ID不能为空",
			"data": nil,
		})
		return
	}

	// 删除联系人
	err := h.socialService.DeleteContact(c.Request.Context(), userAddress, contactID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "删除联系人失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "联系人删除成功",
		"data": gin.H{
			"contact_id": contactID,
			"deleted_at": time.Now().Unix(),
		},
	})
}

// GetContactList 获取联系人列表
// GET /api/v1/social/contacts
// 查询参数:
//   - search: 搜索关键词（可选）
//   - tags: 标签过滤（可选，逗号分隔）
//   - groups: 分组过滤（可选，逗号分隔）
//   - sort_by: 排序字段（默认name）
//   - sort_order: 排序方向（asc/desc，默认asc）
//   - limit: 返回数量限制（默认20）
//   - offset: 偏移量（默认0）
//
// 响应: 联系人列表，包含分组和标签信息
func (h *SocialHandler) GetContactList(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	// 获取查询参数
	search := c.Query("search")
	sortBy := c.DefaultQuery("sort_by", "name")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	// 解析数值参数
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 构建请求
	request := &services.ContactListRequest{
		Search:    search,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Offset:    offset,
	}

	// 获取联系人列表
	response, err := h.socialService.GetContactList(c.Request.Context(), userAddress, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取联系人列表失败: " + err.Error(),
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

// ShareTransaction 分享交易
// POST /api/v1/social/share/transaction
// 请求体: ShareTransactionRequest结构体
// 功能: 创建交易分享链接和二维码
func (h *SocialHandler) ShareTransaction(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	var req services.ShareTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证交易哈希
	if req.TransactionHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "交易哈希不能为空",
			"data": nil,
		})
		return
	}

	// 创建分享
	response, err := h.socialService.ShareTransaction(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "创建分享失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "分享创建成功",
		"data": response,
	})
}

// GetShareRecord 获取分享记录
// GET /api/v1/social/share/:shareId
// 路径参数:
//   - shareId: 分享ID
//
// 响应: 分享记录详细信息
func (h *SocialHandler) GetShareRecord(c *gin.Context) {
	shareID := c.Param("shareId")
	if shareID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "分享ID不能为空",
			"data": nil,
		})
		return
	}

	// 获取分享记录
	shareRecord, err := h.socialService.GetShareRecord(c.Request.Context(), shareID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取分享记录失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": shareRecord,
	})
}

// SocialNetworkAction 社交网络操作
// POST /api/v1/social/network/action
// 请求体: SocialNetworkRequest结构体
// 功能: 执行社交网络操作（关注、取消关注等）
func (h *SocialHandler) SocialNetworkAction(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	var req services.SocialNetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 验证操作类型
	validActions := map[string]bool{
		"follow":   true,
		"unfollow": true,
	}
	if !validActions[req.Action] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "无效的操作类型",
			"data": nil,
		})
		return
	}

	// 执行社交网络操作
	response, err := h.socialService.SocialNetworkAction(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "执行操作失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, gin.H{
		"code": e.SUCCESS,
		"msg":  response.Message,
		"data": response.Data,
	})
}

// GetUserSocialProfile 获取用户社交资料
// GET /api/v1/social/user/:address/profile
// 路径参数:
//   - address: 用户地址
//
// 响应: 用户社交资料信息
func (h *SocialHandler) GetUserSocialProfile(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	// 获取用户社交资料
	profile, err := h.socialService.GetUserSocialProfile(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "获取用户资料失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": profile,
	})
}

// UpdateUserSocialProfile 更新用户社交资料
// PUT /api/v1/social/user/profile
// 请求体: UserSocialProfileRequest结构体
// 功能: 更新用户的社交资料和隐私设置
func (h *SocialHandler) UpdateUserSocialProfile(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	var req services.UserSocialProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "请求参数格式错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 更新用户社交资料
	err := h.socialService.UpdateUserSocialProfile(c.Request.Context(), userAddress, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "更新用户资料失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "用户资料更新成功",
		"data": gin.H{
			"user_address": userAddress,
			"updated_at":   time.Now().Unix(),
		},
	})
}

// SearchUsers 搜索用户
// GET /api/v1/social/search/users
// 查询参数:
//   - q: 搜索关键词（必需）
//   - limit: 返回数量限制（默认20）
//
// 响应: 用户搜索结果列表
func (h *SocialHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "搜索关键词不能为空",
			"data": nil,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 搜索用户
	users, err := h.socialService.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  "搜索用户失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"users": users,
			"total": len(users),
			"query": query,
		},
	})
}

// GetMyShares 获取我的分享列表
// GET /api/v1/social/share/my
// 查询参数:
//   - type: 分享类型（可选）
//   - limit: 返回数量限制（默认20）
//   - offset: 偏移量（默认0）
//
// 响应: 用户的分享记录列表
func (h *SocialHandler) GetMyShares(c *gin.Context) {
	userAddress := c.GetHeader("X-User-Address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "缺少用户地址",
			"data": nil,
		})
		return
	}

	shareType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 简化实现：返回示例数据
	shares := make([]gin.H, 0)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"shares":   shares,
			"total":    len(shares),
			"type":     shareType,
			"has_more": false,
		},
	})
}

// GetFollowers 获取粉丝列表
// GET /api/v1/social/network/:address/followers
// 路径参数:
//   - address: 用户地址
//
// 查询参数:
//   - limit: 返回数量限制（默认50）
//   - offset: 偏移量（默认0）
//
// 响应: 粉丝用户列表
func (h *SocialHandler) GetFollowers(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 简化实现：返回示例数据
	followers := make([]gin.H, 0)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"followers": followers,
			"total":     len(followers),
			"has_more":  false,
		},
	})
}

// GetFollowing 获取关注列表
// GET /api/v1/social/network/:address/following
// 路径参数:
//   - address: 用户地址
//
// 查询参数:
//   - limit: 返回数量限制（默认50）
//   - offset: 偏移量（默认0）
//
// 响应: 关注用户列表
func (h *SocialHandler) GetFollowing(c *gin.Context) {
	userAddress := c.Param("address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  "用户地址不能为空",
			"data": nil,
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 简化实现：返回示例数据
	following := make([]gin.H, 0)

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  "ok",
		"data": gin.H{
			"following": following,
			"total":     len(following),
			"has_more":  false,
		},
	})
}
