// 参照https://github.com/zeromicro/go-zero
package middleware

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/httpx"
)

// PeriodLimit 本地限流中间件
func PeriodLimitLocalMiddleware(periodLimitConfig conf.PeriodLimitConfig) gin.HandlerFunc {
	limit := newPeriodLimitLocal(periodLimitConfig.Period, periodLimitConfig.Quota)
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
	}
}

type (
	PeriodLimitLocal struct {
		period int
		quota  int
		lock   sync.Mutex
	}
)

// newPeriodLimit returns a PeriodLimit with given parameters.
func newPeriodLimitLocal(period, quota int) *PeriodLimitLocal {
	limiter := &PeriodLimitLocal{
		period: period,
		quota:  quota,
		lock:   sync.Mutex{},
	}

	return limiter
}

var reqTimesMap map[string]ReqTimes

type ReqTimes struct {
	times      int
	expireTime time.Time
}

func (h *PeriodLimitLocal) take(key string) (int, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	reqTimes, exists := reqTimesMap[key]
	if time.Now().After(reqTimes.expireTime) {
		exists = false
		delete(reqTimesMap, key)
	}
	reqTimes.times++
	code := Unknown
	if !exists {
		code = Allowed
		reqTimesMap[key] = ReqTimes{
			times:      1,
			expireTime: time.Now().Add(time.Duration(h.period) * time.Second),
		}
	} else if reqTimes.times < h.quota {
		code = Allowed
	} else if reqTimes.times == h.quota {
		code = HitQuota
	} else {
		code = OverQuota
	}
	return code, nil
}
