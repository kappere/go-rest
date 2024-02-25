package task

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTaskFunc(t *testing.T) {
	type args struct {
		cron string
		name string
		t    func()
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "task1",
			args: args{
				cron: "0/3 * * * * ?",
				name: "task2",
				t: func() {
					fmt.Println("Task12 is Running!!!")
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewTaskFunc(tt.args.cron, tt.args.name, tt.args.t)
		})
	}
	time.Sleep(20 * time.Second)
}
