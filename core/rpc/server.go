package rpc

import (
	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/middleware"
)

const (
	RPC_PREFIX = "/_rpc_"
)

func Server(engine *gin.Engine, rpcConf conf.RpcConfig) *gin.RouterGroup {
	return engine.Group(RPC_PREFIX, middleware.Rpc(rpcConf))
}
