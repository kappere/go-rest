package redis

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/rest"
)

var ctx = context.Background()
var Rdb *redis.Client
var ClusterRdb *redis.ClusterClient

func Setup(redisConf *rest.RedisConfig) {
	if redisConf.Addr == "" {
		return
	}
	addrs := strings.Split(redisConf.Addr, ",")
	if len(addrs) > 1 {
		ClusterRdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
			// To route commands by latency or randomly, enable one of the following.
			//RouteByLatency: true,
			//RouteRandomly: true,
		})
		logger.Info("init redis cluster")

		err := ClusterRdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		})
		if err != nil {
			panic(err)
		}
	} else {
		Rdb = redis.NewClient(&redis.Options{
			Addr:     addrs[0],
			Password: redisConf.Password, // no password set
			DB:       0,                  // use default DB
		})
		logger.Info("init redis")

		_, err := Rdb.Ping(ctx).Result()
		if err != nil {
			panic(err)
		}
	}
}
