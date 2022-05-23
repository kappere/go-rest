package api

import (
	"net/http"

	"{{.fullprojectname}}/service"
	"wataru.com/go-rest/core/rest"
)

func Get{{.Appname}}Handler() rest.HandlerFunc {
	return func(c *rest.Context) {
		c.JSON(http.StatusOK, service.Get{{.Appname}}(c))
	}
}
