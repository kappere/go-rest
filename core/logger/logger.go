package logger

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/kappere/go-rest/core/config/conf"
	"github.com/kappere/go-rest/core/task"
)

func InitLogger(logConfig conf.LogConfig, appName string) {
	var prevLogFile *os.File = nil
	os.MkdirAll(logConfig.Path, os.ModeDir)
	archiveLogFunc := func() {
		now := time.Now()
		filename := fmt.Sprintf(logConfig.Path+"/%s_%04d%02d%02d.log", appName, now.Year(), now.Month(), now.Day())
		logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)
		if err != nil {
			slog.Error("Create log file failed", "error", err)
		}
		newWriter := io.MultiWriter(os.Stdout, logFile)

		// 重定向日志到新文件
		log.SetOutput(newWriter)

		if prevLogFile != nil {
			prevLogFile.Close()
		}
		prevLogFile = logFile
	}
	archiveLogFunc()
	task.NewTaskFunc("0 0 0 * * ?", "LogArchive", archiveLogFunc)
}
