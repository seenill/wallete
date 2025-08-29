/*
错误码管理包

本包定义了钱包服务中所有的错误码和相应的错误消息。

错误码分类：
- 2xx: 成功状态码
- 4xx: 客户端错误（请求参数、认证等）
- 5xx: 服务端错误
- 10xxx: 业务错误码（钱包操作、交易处理等）

使用方式：
- API接口返回统一的错误码和消息
- 前端可根据错误码进行本地化显示
- 日志系统可按错误码进行统计和告警
*/
package e

// 定义业务状态码
// 遵循TTP标准和业务规范，确保错误码的一致性和可读性
const (
	// 成功状态码
	SUCCESS = 200 // 请求成功

	// 通用错误码
	ERROR         = 500 // 服务器内部错误
	InvalidParams = 400 // 请求参数错误

	// 认证和授权相关错误码
	ErrorAuth       = 401 // 认证失败（未登录或token无效）
	ErrorPermission = 403 // 权限不足（无权访问资源）
	ErrorRateLimit  = 429 // 请求频率过高（被限流）

	// 钱包相关业务错误码 (10xxx系列)
	ErrorWalletCreate         = 10001 // 创建钱包失败
	ErrorWalletGet            = 10002 // 获取钱包信息失败
	ErrorWalletImport         = 10003 // 导入钱包失败（助记词或私钥错误）
	ErrorWalletKeystore       = 10004 // 钱包keystore处理失败
	ErrorTransactionSend      = 10005 // 发送交易失败
	ErrorTransactionBuild     = 10006 // 构建交易失败
	ErrorGetBalance           = 10007 // 获取余额失败
	ErrorContractCall         = 10008 // 调用智能合约失败
	ErrorInvalidPassword      = 10009 // 钱包密码错误
	ErrorWalletAddressInvalid = 10010 // 无效的钱包地址格式
	ErrorGasSuggestion        = 10011 // 获取Gas建议失败
	ErrorNonceGet             = 10012 // 获取Nonce失败
	ErrorBroadcastRawTx       = 10013 // 广播原始交易失败
	ErrorDeFiOperation        = 10014 // DeFi操作失败
)
