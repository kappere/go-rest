package middleware

import (
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memstore"
	ginredis "github.com/gin-contrib/sessions/redis"
	"wataru.com/go-rest/core/logger"
	"wataru.com/go-rest/core/rest"
)

// SessionMiddleware session中间件，详见https://github.com/gin-contrib/sessions
func SessionMiddleware(conf *rest.Config) rest.HandlerFunc {
	var store sessions.Store
	if conf.Session.StoreType == rest.CACHE_TYPE_MEMORY {
		store = memstore.NewStore([]byte("secret"))
	} else if conf.Session.StoreType == rest.CACHE_TYPE_COOKIE {
		store = cookie.NewStore([]byte("secret"))
	} else if conf.Session.StoreType == rest.CACHE_TYPE_REDIS {
		s, err := ginredis.NewStore(10, "tcp", conf.Redis.Host+":"+strconv.Itoa(conf.Redis.Port), conf.Redis.Password, []byte("secret"))
		if err != nil {
			panic(err)
		}
		store = s
	}
	if store != nil {
		store.Options(sessions.Options{
			Path:     conf.Session.Path,
			Domain:   conf.Session.Domain,
			MaxAge:   conf.Session.MaxAge,
			Secure:   conf.Session.Secure,
			HttpOnly: conf.Session.HttpOnly,
			SameSite: conf.Session.SameSite,
		})
		logger.Info("session type: %s", conf.Session.StoreType)
		return sessions.Sessions(conf.Session.Name, store)
	}
	return nil
}
