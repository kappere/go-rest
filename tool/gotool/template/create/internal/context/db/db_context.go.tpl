package db

import (
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/db"
	gorest_redis "github.com/kappere/go-rest/core/redis"
	"github.com/kappere/go-rest/core/tool/redislock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"{{.fullprojectname}}/internal/config"
	"{{.fullprojectname}}/internal/model"
)

type DbContext struct {
	Config *config.Config
	Db     *gorm.DB
	Redis  *redis.Client

	{{.Appname}}Model *model.{{.Appname}}Model
}

func NewDbContext(c *config.Config) *DbContext {
	redisClient := gorest_redis.NewRedisClient(c.Redis)

	c.Database.Dialector = mysql.Open(c.Database.Dsn)
	database := db.NewDatabase(c.Database, c.App.Debug)
	redislock.SetStore(redisClient)

	ctx := DbContext{
		Config: c,
		Db:     database,
		Redis:  redisClient,

		{{.Appname}}Model: model.New{{.Appname}}Model(database),
	}
	return &ctx
}

func (c *DbContext) Close() {
	sqlDb, _ := c.Db.DB()
	sqlDb.Close()
	c.Redis.Close()
	slog.Info("DbContext closed.")
}
