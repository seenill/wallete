package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"wallet/pkg/e"
)

// ErrorHandler 统一错误捕获与标准响应
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": e.ERROR,
					"msg":  e.GetMsg(e.ERROR),
					"data": rec,
				})
			}
		}()

		c.Next()

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
