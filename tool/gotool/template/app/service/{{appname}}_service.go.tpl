package service

import (
	"{{.fullprojectname}}/model"
	"wataru.com/go-rest/core/db"
	"wataru.com/go-rest/core/rest"
)

func Get{{.Appname}}(c *rest.Context) *rest.Resp {
	{{.appname}} := model.{{.Appname}}{}
	db.Db.Take(&{{.appname}}, 1)
	return rest.Success({{.appname}})
}
