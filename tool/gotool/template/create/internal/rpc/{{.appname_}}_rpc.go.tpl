package rpc

import (
	"{{.fullprojectname}}/internal/model"
	"github.com/kappere/go-rest/core/rpc"
)

type {{.Appname}}Rpc struct {
	rpcService func() rpc.RpcService
}

func New{{.Appname}}Rpc() *{{.Appname}}Rpc {
	return &{{.Appname}}Rpc{
		func() rpc.RpcService {
			return rpc.Service("{{.appname}}")
		},
	}
}

func (r *{{.Appname}}Rpc) Find{{.Appname}}ById(id string) (model.{{.Appname}}, error) {
	{{.appname_}} := model.{{.Appname}}{}
	err := r.rpcService().Call("/{{.appname_}}/get?id="+id, nil).ToObj(&{{.appname_}})
	return {{.appname_}}, err
}
