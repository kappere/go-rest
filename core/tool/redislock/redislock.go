// 参照https://github.com/zeromicro/go-zero
package redislock

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"wataru.com/go-rest/core/logger"
	gorest_redis "wataru.com/go-rest/core/redis"
)

const (
	OBTAIN_LOCK_SCRIPT = `local lockClientId = redis.call('GET', KEYS[1])
if lockClientId == ARGV[1] then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return true
elseif not lockClientId then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
  return true
end
return false`

	DELETE_LOCK_SSCRIPT = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("DEL", KEYS[1])
	return true
else
    return false
end`
)

type RedisLock struct {
	store *redis.Client
	count int32
	key   string
	id    string
}

// Obtain returns a RedisLock
func Obtain(key string) *RedisLock {
	if gorest_redis.Rdb == nil {
		logger.Error("cannot find redis")
		return nil
	}
	return &RedisLock{
		store: gorest_redis.Rdb,
		key:   key,
		id:    uuid.NewString(),
	}
}

// TryLock
func (lock *RedisLock) TryLock(seconds int) (bool, error) {
	if seconds <= 0 {
		return false, errors.New("invalid seconds: " + strconv.Itoa(seconds))
	}
	newCount := atomic.AddInt32(&lock.count, 1)
	if newCount > 1 {
		return true, nil
	}
	resp := lock.store.Eval(context.Background(), OBTAIN_LOCK_SCRIPT, []string{lock.key},
		lock.id, strconv.Itoa(int(seconds)*1000))
	ok, err := resp.Bool()
	if err == redis.Nil {
		atomic.AddInt32(&lock.count, -1)
		return false, nil
	} else if err != nil {
		atomic.AddInt32(&lock.count, -1)
		logger.Error("Error on acquiring lock for %s, %s", lock.key, err.Error())
		return false, err
	} else if !ok {
		atomic.AddInt32(&lock.count, -1)
		return false, nil
	}

	return true, nil
}

// Unlock
func (lock *RedisLock) Unlock() (bool, error) {
	newCount := atomic.AddInt32(&lock.count, -1)
	if newCount > 0 {
		return true, nil
	}
	resp := lock.store.Eval(context.Background(), DELETE_LOCK_SSCRIPT, []string{lock.key}, []string{lock.id})
	ok, err := resp.Bool()
	if err != nil {
		return false, err
	}
	return ok, nil
}
