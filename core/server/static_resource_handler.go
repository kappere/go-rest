package server

import (
	"embed"
	"io/ioutil"
	"net/http"
	"net/url"

	"wataru.com/go-rest/core/logger"
	"wataru.com/go-rest/core/rest"
)

func staticResourceHandler(engine *rest.Engine, staticResourceConf *rest.StaticResourceConfig) {
	staticFs := staticResourceConf.Fs
	if embedFs, ok := staticFs.(embed.FS); ok {
		fsPrefix := staticResourceConf.Location
		fileServer := http.FileServer(http.FS(embedFs))
		engine.GET("/", func(c *rest.Context) {
			data, err := embedFs.ReadFile(fsPrefix + "/index.html")
			if err != nil {
				http.NotFound(c.Writer, c.Request)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, string(data))
		})
		engine.NoRoute(func(c *rest.Context) {
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
		logger.Info("static resource mapping: [NoRoute] => embed:%s", fsPrefix)
	} else if httpDirFs, ok := staticFs.(http.Dir); ok {
		fileServer := http.FileServer(httpDirFs)
		engine.GET("/", func(c *rest.Context) {
			data, err := ioutil.ReadFile(string(httpDirFs) + "/index.html")
			if err != nil {
				http.NotFound(c.Writer, c.Request)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, string(data))
		})
		engine.NoRoute(func(c *rest.Context) {
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
		logger.Info("static resource mapping: [NoRoute] => %s", string(httpDirFs))
	}
}
