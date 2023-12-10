package middleware

import (
	"github.com/gin-gonic/gin"
)

// BasicAuth 权限校验 authFunc鉴权函数，返回值：是否有权限
func BasicAuth(authFunc func(*gin.Context) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authFunc(c) {
			c.Abort()
			return
		}
		c.Next()
	}
}
