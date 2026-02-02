package task_test

import (
	"_/C_/Users/bryce/projects/go-tools/task"
	"fmt"
)

func TestTask() {
	t := task.NewTask()
}

type MyTask struct{}

func (mt *MyTask) Do() {
	fmt.Println("Hi")
}
