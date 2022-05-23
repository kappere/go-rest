package model

import "strconv"

type {{.Appname}} struct {
	Id   *int    `json:"id"`
	Name *string `json:"name"`
}

func ({{.Appname}}) TableName() string {
	return "{{.appname}}"
}
func (u *{{.Appname}}) String() string {
	return strconv.Itoa(*u.Id) + " " + *u.Name
}
