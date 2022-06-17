package api

import (
	"github.com/kappere/go-rest/core/rest"
	"github.com/kappere/go-rest/core/rpc"
	"github.com/kappere/go-rest/core/server"
	"gorm.io/driver/mysql"
	"{{.fullprojectname}}/config"
)

// 静态资源文件
var staticFs interface{}
var fsLocation string

func StaticFs(fs interface{}, location string) {
	staticFs = fs
	fsLocation = location
}

func Run(args []string) {
	rest.LoadConfig(args, &config.Conf, config.ConfFs)
	config.Conf.Database.Dialector = mysql.Open(config.Conf.Database.Dsn)
	config.Conf.StaticResource.Fs = staticFs
	config.Conf.StaticResource.Location = fsLocation
	server.Run(args, &config.Conf.Config, initRoute)
}

func initRoute(engine *rest.Engine) {
	// HTTP路由
	engine.GET("/get{{.Appname}}", Get{{.Appname}}Handler())
    
	// RPC服务间调用路由(POST)
	rpc.Server(engine, &config.Conf.Rpc).
		POST("/get{{.Appname}}", Get{{.Appname}}Handler())
	// RPC调用：rpc.Service("{{.appname}}").Call("/get{{.Appname}}", nil)
}
