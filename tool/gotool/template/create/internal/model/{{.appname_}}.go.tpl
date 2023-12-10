package model

import (
	"strconv"

	"gorm.io/gorm"
)

type {{.Appname}} struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (u {{.Appname}}) TableName() string {
	return "{{.appname_}}"
}
func (u {{.Appname}}) String() string {
	return strconv.Itoa(u.Id) + " " + u.Name
}

type {{.Appname}}Model struct {
	db *gorm.DB
}

func New{{.Appname}}Model(db *gorm.DB) *{{.Appname}}Model {
	return &{{.Appname}}Model{
		db: db,
	}
}

func (m *{{.Appname}}Model) Get{{.Appname}}ById(id int64) {{.Appname}} {
	var {{.appname_}} {{.Appname}}
	m.db.Model({{.Appname}}{}).Where("id=?", id).Take(&{{.appname_}})
	return {{.appname_}}
}
