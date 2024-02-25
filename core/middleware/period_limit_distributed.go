// 参照https://github.com/zeromicro/go-zero
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/httpx"
	rest_redis "github.com/kappere/go-rest/core/redis"
)

// PeriodLimit 分布式限流中间件
func PeriodLimitDistributedMiddleware(periodLimitConfig conf.PeriodLimitConfig, redisConfig conf.RedisConfig, opts ...PeriodOption) (gin.HandlerFunc, func()) {
	var limitStore *redis.Client
	var clusterLimitStore *redis.ClusterClient
	addrLen := len(strings.Split(redisConfig.Addr, ","))
	var err error
	if addrLen > 1 {
		limitStore, err = rest_redis.NewRedisClient(redisConfig)
	} else if addrLen == 1 {
		clusterLimitStore, err = rest_redis.NewRedisClusterClient(redisConfig)
	}
	if err != nil {
		panic(err)
	}
	limit := newPeriodLimit(periodLimitConfig.Period, periodLimitConfig.Quota, limitStore, clusterLimitStore, "REST_PERIOD_LIMIT", opts...)
	return func(c *gin.Context) {
			r, err := limit.take(c.Request.RequestURI)
			if err != nil {
				slog.Error("Peroid limit middleware has error,", "URI", c.Request.RequestURI)
				c.JSON(http.StatusOK, httpx.ErrorWithCode("Peroid limit middleware has error", httpx.STATUS_ERROR_LIMIT))
				c.Abort()
				return
			}
			if r == OverQuota {
				c.JSON(http.StatusOK, httpx.ErrorWithCode("Resource limited", httpx.STATUS_ERROR_LIMIT))
				c.Abort()
				return
			}
			c.Next()
		}, func() {
			limit.close()
		}
}

// to be compatible with aliyun redis, we cannot use `local key = KEYS[1]` to reuse the key
const PERIOD_SCRIPT = `local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCRBY", KEYS[1], 1)
if current == 1 then
    redis.call("expire", KEYS[1], window)
    return 1
elseif current < limit then
    return 1
elseif current == limit then
    return 2
else
    return 3
end`

const (
	// Unknown means not initialized state.
	Unknown = iota
	// Allowed means allowed state.
	Allowed
	// HitQuota means this request exactly hit the quota.
	HitQuota
	// OverQuota means passed the quota.
	OverQuota
)

// ErrUnknownCode is an error that represents unknown status code.
var ErrUnknownCode = errors.New("unknown status code")

type (
	// PeriodOption defines the method to customize a PeriodLimit.
	PeriodOption func(l *PeriodLimit)

	// A PeriodLimit is used to limit requests during a period of time.
	PeriodLimit struct {
		period            int
		quota             int
		limitStore        *redis.Client
		clusterLimitStore *redis.ClusterClient
		keyPrefix         string
		align             bool
	}
)

// newPeriodLimit returns a PeriodLimit with given parameters.
func newPeriodLimit(period, quota int, limitStore *redis.Client, clusterLimitStore *redis.ClusterClient, keyPrefix string,
	opts ...PeriodOption) *PeriodLimit {
	limiter := &PeriodLimit{
		period:            period,
		quota:             quota,
		limitStore:        limitStore,
		clusterLimitStore: clusterLimitStore,
		keyPrefix:         keyPrefix,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

func (h *PeriodLimit) close() {
	if h.limitStore != nil {
		h.limitStore.Close()
	}
	if h.clusterLimitStore != nil {
		h.clusterLimitStore.Close()
	}
}

// Take requests a permit, it returns the permit state.
func (h *PeriodLimit) take(key string) (int, error) {
	var resp *redis.Cmd
	if h.clusterLimitStore != nil {
		resp = h.clusterLimitStore.Eval(context.Background(), PERIOD_SCRIPT, []string{h.keyPrefix + key},
			strconv.Itoa(h.quota),
			strconv.Itoa(h.calcExpireSeconds()),
		)
	} else {
		resp = h.limitStore.Eval(context.Background(), PERIOD_SCRIPT, []string{h.keyPrefix + key},
			strconv.Itoa(h.quota),
			strconv.Itoa(h.calcExpireSeconds()),
		)
	}

	code, err := resp.Int()
	if err != nil {
		return Unknown, ErrUnknownCode
	}

	return code, nil
}

func (h *PeriodLimit) calcExpireSeconds() int {
	if h.align {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return h.period - int(unix%int64(h.period))
	}

	return h.period
}

// Align returns a func to customize a PeriodLimit with alignment.
// For example, if we want to limit end users with 5 sms verification messages every day,
// we need to align with the local timezone and the start of the day.
func Align() PeriodOption {
	return func(l *PeriodLimit) {
		l.align = true
	}
}
