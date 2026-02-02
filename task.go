package task

import (
	"context"
	"fmt"
	"sync"
)

var (
	ErrNotStarted = fmt.Errorf("not started")
	ErrDupStart   = fmt.Errorf("duplicate start")

	ErrWasShutdown = fmt.Errorf("already shutdown")
	ErrDupShutdown = fmt.Errorf("duplicate shutdown")
)

type Tasker[T any] interface {
	// An infinite task could be a for loop that checks KeepRunning() as it's exit condition (ie 'for task.KeepRunning(){ log.Println("do work") }')
	//
	// A finite task would also check KeepRunning() and additionally whenever it is progressing (ie 'for task.KeepRunning() && moreToDo { log.Println("Doing next") }')
	Do() T
}

// Should not be created directly.
//
// Must use NewTask()
type Task[T any] struct {
	mu *sync.Mutex

	t      Tasker[T]
	done   chan struct{}
	notify chan T

	started  bool
	shutdown bool
}

// May only be called once.
//
// Calls tasker.Do() in a separate goroutine.
//
// When your tasker's Do() returns, its TaskReport will be sent over the channel returned here.
func (t *Task[T]) Start(ctx context.Context) (<-chan T, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.shutdown {
		return nil, ErrWasShutdown
	}

	if t.started {
		return nil, ErrDupStart
	}

	t.started = true

	go func() {
		t.notify <- t.t.Do()
		close(t.notify)
		close(t.done)
	}()

	return t.notify, nil
}

// Shuts down a task.
//
// If a task has already returned its *TaskReport via the channel returned by Start(),
// it is not required to call shutdown.
//
// Shutdown is meant to help gracefully shutdown currently running tasks.
func (t *Task[T]) Shutdown(ctx context.Context) error {
	t.mu.Lock()

	if !t.started {
		t.mu.Unlock()
		return ErrNotStarted
	} else if t.shutdown {
		t.mu.Unlock()
		return ErrDupShutdown
	}

	t.shutdown = true
	t.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.done:
		return nil
	}
}

func (t *Task[T]) KeepRunning() bool {
	return !t.shutdown
}

func NewTask[T any](proc Tasker[T]) *Task[T] {
	return &Task[T]{
		mu:     &sync.Mutex{},
		t:      proc,
		done:   make(chan struct{}),
		notify: make(chan T, 1),
	}
}
