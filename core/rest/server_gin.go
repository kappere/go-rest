package rest

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config"
	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/middleware"
	"github.com/kappere/go-rest/core/rpc"
)

var GinEngine *gin.Engine
var startTime time.Time

type Server struct {
	Engine         *gin.Engine
	Config         config.BaseConfig
	closeFunctions []func()
}

func NewServer(baseConfig config.BaseConfig) *Server {
	startTime = time.Now()
	logger.InitLogger(baseConfig.Log, baseConfig.App.Name)
	// 启动服务组件
	setupComponent(baseConfig)
	// 创建engine
	engine := createEngine(baseConfig)
	GinEngine = engine
	server := &Server{
		Engine: engine,
		Config: baseConfig,
	}
	// 初始化中间件
	setupMiddleware(server, baseConfig)
	// 初始化RPC客户端
	rpc.InitClient(baseConfig.Http.Rpc)
	// 静态资源路由
	staticResourceRouter(engine, baseConfig.Http)
	// 初始化路由
	// routeFunc(engine)
	return server
}

func (s *Server) Run() {
	// 启动服务
	startHttpServer(s.Engine, s.Config.Http, startTime)
}

func (s *Server) Close() {
	for i := range s.closeFunctions {
		s.closeFunctions[len(s.closeFunctions)-1-i]()
	}
}

func (s *Server) AddClose(f func()) {
	s.closeFunctions = append(s.closeFunctions, f)
}

func createEngine(conf config.BaseConfig) *gin.Engine {
	if !conf.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func startHttpServer(engine *gin.Engine, httpConfig conf.HttpConfig, startTime time.Time) {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(httpConfig.Port),
		Handler: engine,
	}
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Listen and serve failed:", "error", err)
			os.Exit(1)
		}
	}()
	// 监控服务信息
	setupMonitor()
	slog.Info(fmt.Sprintf("Started server [:%d] in %.3f seconds", httpConfig.Port, float32(time.Now().UnixNano()-startTime.UnixNano())/1e9))

	// 等待中断信号以优雅地关闭服务器（设置 8 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server force stop:", "error", err)
		os.Exit(1)
	}
	slog.Info("Server closed.")
}

// 初始化服务组件
func setupComponent(baseConfig config.BaseConfig) {
	slog.Info("================================")
	slog.Info("app     : " + baseConfig.App.Name)
	slog.Info("profile : " + baseConfig.App.Profile)
	slog.Info("debug   : " + strconv.FormatBool(baseConfig.App.Debug))
	slog.Info("runtime : " + runtime.GOOS + "-" + runtime.GOARCH)
	slog.Info("exec    : " + os.Args[0])
	slog.Info("logdir  : " + baseConfig.Log.Path)
	slog.Info("port    : " + strconv.Itoa(baseConfig.Http.Port))
	slog.Info("================================")
}

// 初始化中间件
func setupMiddleware(server *Server, baseConfig config.BaseConfig) {
	// 错误恢复中间件
	server.Engine.Use(middleware.NiceRecovery())
	slog.Info("[middleware] NiceRecovery")

	// 自定义日志格式
	server.Engine.Use(middleware.NiceLoggerFormatter(nil, baseConfig.App.Debug))
	slog.Info("[middleware] NiceLoggerFormatter")

	// 请求ID
	server.Engine.Use(requestid.New())
	slog.Info("[middleware] requestid")

	// Session
	if baseConfig.Http.Session.StoreType != "" {
		server.Engine.Use(middleware.Session(baseConfig.Http.Session, baseConfig.Redis))
		slog.Info("[middleware] Session (" + baseConfig.Http.Session.StoreType + ")")
	}

	// 限流
	if baseConfig.Http.PeriodLimit.Enable {
		if baseConfig.Http.PeriodLimit.Distributed {
			limit, close := middleware.PeriodLimitDistributedMiddleware(baseConfig.Http.PeriodLimit, baseConfig.Redis)
			server.Engine.Use(limit)
			server.AddClose(close)
		} else {
			limit := middleware.PeriodLimitLocalMiddleware(baseConfig.Http.PeriodLimit)
			server.Engine.Use(limit)
		}
		slog.Info("[middleware] PeriodLimitMiddleware")
	}
}

// 初始化静态资源路由
func staticResourceRouter(engine *gin.Engine, httpConfig conf.HttpConfig) {
	if httpConfig.StaticResource.Fs == nil {
		return
	}
	staticResourceHandler(engine, httpConfig.StaticResource)
}
