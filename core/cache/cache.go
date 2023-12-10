// 简单内存缓存
package cache

import (
	"container/heap"
	"log/slog"
	"time"

	"github.com/kappere/go-rest/core/tool/common"
)

type cacheItem struct {
	key    string
	value  interface{}
	expire time.Time
}

func (c cacheItem) Less(v common.PriorityQueueItem) bool {
	return c.expire.Compare(v.(cacheItem).expire) < 0
}

var (
	cacheMap   = make(map[string]cacheItem)
	cacheQueue = &common.PriorityQueue{}
)

func CachingPerm[V any](key string, defaultValueF func() V) V {
	return Caching(key, defaultValueF, time.Hour*24*365*99)
}

func Caching[V any](key string, defaultValueF func() V, d time.Duration) V {
	now := time.Now()
	if c, ok := cacheMap[key]; ok && c.expire.After(now) {
		return c.value.(V)
	}
	value := defaultValueF()
	c := cacheItem{
		key,
		value,
		time.Now().Add(d),
	}
	cacheMap[key] = c
	heap.Push(cacheQueue, c)
	return value
}

func Invalidate(key string) {
	delete(cacheMap, key)
}

func init() {
	go func() {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						slog.Error("Cache clear routine panic.", "error", err)
					}
				}()
				time.Sleep(1 * time.Minute)
				now := time.Now()
				for cacheQueue.Len() > 0 {
					c := (*cacheQueue)[0].(cacheItem)
					if now.After(c.expire) {
						heap.Pop(cacheQueue)
						delete(cacheMap, c.key)
					} else {
						break
					}
				}
			}()
		}
	}()
}
