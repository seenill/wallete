package core

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"path/filepath"
	"strings"
)

// CreateKeystore 在指定的目录中创建一个新的加密keystore文件
// password: 用于加密钱包的密码
// keystoreDir: 存放keystore文件的目录
// 返回: 新创建的账户对象和可能发生的错误
func CreateKeystore(password string, keystoreDir string) (*accounts.Account, error) {
	// 确保keystore目录存在
	if err := os.MkdirAll(keystoreDir, 0700); err != nil {
		return nil, fmt.Errorf("创建keystore目录失败: %w", err)
	}

	// 创建一个新的keystore服务
	ks := keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// 使用密码创建一个新账户
	account, err := ks.NewAccount(password)
	if err != nil {
		return nil, fmt.Errorf("创建新账户失败: %w", err)
	}

	fmt.Printf("成功创建新钱包账户, 地址: %s\n", account.Address.Hex())
	return &account, nil
}

// GetKeystoreFilePath 根据地址查找keystore文件路径
func GetKeystoreFilePath(addressHex string, keystoreDir string) (string, error) {
	files, err := os.ReadDir(keystoreDir)
	if err != nil {
		return "", fmt.Errorf("读取keystore目录失败: %w", err)
	}

	// keystore文件名中的地址部分是小写的，并且没有 "0x" 前缀
	addressPart := strings.ToLower(strings.TrimPrefix(addressHex, "0x"))

	for _, file := range files {
		if !file.IsDir() {
			// 通过检查文件名是否包含地址部分来查找文件
			// 典型的keystore文件名: UTC--2023-11-21T04-39-50.998Z--a0b...
			if strings.Contains(strings.ToLower(file.Name()), addressPart) {
				return filepath.Join(keystoreDir, file.Name()), nil
			}
		}
	}

	return "", fmt.Errorf("未找到地址 %s 对应的keystore文件", addressHex)
}