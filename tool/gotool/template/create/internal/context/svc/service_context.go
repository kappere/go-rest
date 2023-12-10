package svc

import (
	"log/slog"

	"github.com/kappere/go-rest/core/rest"
	"{{.fullprojectname}}/internal/config"
	"{{.fullprojectname}}/internal/context/db"
	"{{.fullprojectname}}/internal/rpc"
	"{{.fullprojectname}}/internal/service"
)

type ServiceContext struct {
	Server    *rest.Server
	Config    *config.Config
	DbContext *db.DbContext

	Srv Srv
	Rpc Rpc
}

type Srv struct {
	{{.Appname}}Service *service.{{.Appname}}Service
}

type Rpc struct {
	{{.Appname}}Rpc *rpc.{{.Appname}}Rpc
}

func NewServiceContext(server *rest.Server, c *config.Config) *ServiceContext {
	dbContext := db.NewDbContext(c)
	ctx := ServiceContext{
		Server:    server,
		Config:    c,
		DbContext: dbContext,

		Srv: Srv{
			{{.Appname}}Service: service.New{{.Appname}}Service(dbContext),
		},
		Rpc: Rpc{
			{{.Appname}}Rpc: rpc.New{{.Appname}}Rpc(),
		},
	}
	ctx.AddClose(ctx.close)
	return &ctx
}

func (s *ServiceContext) AddClose(f func()) {
	s.Server.AddClose(f)
}

func (c *ServiceContext) close() {
	c.DbContext.Close()
	slog.Info("ServiceContext closed.")
}
