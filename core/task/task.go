// 只是一个定时任务的例子（参照用），创建项目后请删除该文件
//
// 使用文档：https://pkg.go.dev/github.com/robfig/cron
// Field name   | Mandatory? | Allowed values  | Allowed special characters
// ----------   | ---------- | --------------  | --------------------------
// Seconds      | Yes        | 0-59            | * / , -
// Minutes      | Yes        | 0-59            | * / , -
// Hours        | Yes        | 0-23            | * / , -
// Day of month | Yes        | 1-31            | * / , - ?
// Month        | Yes        | 1-12 or JAN-DEC | * / , -
// Day of week  | Yes        | 0-6 or SUN-SAT  | * / , - ?
package task

import (
	"github.com/kappere/go-rest/core/logger"
	"github.com/robfig/cron"
)

type Task interface {
	Process()
}

var c = cron.New()

func NewTask(cron string, name string, t Task) {
	c.AddFunc(cron, func() {
		defer func() {
			logger.Info("==== Task [%s] finished ====", name)
		}()
		logger.Info("==== Task [%s] start ====", name)
		t.Process()
	})
}

func NewTaskFunc(cron string, name string, t func()) {
	c.AddFunc(cron, func() {
		defer func() {
			logger.Info("==== Task [%s] finished ====", name)
		}()
		logger.Info("==== Task [%s] start ====", name)
		t()
	})
}

func Start() {
	c.Start()
}
