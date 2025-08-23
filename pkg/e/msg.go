package e

// MsgFlags 错误码对应的中文错误消息映射表
// 提供统一的错误消息管理，支持国际化和本地化
var MsgFlags = map[int]string{
	// 基础状态消息
	SUCCESS:       "ok",     // 操作成功
	ERROR:         "fail",   // 通用失败消息
	InvalidParams: "请求参数错误", // 请求参数格式错误或缺少必要参数

	// 认证和授权相关错误消息
	ErrorAuth:       "认证失败",   // JWT token无效或已过期
	ErrorPermission: "权限不足",   // 用户没有访问该资源的权限
	ErrorRateLimit:  "请求频率过高", // 超出了API调用限制

	// 钱包操作相关错误消息
	ErrorWalletCreate:         "创建钱包失败",         // 助记词生成或钱包初始化失败
	ErrorWalletGet:            "获取钱包信息失败",       // 查询钱包信息或地址失败
	ErrorWalletImport:         "导入钱包失败",         // 助记词或私钥格式错误
	ErrorWalletKeystore:       "钱包keystore处理失败", // Keystore文件加密或解密失败
	ErrorTransactionSend:      "发送交易失败",         // 交易广播到区块链失败
	ErrorTransactionBuild:     "构建交易失败",         // 交易参数错误或签名失败
	ErrorGetBalance:           "获取余额失败",         // 查询钱包或代币余额失败
	ErrorContractCall:         "调用合约失败",         // 智能合约调用执行失败
	ErrorInvalidPassword:      "钱包密码错误",         // 解锁钱包密码不正确
	ErrorWalletAddressInvalid: "无效的钱包地址",        // 地址格式不符合以太坊标准
	ErrorGasSuggestion:        "获取Gas建议失败",      // 交易费估算服务异常
	ErrorNonceGet:             "获取Nonce失败",      // 交易顺序号获取失败
	ErrorBroadcastRawTx:       "广播原始交易失败",       // 签名交易发送失败
}

// GetMsg 根据错误码获取对应的中文错误消息
// 参数: code - 错误码（对应code.go中定义的常量）
// 返回: 错误消息字符串，如果找不到对应消息则返回通用失败消息
// 用途: API响应中的msg字段，用于前端显示给用户
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	// 如果找不到对应的错误消息，返回通用失败消息
	return MsgFlags[ERROR]
}
