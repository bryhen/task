package task

import (
	"context"
	"fmt"
	"sync"
)

var (
	ErrNotStarted = fmt.Errorf("not started")
	ErrDupStart   = fmt.Errorf("duplicate call to start")

	ErrWasShutdown = fmt.Errorf("previously shutdown")

	ErrDupPayload = fmt.Errorf("duplicate call to payload")
)

type Tasker[T any] interface {
	// An infinite task could be a for loop that checks KeepRunning() as it's exit condition (ie 'for task.KeepRunning() { log.Println("do work") }')
	//
	// A finite task would also check KeepRunning() and additionally whatever necessary when it is progressing (ie 'for task.KeepRunning() && moreToDo { log.Println("Doing next") }')
	//
	// For manual checking of exit conditions, ExitCh() will return a channel that is closed when the task shuold exit.
	Do() T
}

// Should not be created directly.
//
// Must use NewTask()
type Task[T any] struct {
	mu *sync.Mutex

	t    Tasker[T]
	done chan struct{}

	exit chan struct{}

	notify chan T

	exitClosed bool
	started    bool
	notified   bool
}

// May only be called once.
//
// Calls tasker.Do() in a separate goroutine.
//
// When your tasker's Do() returns, its payload will be available through Payload() (Payload() also blocks and can be used to await task completion).
func (t *Task[T]) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.exitClosed {
		return ErrWasShutdown
	}

	if t.started {
		return ErrDupStart
	}

	t.started = true

	go func() {
		t.notify <- t.t.Do()
		close(t.notify)
		close(t.done)
	}()

	return nil
}

// Called by Tasks when they're  done. This allows a task to exit (if it hasn't already) by checking KeepRunning() or <-ExitCh().
//
// It is safe to call both Done() and Shutdown()
func (t *Task[T]) Done() {
	t.mu.Lock()
	if !t.exitClosed {
		close(t.exit)
		t.exitClosed = true
	}
	t.mu.Unlock()
}

// Shuts down a task.
//
// If a task has already returned its payload it is not required to call shutdown.
//
// Shutdown is meant to help gracefully shutdown currently running tasks.
//
// Task's Do() should NEVER call Shutdown() itself (synchronously, at least) as this will cause indefinite locking as Shutdown() waits for Do() to finish.
// Do() should call Done() if it is complete.
//
// It is safe to call both Done() and Shutdown().
func (t *Task[T]) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	if !t.started {
		t.mu.Unlock()
		return ErrNotStarted
	}

	if !t.exitClosed {
		close(t.exit)
		t.exitClosed = true
	}
	t.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.done:
		return nil
	}
}

// This condition can be checked to determine whether the task should exit.
//
// For manual observation of exit conditions, get the channel from ExitCh() which will be closed when the task's Do() should exit.
func (t *Task[T]) KeepRunning() bool {
	select {
	default:
		return true
	case <-t.exit:
		return false
	}
}

// Allows for manual observation of exit signals.
func (t *Task[T]) ExitCh() <-chan struct{} {
	return t.exit
}

// Blocks until the task finishes and provides its payload.
//
// This function will ONLY provide the payload once. Subsequent calls will return an error.
func (t *Task[T]) Payload() (T, error) {
	t.mu.Lock()
	if t.notified {
		return *new(T), ErrDupPayload
	}
	t.notified = true
	t.mu.Unlock()

	return <-t.notify, nil
}

func NewTask[T any](proc Tasker[T]) *Task[T] {
	return &Task[T]{
		mu:     &sync.Mutex{},
		t:      proc,
		done:   make(chan struct{}),
		exit:   make(chan struct{}),
		notify: make(chan T, 1),
	}
}
