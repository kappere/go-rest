package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/httpx"
)

func abs(v int64) int64 {
	if v >= 0 {
		return v
	}
	return -v
}

func Rpc(rpcConf conf.RpcConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rpcConf.Token != "" {
			rpc_token := c.GetHeader("inner_token_enc")
			if rpc_token == "" {
				c.JSON(http.StatusForbidden, httpx.ErrorWithCode("missing rpc token", -552))
				c.Abort()
				return
			}
			tks := strings.Split(rpc_token, "#")
			if len(tks) != 3 {
				c.JSON(http.StatusForbidden, httpx.ErrorWithCode("invalid rpc format", -552))
				c.Abort()
				return
			}
			timestamp, err := strconv.ParseInt(tks[2], 10, 64)
			if err != nil {
				c.JSON(http.StatusForbidden, httpx.ErrorWithCode("invalid timestamp!", -550))
				c.Abort()
				return
			}
			if abs(timestamp-time.Now().UnixMilli()) > 1000*180 {
				c.JSON(http.StatusForbidden, httpx.ErrorWithCode("sync time please!", -550))
				c.Abort()
				return
			}
			enc := tks[0]
			hash := sha256.New()
			hash.Write([]byte(rpcConf.Token + "#" + tks[1] + "#" + tks[2]))
			enc2 := hex.EncodeToString(hash.Sum(nil))
			if enc == "" || enc != enc2 {
				c.JSON(http.StatusForbidden, httpx.ErrorWithCode("invalid rpc token!", -552))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
