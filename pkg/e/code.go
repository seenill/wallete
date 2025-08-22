package e

// 定义业务状态码
const (
	SUCCESS       = 200
	ERROR         = 500
	InvalidParams = 400

	ErrorWalletCreate         = 10001
	ErrorWalletGet            = 10002
	ErrorWalletImport         = 10003
	ErrorWalletKeystore       = 10004
	ErrorTransactionSend      = 10005
	ErrorTransactionBuild     = 10006
	ErrorGetBalance           = 10007
	ErrorContractCall         = 10008
	ErrorInvalidPassword      = 10009
	ErrorWalletAddressInvalid = 10010
)
