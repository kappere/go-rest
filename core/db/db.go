package db

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/kappere/go-rest/core/config/conf"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func NewDatabase(dbConf conf.DatabaseConfig, debug bool) *gorm.DB {
	if dbConf.Dsn == "" {
		return nil
	}
	sqlLogger := logger.New(
		log.Default(),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,
			Colorful:                  debug,
		},
	)

	db, err := gorm.Open(dbConf.Dialector.(gorm.Dialector), &gorm.Config{
		Logger: sqlLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",    // table name prefix, table for `User` would be `t_users`
			SingularTable: true,  // use singular table name, table for `User` would be `user` with this option enabled
			NoLowerCase:   false, // skip the snake_casing of names
			NameReplacer:  nil,   // use name replacer to change struct/field name before convert it to db name
		},
	})
	if err != nil {
		panic(err)
	}
	slog.Info("Init datasource")
	return db
}

func Transaction(db *gorm.DB, fn func(*gorm.DB) interface{}) interface{} {
	var result interface{}
	var pnc interface{}
	f := func(tx0 *gorm.DB) (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("%s", r)
				slog.Error("Transaction rollback for error.", "error", msg)
				err = errors.New(msg)
				pnc = r
			}
		}()
		result = fn(tx0)
		return err
	}
	err := db.Transaction(f)
	if err != nil {
		panic(pnc)
	}
	return result
}
