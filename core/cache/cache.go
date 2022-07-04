// 简单内存缓存
package cache

var cacheMap = make(map[string]interface{})

func Processing[V any](key string, defaultValueF func() V) V {
	if value, ok := cacheMap[key]; ok {
		return value.(V)
	}
	value := defaultValueF()
	cacheMap[key] = value
	return value
}

func Invalidate(key string) {
	delete(cacheMap, key)
}
