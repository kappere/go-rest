package rest

import (
	"embed"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config/conf"
)

func staticResourceHandler(engine *gin.Engine, staticResourceConfig conf.StaticResourceConfig) {
	staticFs := staticResourceConfig.Fs
	if embedFs, ok := staticFs.(embed.FS); ok {
		fsPrefix := staticResourceConfig.Location
		fileServer := http.FileServer(http.FS(embedFs))
		engine.GET("/", func(c *gin.Context) {
			data, err := embedFs.ReadFile(fsPrefix + "/index.html")
			if err != nil {
				http.NotFound(c.Writer, c.Request)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, string(data))
		})
		engine.NoRoute(func(c *gin.Context) {
			r2 := new(http.Request)
			*r2 = *c.Request
			r2.URL = new(url.URL)
			*r2.URL = *c.Request.URL
			r2.URL.Path = fsPrefix + c.Request.URL.Path
			r2.URL.RawPath = c.Request.URL.RawPath
			if c.Request.URL.RawPath != "" {
				r2.URL.RawPath = fsPrefix + c.Request.URL.RawPath
			}
			fileServer.ServeHTTP(c.Writer, r2)
		})
		slog.Info("Static resource mapping: [NoRoute] => embed" + fsPrefix)
	} else if httpDirFs, ok := staticFs.(http.Dir); ok {
		fileServer := http.FileServer(httpDirFs)
		engine.GET("/", func(c *gin.Context) {
			data, err := os.ReadFile(string(httpDirFs) + "/index.html")
			if err != nil {
				http.NotFound(c.Writer, c.Request)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, string(data))
		})
		engine.NoRoute(func(c *gin.Context) {
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
		slog.Info("static resource mapping: [NoRoute] => " + string(httpDirFs))
	}
}
