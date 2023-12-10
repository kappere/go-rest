package middleware

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

var hostname, _ = os.Hostname()

type LogFormatterParams struct {
	gin.LogFormatterParams
	RequestId string
	Debug     bool
}

var niceLogFormatter = func(param LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	var statusColor, methodColor, resetColor string
	if param.Debug {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()

		return fmt.Sprintf("[%s|%s|%s] %s %d %s %s %s %s %s %s %s %s\n",
			hostname,
			param.RequestId,
			param.ClientIP,
			statusColor, param.StatusCode, resetColor,
			methodColor, param.Method, resetColor,
			param.Path,
			param.Request.Proto,
			param.Latency,
			param.ErrorMessage,
		)
	}

	return fmt.Sprintf("[%s|%s|%s] %d %s %s %s %s %s\n",
		hostname,
		param.RequestId,
		param.ClientIP,
		param.StatusCode,
		param.Method,
		param.Path,
		param.Request.Proto,
		param.Latency,
		param.ErrorMessage,
	)
}

// NiceLoggerFormatter 更好的日志中间件
func NiceLoggerFormatter(formatter func(params LogFormatterParams) string, debug bool) gin.HandlerFunc {
	conf := gin.LoggerConfig{}
	if formatter == nil {
		formatter = niceLogFormatter
	}

	notlogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := LogFormatterParams{
				LogFormatterParams: gin.LogFormatterParams{
					Request: c.Request,
					Keys:    c.Keys,
				},
				Debug: debug,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()
			param.RequestId = requestid.Get(c)

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			slog.Info(formatter(param))
		}
	}
}
