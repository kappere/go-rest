package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"wataru.com/go-rest/core/logger"
	"wataru.com/go-rest/core/rest"
)

// Db 实例
var Db *gorm.DB

func Setup(dbConf *rest.DatabaseConfig) func() {
	if dbConf.Dsn == "" {
		return func() {}
	}
	_db, err := gorm.Open(dbConf.Dialector.(gorm.Dialector), &gorm.Config{
		Logger: logger.GetGormLogger(),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",    // table name prefix, table for `User` would be `t_users`
			SingularTable: true,  // use singular table name, table for `User` would be `user` with this option enabled
			NoLowerCase:   false, // skip the snake_casing of names
			NameReplacer:  nil,   // use name replacer to change struct/field name before convert it to db name
		},
	})
	Db = _db
	if err != nil {
		panic(err)
	}
	logger.Info("init datasource")
	return func() {
		sqlDB, err := Db.DB()
		if err != nil {
			logger.Error("datasource close failed: %v", err)
			return
		}
		err = sqlDB.Close()
		if err != nil {
			logger.Error("datasource close failed: %v", err)
		} else {
			logger.Info("datasource closed")
		}
	}
}

func Transaction(fn func(*gorm.DB) interface{}) interface{} {
	var result interface{}
	var pnc interface{}
	f := func(tx0 *gorm.DB) (err error) {
		defer func() {
			if r := recover(); r != nil {
				msg := fmt.Sprintf("%s", r)
				logger.Error("Transaction rollback for error: %s", msg)
				err = errors.New(msg)
				pnc = r
			}
		}()
		result = fn(tx0)
		return err
	}
	err := Db.Transaction(f)
	if err != nil {
		panic(pnc)
	}
	return result
}
