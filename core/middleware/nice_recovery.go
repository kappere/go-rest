package middleware

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"wataru.com/go-rest/core/logger"
	"wataru.com/go-rest/core/rest"
)

func NiceRecovery() rest.HandlerFunc {
	return func(c *rest.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				logger.Error("%v", err)
				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.JSON(http.StatusOK, rest.Error(err.(error).Error()))
					c.Abort()
				} else {
					pnc, e := err.(*rest.BizPanic)
					if e {
						c.JSON(http.StatusOK, rest.ErrorWithCode(fmt.Sprintf("%v", pnc), pnc.Code))
					} else {
						httpRequest, _ := httputil.DumpRequest(c.Request, false)
						headers := strings.Split(string(httpRequest), "\r\n")
						headersToStr := strings.Join(headers, "\r\n")
						logger.Raw("[Recovery] panic recovered: %v\n%s%s", err, headersToStr, debug.Stack())
						c.JSON(http.StatusOK, rest.Error("服务器异常，请联系管理员"))
					}
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}
