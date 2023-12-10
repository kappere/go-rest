package config

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/kappere/go-rest/core/config/conf"
)

type BaseConfig struct {
	App      conf.AppConfig
	Http     conf.HttpConfig
	Log      conf.LogConfig
	Database conf.DatabaseConfig
	Redis    conf.RedisConfig
}

var DefaultBaseConfig = BaseConfig{
	App: conf.AppConfig{
		Name:    "app",
		Profile: "prod",
		Debug:   false,
	},
	Http: conf.HttpConfig{
		Port: 80,
		Session: conf.SessionConfig{
			Name:      "sessionid",
			Domain:    "",
			Path:      "/",
			Secure:    false,
			HttpOnly:  true,
			MaxAge:    86400 * 30,
			SameSite:  http.SameSiteDefaultMode,
			StoreType: "memory",
		},
		PeriodLimit: conf.PeriodLimitConfig{
			Enable: false,
			Period: 5,
			Quota:  100,
		},
		OAuth2: conf.OAuth2Config{
			Enable:   false,
			Expire:   7200,
			TokenUri: "/token",
		},
		Rpc: conf.RpcConfig{
			Token: strconv.Itoa(rand.Int()),
			// Kubernetes IpProxy
			Type: "IpProxy",
			IpProxy: conf.IpProxyConfig{
				Proxy: map[string]string{
					"*": "http://127.0.0.1:8080",
				},
			},
			Kubernetes: conf.KubernetesConfig{
				Namespace: "default",
				Proxy: map[string]string{
					"*": "http://127.0.0.1:8080/api/v1/namespaces/{namespace}/services/http:{app}:/proxy",
				},
				PortName: "http",
			},
		},
	},
	Log: conf.LogConfig{
		Path: "log",
	},
	Database: conf.DatabaseConfig{},
	Redis:    conf.RedisConfig{},
}
