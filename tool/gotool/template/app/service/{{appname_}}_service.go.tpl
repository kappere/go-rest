package service

import (
	"{{.fullprojectname}}/model"
	"github.com/kappere/go-rest/core/db"
	"github.com/kappere/go-rest/core/rest"
)

func Get{{.Appname}}(c *rest.Context) *rest.Resp {
	{{.appname_}} := model.{{.Appname}}{}
	db.Db.Take(&{{.appname_}}, 1)
	return rest.Success({{.appname_}})
}
