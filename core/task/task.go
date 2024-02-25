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
	"log"
	"runtime"

	"github.com/google/uuid"
	"github.com/robfig/cron"
)

type Task interface {
	Process()
}

var c = cron.New()

var task_status = make(map[string]bool)

func NewTask(cron string, name string, t Task) {
	NewTaskFunc(cron, name, t.Process)
}

func NewTaskFunc(cron string, name string, t func()) {
	task_id := uuid.NewString()
	task_status[task_id] = false
	c.AddFunc(cron, func() {
		if checkPreviousTaskStatus(task_id) {
			log.Printf("==== Task [%s] previous version was running ====", name)
			return
		}
		task_status[task_id] = true
		defer func() {
			task_status[task_id] = false
			if r := recover(); r != nil {
				log.Printf("==== Task [%s] failed ====", name)
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("cron: panic running job: %v\n%s", r, buf)
			} else {
				log.Printf("==== Task [%s] finished ====", name)
			}
		}()
		log.Printf("==== Task [%s] start ====", name)
		t()
	})
}

func checkPreviousTaskStatus(task_id string) bool {
	status, exist := task_status[task_id]
	return exist && status
}

func init() {
	c.Start()
}
