package rest

import (
	"fmt"
	"log/slog"
	"math"
	"runtime"
	"time"
)

func setupMonitor() {
	// 监控服务信息
	go func() {
		for {
			collectStatisticInfo()
			time.Sleep(60 * time.Second)
		}
	}()
}

type Stat struct {
	Routine int
	Memory  uint64
	Heap    uint64
	Stack   uint64
	Time    time.Time
}

var prevStat Stat

const STAT_THRESHOLD float64 = 0.1

func (s Stat) statExpire(currentStat Stat) bool {
	return int(time.Now().Unix()-s.Time.Unix()) > 3600 || (s.Routine != currentStat.Routine ||
		math.Abs(float64(s.Memory)-float64(currentStat.Memory))/float64(s.Memory) > STAT_THRESHOLD ||
		math.Abs(float64(s.Heap)-float64(currentStat.Heap))/float64(s.Heap) > STAT_THRESHOLD ||
		math.Abs(float64(s.Stack)-float64(currentStat.Stack))/float64(s.Stack) > STAT_THRESHOLD)
}

func collectStatisticInfo() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Collect statistic info failed:", "error", err)
		}
	}()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	currentStat := Stat{
		Routine: runtime.NumGoroutine(),
		Memory:  m.Sys,
		Heap:    m.HeapSys,
		Stack:   m.StackSys,
		Time:    time.Now(),
	}
	if prevStat.statExpire(currentStat) {
		prevStat = currentStat
		slog.Info(fmt.Sprintf("Stat: num_goroutine=%d, memory=%dm, heap=%dm, stack=%dm",
			currentStat.Routine,
			currentStat.Memory/1024/1024,
			currentStat.Heap/1024/1024,
			currentStat.Stack/1024/1024))
	}
}
