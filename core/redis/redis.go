package redis

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/config/conf"
)

func NewRedisClient(redisConfig conf.RedisConfig) (*redis.Client, error) {
	if redisConfig.Addr == "" {
		return nil, errors.New("empty redis address")
	}
	var rdb *redis.Client
	addrs := strings.Split(redisConfig.Addr, ",")
	rdb = redis.NewClient(&redis.Options{
		Addr:     addrs[0],
		Password: redisConfig.Password, // no password set
		DB:       0,                    // use default DB
	})
	slog.Info("Init redis,", "addr", redisConfig.Addr)

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func NewRedisClusterClient(redisConfig conf.RedisConfig) (*redis.ClusterClient, error) {
	if redisConfig.Addr == "" {
		return nil, errors.New("empty redis address")
	}
	var clusterRdb *redis.ClusterClient
	addrs := strings.Split(redisConfig.Addr, ",")
	clusterRdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,
		// To route commands by latency or randomly, enable one of the following.
		//RouteByLatency: true,
		//RouteRandomly: true,
	})
	slog.Info("Init redis cluster,", "addr", redisConfig.Addr)

	err := clusterRdb.ForEachShard(context.Background(), func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
	if err != nil {
		return nil, err
	}
	return clusterRdb, nil
}
