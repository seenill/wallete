package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int           // 请求次数限制
	window   time.Duration // 时间窗口
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 获取用户的请求记录
	requests, exists := rl.requests[key]
	if !exists {
		rl.requests[key] = []time.Time{now}
		return true
	}

	// 清理过期的请求记录
	validRequests := make([]time.Time, 0, len(requests))
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// 检查是否超过限制
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	// 添加当前请求
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

// GetRetryAfter 获取重试等待时间
func (rl *RateLimiter) GetRetryAfter(key string) time.Duration {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	requests, exists := rl.requests[key]
	if !exists || len(requests) == 0 {
		return 0
	}

	// 找到最早的请求时间
	earliestRequest := requests[0]
	for _, reqTime := range requests[1:] {
		if reqTime.Before(earliestRequest) {
			earliestRequest = reqTime
		}
	}

	// 计算需要等待的时间
	waitUntil := earliestRequest.Add(rl.window)
	if time.Now().Before(waitUntil) {
		return waitUntil.Sub(time.Now())
	}

	return 0
}

// CleanupOldRequests 清理过期的请求记录
func (rl *RateLimiter) CleanupOldRequests() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window * 2) // 清理2倍时间窗口之前的记录

	for key, requests := range rl.requests {
		validRequests := make([]time.Time, 0, len(requests))
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// 全局速率限制器实例
var (
	generalLimiter     *RateLimiter // 通用API限制
	transactionLimiter *RateLimiter // 交易API限制
	authLimiter        *RateLimiter // 认证API限制
)

// InitRateLimiters 初始化速率限制器
func InitRateLimiters() {
	// 通用API：每分钟300个请求（增加限制以减少429错误）
	generalLimiter = NewRateLimiter(300, time.Minute)

	// 交易API：每分钟10个请求（更严格）
	transactionLimiter = NewRateLimiter(10, time.Minute)

	// 认证API：每分钟5个请求
	authLimiter = NewRateLimiter(5, time.Minute)

	// 启动清理协程
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			generalLimiter.CleanupOldRequests()
			transactionLimiter.CleanupOldRequests()
			authLimiter.CleanupOldRequests()
		}
	}()
}

// getClientKey 获取客户端标识
func getClientKey(c *gin.Context) string {
	// 优先使用认证用户ID
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%s", userID)
	}

	// 其次使用API密钥
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("api:%s", apiKey)
	}

	// 最后使用IP地址
	clientIP := c.ClientIP()
	return fmt.Sprintf("ip:%s", clientIP)
}

// RateLimit 通用速率限制中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if generalLimiter == nil {
			c.Next()
			return
		}

		key := getClientKey(c)
		if !generalLimiter.Allow(key) {
			retryAfter := generalLimiter.GetRetryAfter(key)
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": e.ErrorRateLimit,
				"msg":  e.GetMsg(e.ErrorRateLimit),
				"data": fmt.Sprintf("请求过于频繁，请 %d 秒后重试", int(retryAfter.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TransactionRateLimit 交易接口速率限制中间件
func TransactionRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if transactionLimiter == nil {
			c.Next()
			return
		}

		key := getClientKey(c)
		if !transactionLimiter.Allow(key) {
			retryAfter := transactionLimiter.GetRetryAfter(key)
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": e.ErrorRateLimit,
				"msg":  e.GetMsg(e.ErrorRateLimit),
				"data": fmt.Sprintf("交易请求过于频繁，请 %d 秒后重试", int(retryAfter.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimit 认证接口速率限制中间件
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authLimiter == nil {
			c.Next()
			return
		}

		key := getClientKey(c)
		if !authLimiter.Allow(key) {
			retryAfter := authLimiter.GetRetryAfter(key)
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code": e.ErrorRateLimit,
				"msg":  e.GetMsg(e.ErrorRateLimit),
				"data": fmt.Sprintf("认证请求过于频繁，请 %d 秒后重试", int(retryAfter.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TransactionValidation 交易验证中间件
func TransactionValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对交易相关的接口进行验证
		if !isTransactionEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 验证请求头
		if c.GetHeader("Content-Type") != "application/json" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": e.InvalidParams,
				"msg":  "Content-Type must be application/json",
				"data": nil,
			})
			c.Abort()
			return
		}

		// 验证请求大小（防止DOS攻击）
		if c.Request.ContentLength > 1024*1024 { // 1MB限制
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"code": e.InvalidParams,
				"msg":  "Request body too large",
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isTransactionEndpoint 检查是否为交易相关接口
func isTransactionEndpoint(path string) bool {
	transactionPaths := []string{
		"/api/v1/transactions/send",
		"/api/v1/transactions/send-erc20",
		"/api/v1/transactions/send-advanced",
		"/api/v1/transactions/send-erc20-advanced",
		"/api/v1/transactions/broadcast",
		"/api/v1/tokens/",
	}

	for _, txPath := range transactionPaths {
		if strings.Contains(path, txPath) {
			return true
		}
	}

	return false
}

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-API-Key,X-User-Address")
		c.Header("Access-Control-Expose-Headers", "Content-Length,X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求ID
			requestID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), c.ClientIP())
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}
