package middleware

import (
	"net/http"

	"github.com/kappere/go-rest/core/rest"
)

// AuthorizationMiddleware 权限校验 authFunc鉴权函数，返回值：是否有权限，failCallback无权限时的回调
func AuthorizationMiddleware(authFunc func(*rest.Context) bool, failCallback func(*rest.Context)) rest.HandlerFunc {
	return func(c *rest.Context) {
		if !authFunc(c) {
			if failCallback != nil {
				failCallback(c)
			} else {
				c.JSON(http.StatusOK, rest.ErrorWithCode("no authorization", rest.STATUS_NO_AUTHORIZATION))
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
