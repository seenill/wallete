/*
加密管理包

本包提供了钱包应用中需要的安全加密功能，包括：

主要功能：
- AES-GCM对称加密（高安全性、防篡改）
- Scrypt密钥派生（防彩虹表攻击）
- 随机盐值生成（增强安全性）
- 助记词加密存储（保护用户私钥信息）

技术特性：
- 使用AES-256-GCM算法，提供加密和认证
- Scrypt参数：N=32768, r=8, p=1，平衡安全性和性能
- 随机盐值防止彩虹表和字典攻击
- Base64和Hex编码支持，便于JSON序列化

安全注意事项：
- 主密码应该使用强密码策略
- 加密数据应该安全存储，避免明文传输
- 密钥派生过程可能耗时，建议在后台处理
*/
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

// EncryptedData 加密后的数据结构体
// 包含所有解密所需的信息，支持JSON序列化存储
type EncryptedData struct {
	Data  string `json:"data"`  // Base64编码的AES-GCM加密数据
	Salt  string `json:"salt"`  // 用于密钥派生的随机盐值（Hex编码）
	Nonce string `json:"nonce"` // AES-GCM加密的随机数（Hex编码）
}

// CryptoManager 加密管理器
// 提供统一的加密及解密服务，支持密码加密和默认密钥加密两种模式
type CryptoManager struct {
	defaultKey []byte // 主密码派生的默认加密密钥（AES-256）
}

// NewCryptoManager 创建加密管理器
func NewCryptoManager(masterPassword string) *CryptoManager {
	// 使用固定盐值派生默认密钥（实际生产环境应该使用更安全的方式）
	salt := []byte("wallet_default_salt_2024")
	key := deriveKey(masterPassword, salt)

	return &CryptoManager{
		defaultKey: key,
	}
}

// deriveKey 从密码派生密钥
func deriveKey(password string, salt []byte) []byte {
	// 使用scrypt进行密钥派生，参数: N=32768, r=8, p=1, keyLen=32
	key, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		// 如果scrypt失败，降级到SHA256
		hash := sha256.Sum256([]byte(password + string(salt)))
		return hash[:]
	}
	return key
}

// EncryptWithPassword 使用密码加密数据
func (cm *CryptoManager) EncryptWithPassword(plaintext, password string) (*EncryptedData, error) {
	// 生成随机盐值
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("生成盐值失败: %w", err)
	}

	// 派生密钥
	key := deriveKey(password, salt)

	// 使用派生的密钥加密
	encryptedData, nonce, err := cm.encryptAESGCM([]byte(plaintext), key)
	if err != nil {
		return nil, err
	}

	return &EncryptedData{
		Data:  base64.StdEncoding.EncodeToString(encryptedData),
		Salt:  hex.EncodeToString(salt),
		Nonce: hex.EncodeToString(nonce),
	}, nil
}

// DecryptWithPassword 使用密码解密数据
func (cm *CryptoManager) DecryptWithPassword(encData *EncryptedData, password string) (string, error) {
	// 解码盐值
	salt, err := hex.DecodeString(encData.Salt)
	if err != nil {
		return "", fmt.Errorf("解码盐值失败: %w", err)
	}

	// 派生密钥
	key := deriveKey(password, salt)

	// 解码数据
	ciphertext, err := base64.StdEncoding.DecodeString(encData.Data)
	if err != nil {
		return "", fmt.Errorf("解码加密数据失败: %w", err)
	}

	// 解码nonce
	nonce, err := hex.DecodeString(encData.Nonce)
	if err != nil {
		return "", fmt.Errorf("解码nonce失败: %w", err)
	}

	// 解密
	plaintext, err := cm.decryptAESGCM(ciphertext, nonce, key)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptDefault 使用默认密钥加密数据
func (cm *CryptoManager) EncryptDefault(plaintext string) (*EncryptedData, error) {
	encryptedData, nonce, err := cm.encryptAESGCM([]byte(plaintext), cm.defaultKey)
	if err != nil {
		return nil, err
	}

	return &EncryptedData{
		Data:  base64.StdEncoding.EncodeToString(encryptedData),
		Salt:  "", // 默认密钥不需要盐值
		Nonce: hex.EncodeToString(nonce),
	}, nil
}

// DecryptDefault 使用默认密钥解密数据
func (cm *CryptoManager) DecryptDefault(encData *EncryptedData) (string, error) {
	// 解码数据
	ciphertext, err := base64.StdEncoding.DecodeString(encData.Data)
	if err != nil {
		return "", fmt.Errorf("解码加密数据失败: %w", err)
	}

	// 解码nonce
	nonce, err := hex.DecodeString(encData.Nonce)
	if err != nil {
		return "", fmt.Errorf("解码nonce失败: %w", err)
	}

	// 解密
	plaintext, err := cm.decryptAESGCM(ciphertext, nonce, cm.defaultKey)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// encryptAESGCM 使用AES-GCM加密
func (cm *CryptoManager) encryptAESGCM(plaintext, key []byte) ([]byte, []byte, error) {
	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nonce, nil
}

// decryptAESGCM 使用AES-GCM解密
func (cm *CryptoManager) decryptAESGCM(ciphertext, nonce, key []byte) ([]byte, error) {
	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES cipher失败: %w", err)
	}

	// 创建GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	// 验证nonce长度
	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("无效的nonce长度")
	}

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	return plaintext, nil
}

// HashPassword 哈希密码（用于密码验证）
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword 验证密码哈希
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

// GenerateSecureKey 生成安全的随机密钥
func GenerateSecureKey(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %w", err)
	}

	return hex.EncodeToString(key), nil
}
