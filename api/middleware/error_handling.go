package middleware

import (
	"fmt"
	"net/http"
	"wallet/pkg/e"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 是一个全局的错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var code int
				var msg string

				// 检查是否是我们自定义的错误类型
				if customErr, ok := err.(error); ok {
					// 这里可以根据不同的错误类型进行更精细的处理
					// 为了简单起见，我们先统一处理
					code = e.ERROR
					msg = customErr.Error()
				} else {
					// 处理其他类型的panic
					code = http.StatusInternalServerError
					msg = fmt.Sprintf("%v", err)
				}

				c.JSON(http.StatusOK, gin.H{
					"code": code,
					"msg":  e.GetMsg(code),
					"data": msg,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}