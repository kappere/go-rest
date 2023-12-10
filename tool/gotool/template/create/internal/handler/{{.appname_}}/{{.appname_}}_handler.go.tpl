package {{.appname_}}

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"{{.fullprojectname}}/internal/context/svc"
	"github.com/kappere/go-rest/core/httpx"
)

func Find{{.Appname}}ById(c *gin.Context, ctx *svc.ServiceContext) {
	id, _ := strconv.ParseInt(c.Query("id"), 10, 64)
	{{.appname_}} := ctx.Srv.{{.Appname}}Service.Find{{.Appname}}ById(id)
	if {{.appname_}}.Id == 0 {
		c.PureJSON(http.StatusOK, httpx.Error("Not found {{.appname_}}!"))
		return
	}
	c.PureJSON(http.StatusOK, httpx.Ok({{.appname_}}))
}

func RpcFind{{.Appname}}ById(c *gin.Context, ctx *svc.ServiceContext) {
	{{.appname_}}, err := ctx.Rpc.{{.Appname}}Rpc.Find{{.Appname}}ById(c.Query("id"))
	if err != nil {
		c.PureJSON(http.StatusOK, httpx.ErrorWithCode(err.Error(), -1))
		return
	}
	c.PureJSON(http.StatusOK, httpx.Ok({{.appname_}}))
}
