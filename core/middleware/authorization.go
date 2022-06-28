package middleware

import (
	"github.com/kappere/go-rest/core/rest"
)

// BasicAuth 权限校验 authFunc鉴权函数，返回值：是否有权限
func BasicAuth(authFunc func(*rest.Context) bool) rest.HandlerFunc {
	return func(c *rest.Context) {
		if !authFunc(c) {
			c.Abort()
			return
		}
		c.Next()
	}
}
