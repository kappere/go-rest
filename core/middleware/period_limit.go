// 参照https://github.com/zeromicro/go-zero
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kappere/go-rest/core/logger"
	"github.com/kappere/go-rest/core/rest"
)

// PeriodLimit 分布式限流中间件
func PeriodLimitMiddleware(period int, quota int, limitStore *redis.Client, keyPrefix string, opts ...PeriodOption) rest.HandlerFunc {
	limit := newPeriodLimit(period, quota, limitStore, keyPrefix)
	return func(c *rest.Context) {
		r, err := limit.take(c.Request.RequestURI)
		if err != nil {
			logger.Error("peroid limit middleware has error, URI: %s", c.Request.RequestURI)
			c.JSON(http.StatusOK, rest.ErrorWithCode("peroid limit middleware has error", rest.STATUS_ERROR_LIMIT))
			c.Abort()
			return
		}
		if r == OverQuota {
			c.JSON(http.StatusOK, rest.ErrorWithCode("resource limited", rest.STATUS_ERROR_LIMIT))
			c.Abort()
			return
		}
		c.Next()
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
    return 0
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

	internalOverQuota = 0
	internalAllowed   = 1
	internalHitQuota  = 2
)

// ErrUnknownCode is an error that represents unknown status code.
var ErrUnknownCode = errors.New("unknown status code")

type (
	// PeriodOption defines the method to customize a PeriodLimit.
	PeriodOption func(l *PeriodLimit)

	// A PeriodLimit is used to limit requests during a period of time.
	PeriodLimit struct {
		period     int
		quota      int
		limitStore *redis.Client
		keyPrefix  string
		align      bool
	}
)

// newPeriodLimit returns a PeriodLimit with given parameters.
func newPeriodLimit(period, quota int, limitStore *redis.Client, keyPrefix string,
	opts ...PeriodOption) *PeriodLimit {
	limiter := &PeriodLimit{
		period:     period,
		quota:      quota,
		limitStore: limitStore,
		keyPrefix:  keyPrefix,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// Take requests a permit, it returns the permit state.
func (h *PeriodLimit) take(key string) (int, error) {
	resp := h.limitStore.Eval(context.Background(), PERIOD_SCRIPT, []string{h.keyPrefix + key},
		strconv.Itoa(h.quota),
		strconv.Itoa(h.calcExpireSeconds()),
	)

	code, err := resp.Int64()
	if err != nil {
		return Unknown, ErrUnknownCode
	}

	switch code {
	case internalOverQuota:
		return OverQuota, nil
	case internalAllowed:
		return Allowed, nil
	case internalHitQuota:
		return HitQuota, nil
	default:
		return Unknown, ErrUnknownCode
	}
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
