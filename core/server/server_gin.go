package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/db"
	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/middleware"
	"github.com/kappere/go-rest/core/redis"
	"github.com/kappere/go-rest/core/rest"
	"github.com/kappere/go-rest/core/rpc"
	"github.com/kappere/go-rest/core/task"
)

var GinEngine *rest.Engine

func Run(args []string, conf *rest.Config, routeFunc func(*rest.Engine)) {
	startTime := time.Now()
	// 启动服务组件
	cancel := setupComponent(conf)
	defer cancel()
	// 监控服务信息
	setupMonitor()
	// 创建engine
	engine := createEngine(conf)
	GinEngine = engine
	// 初始化中间件
	setupMiddleware(engine, conf)
	// 静态资源路由
	staticResourceRouter(engine, conf)
	// 初始化路由
	routeFunc(engine)
	// 启动定时任务
	task.Start()
	// 启动服务
	startServer(engine, conf, startTime)
}

func createEngine(conf *rest.Config) *rest.Engine {
	if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = *logger.GetWritter()
	gin.DefaultErrorWriter = *logger.GetWritter()
	return gin.New()
}

func startServer(engine *rest.Engine, conf *rest.Config, startTime time.Time) {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(conf.Port),
		Handler: engine,
	}
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("listen: %s\n", err)
			os.Exit(1)
		}
	}()
	logger.Info("started server [:%d] in %.3f seconds", conf.Port, float32(time.Now().UnixNano()-startTime.UnixNano())/1e9)
	if conf.Debug {
		logger.Info("visit: http://localhost:%d", conf.Port)
	}

	// 等待中断信号以优雅地关闭服务器（设置 8 秒的超时时间）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server force stop: %v", err)
		os.Exit(1)
	}
	logger.Info("server exit")
}

// 初始化服务组件
func setupComponent(conf *rest.Config) func() {
	// 初始化日志配置
	logger.Setup(&conf.Log, logger.InfoLevel, logger.ByDay, 60)

	logger.Info("===========================")
	logger.Info("app     : %s", conf.AppName)
	logger.Info("mode    : %s", conf.Profile)
	logger.Info("runtime : %s", runtime.GOOS+"-"+runtime.GOARCH)
	logger.Info("exec    : %s", os.Args[0])
	logger.Info("logdir  : %s", conf.Log.Path)
	logger.Info("===========================")

	// 初始化Redis
	redis.Setup(&conf.Redis)

	// 初始化数据库
	dbCancel := db.Setup(&conf.Database)
	return func() {
		dbCancel()
	}
}

// 初始化中间件
func setupMiddleware(engine *rest.Engine, conf *rest.Config) {
	// 错误恢复中间件
	engine.Use(middleware.NiceRecovery())
	// 限流中间件
	if conf.PeriodLimit.Enable {
		if redis.Rdb == nil && redis.ClusterRdb == nil {
			panic("please config redis first")
		}
		engine.Use(middleware.PeriodLimitMiddleware(conf.PeriodLimit.Period, conf.PeriodLimit.Quota, redis.Rdb, redis.ClusterRdb, "PERIODLIMIT"))
	}
	// 自定义日志格式
	engine.Use(middleware.NiceLoggerFormatter(func(param middleware.LogFormatterParams) string {
		return fmt.Sprintf("%s%s [%s][%s] %s %s %s %d %s %s\n",
			logger.INFO_PREFIX,
			logger.FormatDateTime(param.TimeStamp),
			param.RequestId,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))
	// 请求ID
	engine.Use(requestid.New())

	// 初始化session中间件
	if sessionMiddleware := middleware.SessionMiddleware(conf); sessionMiddleware != nil {
		engine.Use(sessionMiddleware)
	}

	// 初始化RPC客户端
	rpc.InitClient(&conf.Rpc)
}

// 初始化静态资源路由
func staticResourceRouter(engine *rest.Engine, conf *rest.Config) {
	if conf.StaticResource.Fs == nil {
		return
	}
	staticResourceHandler(engine, &conf.StaticResource)
}
