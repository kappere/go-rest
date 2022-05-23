package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	gormLoggerPkg "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// v[0] level
// v[1] file
// v[2] time
// v[3] sql
func (l *GormLogger) Print(v ...interface{}) {
	logSql("[%s] %s %v %s", v...)
}

// v[0] level
// v[1] file
// v[2] time
// v[3] sql
func (l *GormLogger) Printf(format string, v ...interface{}) {
	Log("[SQL]   ", -1, format, v...)
}

// ErrRecordNotFound record not found error
var ErrRecordNotFound = errors.New("record not found")

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// Config GormLogger config
type Config struct {
	SlowThreshold             time.Duration
	Colorful                  bool
	IgnoreRecordNotFoundError bool
	LogLevel                  LogLevel
}

var (
	// Discard Discard GormLogger will print any log to ioutil.Discard
	Discard = NewGormLogger(Config{})
	// Default Default GormLogger
	DefaultGormLogger = NewGormLogger(Config{
		SlowThreshold:             500 * time.Millisecond, // 慢 SQL 阈值
		LogLevel:                  InfoLevel,              // 日志级别
		IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
		Colorful:                  false,                  // 禁用彩色打印
	})
)

// New initialize GormLogger
func NewGormLogger(config Config) *GormLogger {
	var (
		infoStr      = "[%s]: [info] "
		warnStr      = "[%s]: [warn] "
		errStr       = "[%s]: [error] "
		traceStr     = "[%s]: [%.3fms] [rows:%v] %s"
		traceWarnStr = "[%s]: %s [%.3fms] [rows:%v] %s"
		traceErrStr  = "[%s]: %s [%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "[%s]: " + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "[%s]: " + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "[%s]: " + Reset + Red + "[error] " + Reset
		traceStr = Green + "[%s]: " + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "[%s]: " + Yellow + "%s " + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "[%s]: " + MagentaBold + "%s " + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}

	return &GormLogger{
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type GormLogger struct {
	Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *GormLogger) LogMode(level gormLoggerPkg.LogLevel) gormLoggerPkg.Interface {
	newlogger := *l
	// newlogger.LogLevel = LogLevel(int(level))
	return &newlogger
}

// Info print info
func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= InfoLevel {
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= WarnLevel {
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= ErrorLevel {
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= PanicLevel {
		return
	}

	elapsed := time.Since(begin)
	f := utils.FileWithLineNum()
	f = f[len(f)-30:]
	switch {
	case err != nil && l.LogLevel >= ErrorLevel && (!errors.Is(err, ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, f, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceErrStr, f, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= WarnLevel:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, f, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, f, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == InfoLevel:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, f, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, f, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
