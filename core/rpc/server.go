package rpc

import (
	"github.com/gin-gonic/gin"
	"wataru.com/go-rest/core/middleware"
	"wataru.com/go-rest/core/rest"
)

const (
	RPC_PREFIX = "/_remote_procedure_call"
)

func Server(engine *rest.Engine, rpcConf *rest.RpcConfig) *gin.RouterGroup {
	return engine.Group(RPC_PREFIX, middleware.RpcMiddleware(rpcConf))
}
