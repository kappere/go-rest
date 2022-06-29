package api

import (
	"net/http"

	"{{.fullprojectname}}/service"
	"github.com/kappere/go-rest/core/rest"
)

func Get{{.Appname}}Handler() rest.HandlerFunc {
	return func(c *rest.Context) {
		c.JSON(http.StatusOK, service.{{.Appname}}.Get{{.Appname}}(c))
	}
}
