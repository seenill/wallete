/*
错误处理中间件

本文件实现了统一的错误处理中间件，确保所有API返回统一格式的错误响应。

主要功能：
- 捕获和处理应用程panic
- 统一处理Gin框架的内部错误
- 统一错误响应格式
- 防止错误信息泄露敏感信息
*/
package middleware

import (
	"net/h
package middleware
import (
ub.com/gin-g
	"net/http"

	"github.com/gin-gonic/gin"
// ErrorHandler 统一错误捕获与标准响应中间件
// 功能:
// 1. 捕获并处理应用程序中的panic异常
// 2. 将Gin框架内部错误转换为统一格式的JSON响应
// 3. 确保所有错误都按照{code, msg, data}的格式返回
// 4. 防止错误堆栈信息泄露给客户端
// 使用: 应用于所有路由的全局中间件
)

unc {
	return func(c *gin.C
// ErrorHandler 统一错误捕获与标准响应
r捕获panic异常
		defer func() {
			if rec
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": e.ERROR,
 继续执行后续中间件和处理器
		c.Next()

		// 处理Gin框架中的内部错误
					"msg":  e.GetMsg(e.ERROR),
borted() {
			// 
					"data": rec,
				})
			}
		}()

		c.Next()

InternalServerError, gin.H{
		// gin.Context 内部错误转成统一格式
		if len(c.Errors) > 0 && !c.IsAborted() {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": e.ERROR,
				"msg":  e.GetMsg(e.ERROR),
				"data": c.Errors.String(),
			})
		}
	}
}