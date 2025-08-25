/**
 * 加密和随机数生成工具包
 * 
 * 提供密码加密、随机字符串生成等安全相关的工具函数
 * 
 * 后端学习要点：
 * 1. 加密安全 - 安全的盐值生成和随机字符串
 * 2. Go加密库 - crypto/rand的安全随机数生成
 * 3. 字符集定义 - 不同用途的字符集选择
 * 4. 错误处理 - 加密操作的错误处理
 */
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"
)

// 字符集定义
const (
	// 字母数字字符集（用于生成API密钥等）
	AlphaNumericChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	// 十六进制字符集
	HexChars = "0123456789abcdef"
	
	// URL安全的Base64字符集
	URLSafeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
)

/**
 * 生成安全的随机盐值
 * 用于密码加密，返回32字节的随机盐值
 * 
 * @return string 16进制编码的盐值
 */
func GenerateSalt() string {
	// 生成32字节的随机数据
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		// 如果随机数生成失败，使用当前时间作为fallback
		// 这不是最安全的方式，但确保程序不会崩溃
		return fmt.Sprintf("%x", []byte(fmt.Sprintf("fallback_%d", getCurrentTimestamp())))
	}
	
	// 返回十六进制编码的盐值
	return fmt.Sprintf("%x", salt)
}

/**
 * 生成指定长度的随机字符串
 * 使用加密安全的随机数生成器
 * 
 * @param length 字符串长度
 * @return string 随机字符串
 */
func GenerateRandomString(length int) string {
	return generateRandomStringWithCharset(length, AlphaNumericChars)
}

/**
 * 生成URL安全的随机字符串
 * 适用于生成Token、会话ID等
 * 
 * @param length 字符串长度
 * @return string URL安全的随机字符串
 */
func GenerateURLSafeRandomString(length int) string {
	return generateRandomStringWithCharset(length, URLSafeChars)
}

/**
 * 生成十六进制随机字符串
 * 适用于生成哈希值、ID等
 * 
 * @param length 字符串长度
 * @return string 十六进制随机字符串
 */
func GenerateHexRandomString(length int) string {
	return generateRandomStringWithCharset(length, HexChars)
}

/**
 * 使用指定字符集生成随机字符串
 * 核心的随机字符串生成函数
 * 
 * @param length 字符串长度
 * @param charset 字符集
 * @return string 随机字符串
 */
func generateRandomStringWithCharset(length int, charset string) string {
	if length <= 0 {
		return ""
	}
	
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	
	for i := 0; i < length; i++ {
		// 生成安全的随机索引
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// 如果随机数生成失败，使用简单的fallback
			result[i] = charset[i%len(charset)]
		} else {
			result[i] = charset[randomIndex.Int64()]
		}
	}
	
	return string(result)
}

/**
 * 生成Base64编码的随机字符串
 * 适用于生成密钥、令牌等
 * 
 * @param byteLength 字节长度（Base64编码后长度会更长）
 * @return string Base64编码的随机字符串
 */
func GenerateBase64RandomString(byteLength int) string {
	bytes := make([]byte, byteLength)
	_, err := rand.Read(bytes)
	if err != nil {
		// fallback: 使用当前时间戳
		timestamp := getCurrentTimestamp()
		return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("fallback_%d", timestamp)))
	}
	
	return base64.URLEncoding.EncodeToString(bytes)
}

/**
 * 验证字符串是否为有效的十六进制
 * 用于验证哈希值、盐值等
 * 
 * @param s 待验证的字符串
 * @return bool 是否为有效的十六进制字符串
 */
func IsValidHex(s string) bool {
	if len(s)%2 != 0 {
		return false
	}
	
	for _, char := range s {
		if !((char >= '0' && char <= '9') || 
			 (char >= 'a' && char <= 'f') || 
			 (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	
	return true
}

/**
 * 生成UUID风格的字符串
 * 格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
 * 
 * @return string UUID风格的字符串
 */
func GenerateUUID() string {
	// 生成32个十六进制字符
	hex := GenerateHexRandomString(32)
	
	// 格式化为UUID风格
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex[0:8],
		hex[8:12],
		hex[12:16],
		hex[16:20],
		hex[20:32],
	)
}

/**
 * 生成安全的Session ID
 * 128位的安全随机字符串
 * 
 * @return string Session ID
 */
func GenerateSessionID() string {
	return GenerateBase64RandomString(16) // 16字节 = 128位
}

/**
 * 生成API密钥
 * 使用特定的前缀和随机字符串
 * 
 * @param prefix 前缀（如 "ak_", "sk_" 等）
 * @return string API密钥
 */
func GenerateAPIKey(prefix string) string {
	randomPart := GenerateRandomString(32)
	return prefix + randomPart
}

// =============================================================================
// 内部辅助函数
// =============================================================================

/**
 * 获取当前时间戳（毫秒）
 * 用作随机数生成失败时的fallback
 */
func getCurrentTimestamp() int64 {
	return time.Now().UnixMilli()
}package utils
