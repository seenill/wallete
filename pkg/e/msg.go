package e

var MsgFlags = map[int]string{
	SUCCESS:       "ok",
	ERROR:         "fail",
	InvalidParams: "请求参数错误",

	ErrorWalletCreate:         "创建钱包失败",
	ErrorWalletGet:            "获取钱包信息失败",
	ErrorWalletImport:         "导入钱包失败",
	ErrorWalletKeystore:       "钱包keystore处理失败",
	ErrorTransactionSend:      "发送交易失败",
	ErrorTransactionBuild:     "构建交易失败",
	ErrorGetBalance:           "获取余额失败",
	ErrorContractCall:         "调用合约失败",
	ErrorInvalidPassword:      "钱包密码错误",
	ErrorWalletAddressInvalid: "无效的钱包地址",
}

// GetMsg 根据代码获取错误信息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
