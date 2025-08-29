/*
1inch API处理器

本文件实现了1inch聚合器API的HTTP处理器，提供以下功能：
1. 交易报价查询
2. 交易执行
3. 代币信息查询
4. 流动性源查询
*/
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wallet/pkg/e"
	"wallet/services"

	"github.com/gin-gonic/gin"
)

// OneInchHandler 1inch API处理器
type OneInchHandler struct {
	defiService *services.DeFiService
}

// NewOneInchHandler 创建1inch处理器实例
func NewOneInchHandler(defiService *services.DeFiService) *OneInchHandler {
	return &OneInchHandler{
		defiService: defiService,
	}
}

// OneInchQuoteRequest 1inch报价请求
type OneInchQuoteRequest struct {
	FromTokenAddress string `json:"fromTokenAddress" binding:"required"`
	ToTokenAddress   string `json:"toTokenAddress" binding:"required"`
	Amount           string `json:"amount" binding:"required"`
	Slippage         string `json:"slippage"`
	GasPrice         string `json:"gasPrice"`
}

// GetQuote 获取1inch交易报价
// @Summary 获取1inch交易报价
// @Description 获取两个代币之间的最优交易报价
// @Tags DeFi
// @Accept json
// @Produce json
// @Param fromTokenAddress query string true "输入代币地址"
// @Param toTokenAddress query string true "输出代币地址"
// @Param amount query string true "输入代币数量（最小单位）"
// @Param slippage query string false "滑点容忍度（百分比，默认1）"
// @Param gasPrice query string false "Gas价格（wei）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/defi/oneinch/quote [get]
func (h *OneInchHandler) GetQuote(c *gin.Context) {
	// 从查询参数获取数据
	fromTokenAddress := c.Query("fromTokenAddress")
	toTokenAddress := c.Query("toTokenAddress")
	amount := c.Query("amount")
	slippage := c.Query("slippage")
	gasPrice := c.Query("gasPrice")

	// 验证必需参数
	if fromTokenAddress == "" || toTokenAddress == "" || amount == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "fromTokenAddress, toTokenAddress and amount are required",
		})
		return
	}

	// 构建交换请求
	swapReq := &services.SwapRequest{
		TokenIn:  fromTokenAddress,
		TokenOut: toTokenAddress,
		AmountIn: amount,
		Slippage: slippage,
		GasPrice: gasPrice,
		// 注意：这里缺少UserAddress和Deadline，因为这是报价查询
	}

	// 获取报价
	quote, err := h.defiService.GetSwapQuote(swapReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": quote,
	})
}

// OneInchSwapRequest 1inch交换请求
type OneInchSwapRequest struct {
	FromTokenAddress string `json:"fromTokenAddress" binding:"required"`
	ToTokenAddress   string `json:"toTokenAddress" binding:"required"`
	Amount           string `json:"amount" binding:"required"`
	FromAddress      string `json:"fromAddress" binding:"required"`
	Slippage         string `json:"slippage"`
	GasPrice         string `json:"gasPrice"`
}

// GetSwap 获取1inch交换交易数据
// @Summary 获取1inch交换交易数据
// @Description 获取用于执行代币交换的交易数据
// @Tags DeFi
// @Accept json
// @Produce json
// @Param fromTokenAddress query string true "输入代币地址"
// @Param toTokenAddress query string true "输出代币地址"
// @Param amount query string true "输入代币数量（最小单位）"
// @Param fromAddress query string true "发送方地址"
// @Param slippage query string false "滑点容忍度（百分比，默认1）"
// @Param gasPrice query string false "Gas价格（wei）"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object]interface{}
// @Router /api/v1/defi/oneinch/swap [get]
func (h *OneInchHandler) GetSwap(c *gin.Context) {
	// 从查询参数获取数据
	fromTokenAddress := c.Query("fromTokenAddress")
	toTokenAddress := c.Query("toTokenAddress")
	amount := c.Query("amount")
	fromAddress := c.Query("fromAddress")
	slippage := c.Query("slippage")
	gasPrice := c.Query("gasPrice")

	// 验证必需参数
	if fromTokenAddress == "" || toTokenAddress == "" || amount == "" || fromAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "fromTokenAddress, toTokenAddress, amount and fromAddress are required",
		})
		return
	}

	// 构建交换请求
	swapReq := &services.SwapRequest{
		TokenIn:     fromTokenAddress,
		TokenOut:    toTokenAddress,
		AmountIn:    amount,
		Slippage:    slippage,
		UserAddress: fromAddress,
		GasPrice:    gasPrice,
		// Deadline设置为当前时间+20分钟
		Deadline: 0,
	}

	// 执行交换
	result, err := h.defiService.ExecuteSwap(swapReq, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": result,
	})
}

// GetTokens 获取支持的代币列表
// @Summary 获取支持的代币列表
// @Description 获取1inch支持的所有代币信息
// @Tags DeFi
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/defi/oneinch/tokens [get]
func (h *OneInchHandler) GetTokens(c *gin.Context) {
	// 直接调用1inch服务获取代币列表
	oneInchService := h.defiService.GetOneInchService()
	if oneInchService == nil || oneInchService.GetAPIKey() == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "1inch API key not configured",
		})
		return
	}

	tokens, err := oneInchService.GetTokens(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": tokens,
	})
}

// GetLiquiditySources 获取流动性源
// @Summary 获取流动性源
// @Description 获取1inch支持的所有流动性源
// @Tags DeFi
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/defi/oneinch/liquidity-sources [get]
func (h *OneInchHandler) GetLiquiditySources(c *gin.Context) {
	// 直接调用1inch服务获取流动性源
	oneInchService := h.defiService.GetOneInchService()
	if oneInchService == nil || oneInchService.GetAPIKey() == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": e.InvalidParams,
			"msg":  e.GetMsg(e.InvalidParams),
			"data": "1inch API key not configured",
		})
		return
	}

	// 构建请求URL
	url := fmt.Sprintf("https://api.1inch.dev/swap/v5.2/%d/liquidity-sources", oneInchService.GetChainID())

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": fmt.Sprintf("failed to create request: %v", err),
		})
		return
	}

	// 设置请求头
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oneInchService.GetAPIKey()))
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := oneInchService.GetHTTPClient().Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": fmt.Sprintf("failed to send request: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": fmt.Sprintf("failed to read response body: %v", err),
		})
		return
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode, string(body)),
		})
		return
	}

	var liquiditySources interface{}
	if err := json.Unmarshal(body, &liquiditySources); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": e.ERROR,
			"msg":  e.GetMsg(e.ERROR),
			"data": fmt.Sprintf("failed to parse response: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": e.SUCCESS,
		"msg":  e.GetMsg(e.SUCCESS),
		"data": liquiditySources,
	})
}
