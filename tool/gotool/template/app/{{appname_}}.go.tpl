package main

import (
	"embed"
	"os"

	"{{.fullprojectname}}/api"
	_ "{{.fullprojectname}}/task"
)

/**
外部静态资源初始化：
	api.StaticFs(http.Dir("public"), "")
内部embed静态资源初始化：
	//go:embed public/*
	var fs embed.FS
	api.StaticFs(fs, "public")
*/

//go:embed public/*
var fs embed.FS

func main() {
	api.StaticFs(fs, "public")
	api.Run(os.Args)
}
