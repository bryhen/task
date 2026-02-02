package task_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bryhen/task"
)

func TestContinuousTask(t *testing.T) {
	mt := &MyTask{}
	tsk := task.NewContinuousTask(mt, true, time.Second)
	mt.task = tsk

	err := tsk.Start(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		fmt.Println("Start at ", time.Now().UnixMilli())
		time.Sleep(time.Second*9 + time.Millisecond*500)
		tsk.Shutdown(t.Context())
		fmt.Println("Exit at ", time.Now().UnixMilli())
	}()

	rep, err := tsk.Report()
	if err != nil {
		t.Fatal(err)
	}

	_, err = tsk.Report()
	if err == nil {
		t.Fatal(fmt.Errorf("expected err"))
	}

	if rep.ErrSetup != nil {
		t.Log("Err Setup")
	}

	if rep.ErrDo != nil {
		t.Log("Err Do")
	}

	if rep.ErrTeardown != nil {
		t.Log("Err Teardown")
	}

	t.Log(rep)
}

type MyTask struct {
	task *task.Task[*task.ContinuousTaskReport[int]]
	i    int
}

func (mt *MyTask) Setup() error {
	fmt.Println("Setup", time.Now().UnixMilli())
	//	return fmt.Errorf("setup")
	return nil
}

func (mt *MyTask) Do() error {
	fmt.Println("Do", time.Now().UnixMilli())
	mt.i++

	if mt.i == 20 {
		mt.task.Done()
	}
	// return fmt.Errorf("err do")
	return nil
}

func (mt *MyTask) Teardown() (int, error) {
	fmt.Println("Teardown", time.Now().UnixMilli())
	// return 0, fmt.Errorf("err teradown")
	return 0, nil
}
