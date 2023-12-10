package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/httpx"
)

func NiceRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
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
				slog.Error("Nice recovery.", "error", err)
				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					c.JSON(http.StatusOK, httpx.Error(err.(error).Error()))
					c.Abort()
				} else {
					httpRequest, _ := httputil.DumpRequest(c.Request, false)
					headers := strings.Split(string(httpRequest), "\r\n")
					headersToStr := strings.Join(headers, "\r\n")
					slog.Error("[Recovery] Panic recovered.")
					slog.Error(headersToStr)
					slog.Error(string(debug.Stack()))
					c.JSON(http.StatusOK, httpx.Error("服务器异常，请联系管理员"))
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}
