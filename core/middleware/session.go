package middleware

import (
	"log/slog"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config/conf"
)

const (
	STORAGE_TYPE_NONE   = "none"
	STORAGE_TYPE_MEMORY = "memory"
	STORAGE_TYPE_COOKIE = "cookie"
	STORAGE_TYPE_REDIS  = "redis"
)

// Session session中间件，详见https://github.com/gin-contrib/sessions
func Session(sessionConfig conf.SessionConfig, redisConfig conf.RedisConfig) gin.HandlerFunc {
	var store sessions.Store
	if sessionConfig.StoreType == STORAGE_TYPE_MEMORY {
		store = memstore.NewStore([]byte("secret"))
	} else if sessionConfig.StoreType == STORAGE_TYPE_COOKIE {
		store = cookie.NewStore([]byte("secret"))
	} else if sessionConfig.StoreType == STORAGE_TYPE_REDIS {
		s, err := redis.NewStore(10, "tcp", redisConfig.Addr, redisConfig.Password, []byte("secret"))
		if err != nil {
			panic(err)
		}
		store = s
	} else {
		slog.Error("Invalid session type", "type", sessionConfig.StoreType)
	}
	if store != nil {
		store.Options(sessions.Options{
			Path:     sessionConfig.Path,
			Domain:   sessionConfig.Domain,
			MaxAge:   sessionConfig.MaxAge,
			Secure:   sessionConfig.Secure,
			HttpOnly: sessionConfig.HttpOnly,
			SameSite: sessionConfig.SameSite,
		})
		return sessions.Sessions(sessionConfig.Name, store)
	}
	return nil
}
