package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kappere/go-rest/core/rest"
)

func abs(v int64) int64 {
	if v >= 0 {
		return v
	}
	return -v
}

func RpcMiddleware(rpcConf *rest.RpcConfig) rest.HandlerFunc {
	return func(c *rest.Context) {
		rpc_token := c.GetHeader("inner_token_enc")
		if rpc_token == "" {
			c.JSON(http.StatusForbidden, rest.ErrorWithCode("missing rpc token", -552))
			c.Abort()
			return
		}
		tks := strings.Split(rpc_token, "#")
		if len(tks) != 3 {
			c.JSON(http.StatusForbidden, rest.ErrorWithCode("invalid rpc format", -552))
			c.Abort()
			return
		}
		timestamp, _ := strconv.ParseInt(tks[2], 10, 64)
		if abs(timestamp-time.Now().UnixMilli()) > 1000*180 {
			c.JSON(http.StatusForbidden, rest.ErrorWithCode("sync time please!", -550))
			c.Abort()
			return
		}
		enc := tks[0]
		hash := sha256.New()
		hash.Write([]byte(rpcConf.Token + "#" + tks[1] + "#" + tks[2]))
		enc2 := hex.EncodeToString(hash.Sum(nil))
		if enc == "" || enc != enc2 {
			c.JSON(http.StatusForbidden, rest.ErrorWithCode("invalid rpc token!", -552))
			c.Abort()
			return
		}
		c.Next()
	}
}
