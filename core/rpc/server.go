package rpc

import (
	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/middleware"
	"github.com/kappere/go-rest/core/rest"
)

const (
	RPC_PREFIX = "/_remote_procedure_call"
)

func Server(engine *rest.Engine, rpcConf *rest.RpcConfig) *gin.RouterGroup {
	return engine.Group(RPC_PREFIX, middleware.Rpc(rpcConf))
}
