package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/rest"
)

var ctx = context.Background()
var Rdb *redis.Client

func Setup(redisConf *rest.RedisConfig) {
	if redisConf.Host == "" {
		return
	}
	addr := redisConf.Host + ":" + strconv.Itoa(redisConf.Port)
	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: redisConf.Password, // no password set
		DB:       0,                  // use default DB
	})
	logger.Info("init redis")

	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

func Test() {
	err := Rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := Rdb.Get(ctx, "key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := Rdb.Get(ctx, "key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist
}
