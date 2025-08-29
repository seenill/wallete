/*
1inch聚合器服务

本文件实现了与1inch API的集成，提供以下功能：
1. 价格报价查询
2. 交易路由优化
3. 交易执行
4. 代币信息查询

1inch API端点：
- Quote: https://api.1inch.dev/swap/v5.2/{chain_id}/quote
- Swap: https://api.1inch.dev/swap/v5.2/{chain_id}/swap
- Tokens: https://api.1inch.dev/token/v1.2/{chain_id}
- Liquidity sources: https://api.1inch.dev/swap/v5.2/{chain_id}/liquidity-sources
*/
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OneInchService 1inch聚合器服务
type OneInchService struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	chainID    int64
}

// OneInchQuoteRequest 1inch报价请求
type OneInchQuoteRequest struct {
	FromTokenAddress string `json:"fromTokenAddress"`
	ToTokenAddress   string `json:"toTokenAddress"`
	Amount           string `json:"amount"`
	GasPrice         string `json:"gasPrice,omitempty"`
}

// OneInchQuoteResponse 1inch报价响应
type OneInchQuoteResponse struct {
	FromToken       *OneInchToken `json:"fromToken"`
	ToToken         *OneInchToken `json:"toToken"`
	FromTokenAmount string        `json:"fromTokenAmount"`
	ToTokenAmount   string        `json:"toTokenAmount"`
	EstimatedGas    int64         `json:"estimatedGas"`
	GasPrice        string        `json:"gasPrice"`
	Protocols       interface{}   `json:"protocols"`
	Tx              *OneInchTx    `json:"tx"`
}

// OneInchSwapRequest 1inch交换请求
type OneInchSwapRequest struct {
	FromTokenAddress string `json:"fromTokenAddress"`
	ToTokenAddress   string `json:"toTokenAddress"`
	Amount           string `json:"amount"`
	FromAddress      string `json:"fromAddress"`
	Slippage         string `json:"slippage"`
	GasPrice         string `json:"gasPrice,omitempty"`
	DisableEstimate  bool   `json:"disableEstimate,omitempty"`
	AllowPartialFill bool   `json:"allowPartialFill,omitempty"`
}

// OneInchSwapResponse 1inch交换响应
type OneInchSwapResponse struct {
	FromToken       *OneInchToken `json:"fromToken"`
	ToToken         *OneInchToken `json:"toToken"`
	FromTokenAmount string        `json:"fromTokenAmount"`
	ToTokenAmount   string        `json:"toTokenAmount"`
	Protocols       interface{}   `json:"protocols"`
	Tx              *OneInchTx    `json:"tx"`
}

// OneInchToken 1inch代币信息
type OneInchToken struct {
	Address  string   `json:"address"`
	Symbol   string   `json:"symbol"`
	Name     string   `json:"name"`
	Decimals int      `json:"decimals"`
	LogoURI  string   `json:"logoURI"`
	Tags     []string `json:"tags"`
}

// OneInchTx 1inch交易信息
type OneInchTx struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	GasPrice string `json:"gasPrice"`
	Gas      int64  `json:"gas"`
}

// TokenListResponse 代币列表响应
type TokenListResponse map[string]*OneInchToken

// GetAPIKey 获取API密钥
func (s *OneInchService) GetAPIKey() string {
	return s.apiKey
}

// GetChainID 获取链ID
func (s *OneInchService) GetChainID() int64 {
	return s.chainID
}

// GetHTTPClient 获取HTTP客户端
func (s *OneInchService) GetHTTPClient() *http.Client {
	return s.httpClient
}

// NewOneInchService 创建1inch服务实例
func NewOneInchService(apiKey string) *OneInchService {
	return &OneInchService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.1inch.dev",
		chainID: 1, // Ethereum主网
	}
}

// GetQuote 获取交易报价
func (s *OneInchService) GetQuote(ctx context.Context, req *OneInchQuoteRequest) (*OneInchQuoteResponse, error) {
	// 构建查询参数
	params := url.Values{}
	params.Add("fromTokenAddress", req.FromTokenAddress)
	params.Add("toTokenAddress", req.ToTokenAddress)
	params.Add("amount", req.Amount)

	if req.GasPrice != "" {
		params.Add("gasPrice", req.GasPrice)
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/swap/v5.2/%d/quote?%s", s.baseURL, s.chainID, params.Encode())

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var quoteResp OneInchQuoteResponse
	if err := json.Unmarshal(body, &quoteResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &quoteResp, nil
}

// GetSwap 获取交换交易数据
func (s *OneInchService) GetSwap(ctx context.Context, req *OneInchSwapRequest) (*OneInchSwapResponse, error) {
	// 构建查询参数
	params := url.Values{}
	params.Add("fromTokenAddress", req.FromTokenAddress)
	params.Add("toTokenAddress", req.ToTokenAddress)
	params.Add("amount", req.Amount)
	params.Add("fromAddress", req.FromAddress)
	params.Add("slippage", req.Slippage)

	if req.GasPrice != "" {
		params.Add("gasPrice", req.GasPrice)
	}

	if req.DisableEstimate {
		params.Add("disableEstimate", "true")
	}

	if req.AllowPartialFill {
		params.Add("allowPartialFill", "true")
	}

	// 构建请求URL
	url := fmt.Sprintf("%s/swap/v5.2/%d/swap?%s", s.baseURL, s.chainID, params.Encode())

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var swapResp OneInchSwapResponse
	if err := json.Unmarshal(body, &swapResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &swapResp, nil
}

// GetTokens 获取代币列表
func (s *OneInchService) GetTokens(ctx context.Context) (*TokenListResponse, error) {
	// 构建请求URL
	url := fmt.Sprintf("%s/token/v1.2/%d", s.baseURL, s.chainID)

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var tokens TokenListResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokens, nil
}

// FormatAmount 格式化代币数量（将人类可读格式转换为最小单位）
func (s *OneInchService) FormatAmount(amount float64, decimals int) *big.Int {
	// 将amount乘以10^decimals
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountInt := new(big.Int).SetInt64(int64(amount * float64(multiplier.Int64())))
	return amountInt
}

// ParseAmount 解析代币数量（将最小单位转换为人类可读格式）
func (s *OneInchService) ParseAmount(amount *big.Int, decimals int) float64 {
	// 将amount除以10^decimals
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountFloat, _ := new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(divisor)).Float64()
	return amountFloat
}

// GetTokenBySymbol 根据符号获取代币信息
func (s *OneInchService) GetTokenBySymbol(ctx context.Context, symbol string) (*OneInchToken, error) {
	tokens, err := s.GetTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	// 查找匹配的代币
	for _, token := range *tokens {
		if strings.ToUpper(token.Symbol) == strings.ToUpper(symbol) {
			return token, nil
		}
	}

	return nil, fmt.Errorf("token with symbol %s not found", symbol)
}

// CalculatePriceImpact 计算价格影响
func (s *OneInchService) CalculatePriceImpact(fromAmount, toAmount *big.Int, fromPrice, toPrice float64) float64 {
	// 计算预期价值
	expectedValue := new(big.Float).Mul(new(big.Float).SetInt(fromAmount), big.NewFloat(fromPrice))
	actualValue := new(big.Float).Mul(new(big.Float).SetInt(toAmount), big.NewFloat(toPrice))

	// 计算价格影响
	priceImpact := new(big.Float).Quo(
		new(big.Float).Sub(expectedValue, actualValue),
		expectedValue,
	)

	priceImpactPercent, _ := priceImpact.Float64()
	return priceImpactPercent * 100
}
