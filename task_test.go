package task_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bryhen/task"
)

func TestContinuousTask(t *testing.T) {
	mt := &MyTask{}
	tsk := task.NewIntervalTask(mt, 0, true)
	mt.task = tsk

	repCh, err := tsk.Start(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * 10)
		tsk.Shutdown(t.Context())
	}()

	rep := <-repCh

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
	fmt.Println("Setup")
	//	return fmt.Errorf("setup")
	return nil
}

func (mt *MyTask) Do() error {
	fmt.Println("Do")
	mt.i++

	if mt.i == 20 {
		mt.task.Done()
	}
	return fmt.Errorf("err do")
	// return nil
}

func (mt *MyTask) Teardown() (int, error) {
	fmt.Println("Teardown")
	return 0, fmt.Errorf("err teradown")
	//	return 0, nil
}
