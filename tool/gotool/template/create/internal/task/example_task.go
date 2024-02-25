// 定时任务示例
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

/*
import (
	"log/slog"

	"github.com/kappere/go-rest/core/task"
)

// task with function
func RunExampleTask1() {
	slog.Info("ExampleTask1 is running")
}

// task with struct
type ExampleTask2 struct {
}

func (t *ExampleTask2) Process() {
	slog.Info("ExampleTask1 is running")
}

func init() {
	task.NewTaskFunc("0 0/1 * * * ?", "ExampleTask1", RunExampleTask1)
	task.NewTask("0 0/1 * * * ?", "ExampleTask2", &ExampleTask2{})
}
*/
