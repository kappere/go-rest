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

/*
import (
	"wataru.com/go-rest/core/logger"
	"wataru.com/go-rest/core/task"
)

// task with function
func RunExampleTask1() {
	logger.Info("ExampleTask1 is running")
}

// task with struct
type ExampleTask2 struct {
}

func (t *ExampleTask2) Process() {
	logger.Info("ExampleTask2 is running")
}

func init() {
	task.NewTaskFunc("0 0/1 * * * *", "ExampleTask1", RunExampleTask1)
	task.NewTask("0 0/1 * * * *", "ExampleTask2", &ExampleTask2{})
}
*/
