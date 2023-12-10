package service

import (
	"{{.fullprojectname}}/internal/context/db"
	"{{.fullprojectname}}/internal/model"
)

type {{.Appname}}Service struct {
	dbCtx *db.DbContext
}

func New{{.Appname}}Service(dbCtx *db.DbContext) *{{.Appname}}Service {
	return &{{.Appname}}Service{
		dbCtx: dbCtx,
	}
}

func (s *{{.Appname}}Service) Find{{.Appname}}ById(id int64) model.{{.Appname}} {
	return s.dbCtx.{{.Appname}}Model.Get{{.Appname}}ById(id)
}
