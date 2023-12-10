package main

import (
	"flag"

	"github.com/kappere/go-rest/core/rest"
	"{{.fullprojectname}}/internal/config"
	"{{.fullprojectname}}/internal/context/svc"
	"{{.fullprojectname}}/internal/handler"
)

func main() {
	configFile := flag.String("config", "etc/{{.appname}}.yaml", "the config file")
	flag.Parse()

	c := config.Load(*configFile)

	server := rest.NewServer(c.BaseConfig)
	defer server.Close()

	ctx := svc.NewServiceContext(server, c)
	handler.RegisterHandlers(ctx)

	server.Run()
}
