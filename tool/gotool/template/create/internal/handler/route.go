package handler

import (
	"github.com/gin-gonic/gin"
	"{{.fullprojectname}}/internal/context/svc"
	"{{.fullprojectname}}/internal/handler/{{.appname}}"
	"github.com/kappere/go-rest/core/rpc"
)

func RegisterHandlers(ctx *svc.ServiceContext) {
	engine := ctx.Server.Engine

	{{.appname_}}Group := engine.Group("/{{.appname_}}")
	{{.appname_}}Group.GET("/get", func(c *gin.Context) { {{.appname_}}.Find{{.Appname}}ById(c, ctx) })
	{{.appname_}}Group.GET("/rget", func(c *gin.Context) { {{.appname_}}.RpcFind{{.Appname}}ById(c, ctx) })

	rpcServer := rpc.Server(engine, ctx.Config.Http.Rpc)
	rpcServer.POST("/{{.appname_}}/get", func(c *gin.Context) { {{.appname_}}.Find{{.Appname}}ById(c, ctx) })
}
