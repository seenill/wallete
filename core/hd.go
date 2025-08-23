/*
HD钱包核心功能包

本包实现了分层确定性（HD）钱包的核心功能，包括：
- BIP39助记词生成和验证
- BIP44地址派生（支持以太坊和其他EVM兼容链）
- 私钥和地址管理
- 批量地址生成

支持的派生路径格式：
m/44'/60'/0'/0/0 - 以太坊主网标准路径
m/44'/60'/0'/0/1 - 以太坊第二个地址
其中 44' 是BIP44约定，60' 是以太坊的coin_type
*/
package core

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	bip39 "github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic 生成BIP39标准的助记词
// 参数: strength - 熟值强度（128位生成12个单词，256位生成24个单词）
// 返回: 助记词字符串和错误信息
// 注意: 助记词是钱包恢复的唯一凭证，必须安全保存
func GenerateMnemonic(strength int) (string, error) {
	if strength != 128 && strength != 256 {
		strength = 128
	}
	entropy, err := bip39.NewEntropy(strength)
	if err != nil {
		return "", fmt.Errorf("生成熵失败: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("生成助记词失败: %w", err)
	}
	return mnemonic, nil
}

// DeriveAddressFromMnemonic 从助记词和派生路径生成地址
// 参数:
//
//	mnemonic - BIP39助记词（用空格分隔的单词）
//	derivationPath - BIP44派生路径（例如: m/44'/60'/0'/0/0）
//
// 返回: 以太坊地址字符串（0x开头）和错误信息
// 用途: 用于显示地址或验证钱包可访问性
func DeriveAddressFromMnemonic(mnemonic, derivationPath string) (string, error) {
	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", fmt.Errorf("根据助记词创建钱包失败: %w", err)
	}

	path, err := hdwallet.ParseDerivationPath(derivationPath)
	if err != nil {
		return "", fmt.Errorf("解析派生路径失败: %w", err)
	}

	account, err := w.Derive(path, true)
	if err != nil {
		return "", fmt.Errorf("派生账户失败: %w", err)
	}

	addr := account.Address.Hex()
	return addr, nil
}

// DerivePrivateKeyFromMnemonic 从助记词和派生路径生成私钥和地址
// 参数:
//
//	mnemonic - BIP39助记词
//	derivationPath - BIP44派生路径
//
// 返回: ECDSA私钥、以太坊地址和错误信息
// 用途: 用于交易签名，私钥需要安全处理
// 警告: 私钥有超级权限，不可泄露给第三方
func DerivePrivateKeyFromMnemonic(mnemonic, derivationPath string) (*ecdsa.PrivateKey, common.Address, error) {
	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("根据助记词创建钱包失败: %w", err)
	}

	path, err := hdwallet.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("解析派生路径失败: %w", err)
	}

	account, err := w.Derive(path, true)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("派生账户失败: %w", err)
	}

	priv, err := w.PrivateKey(account)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("导出私钥失败: %w", err)
	}

	return priv, account.Address, nil
}

// DeriveAddressesFromMnemonic 从助记词批量生成多个地址
// 参数:
//
//	mnemonic - BIP39助记词
//	pathPrefix - 派生路径前缀（例如: "m/44'/60'/0'/0"）
//	start - 起始索引（从0开始）
//	count - 生成地址数量（必须>0）
//
// 返回: 地址数组和错误信息
// 用途: 为用户显示多个地址选项，或批量导入地址
// 注意: 最终派生路径 = pathPrefix + "/{start+i}"，i从0到count-1
func DeriveAddressesFromMnemonic(mnemonic, pathPrefix string, start, count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count 必须大于 0")
	}
	if pathPrefix == "" {
		pathPrefix = "m/44'/60'/0'/0"
	}

	w, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("根据助记词创建钱包失败: %w", err)
	}

	addresses := make([]string, 0, count)
	for i := 0; i < count; i++ {
		fullPath := fmt.Sprintf("%s/%d", pathPrefix, start+i)
		path, err := hdwallet.ParseDerivationPath(fullPath)
		if err != nil {
			return nil, fmt.Errorf("解析派生路径失败(%s): %w", fullPath, err)
		}
		account, err := w.Derive(path, true)
		if err != nil {
			return nil, fmt.Errorf("派生账户失败(%s): %w", fullPath, err)
		}
		addresses = append(addresses, account.Address.Hex())
	}
	return addresses, nil
}
