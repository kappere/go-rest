package server

import (
	"fmt"
	"runtime"
	"time"

	"wataru.com/go-rest/core/logger"
)

func setupMonitor() {
	// 监控服务信息
	go func() {
		for {
			time.Sleep(60 * time.Second)
			collectStatisticInfo()
		}
	}()
}

type Stat struct {
	Content string
	Time    time.Time
}

var prevStat Stat

func (s Stat) statExpire(currentStatContent string) bool {
	return int(time.Now().Unix()-s.Time.Unix()) > 3600 || s.Content != currentStatContent
}

func collectStatisticInfo() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("%v", err)
		}
	}()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stat := fmt.Sprintf("stat: [num_goroutine=%d, memory=%dm, heap=%dm, stack=%dm]",
		runtime.NumGoroutine(), m.Sys/1024/1024, m.HeapSys/1024/1024, m.StackSys/1024/1024)
	if prevStat.statExpire(stat) {
		prevStat = Stat{
			Content: stat,
			Time:    time.Now(),
		}
		logger.Info(stat)
	}
}
