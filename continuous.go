package task

import "time"

type Stepper[T any] interface {
	Setup() error
	Do() error
	Teardown() (T, error) // Can return any relevant data here
}

type continuous[T any] struct {
	task *Task[*ContinuousTaskReport[T]]

	interval time.Duration
	steps    Stepper[T]

	exitOnErr bool
}

type ContinuousTaskReport[T any] struct {
	Payload     T
	ErrSetup    error
	ErrDo       error // Will not be recorded unless exitOnErr is true.
	ErrTeardown error
}

// Performs the steps in order.
// Calls Setup(), if an error occurs, returns early without running Do().
// Calls Do() at each interval (or no interval if the interval is 0). After Shutdown() or Done() is called by the parent task, it calls Teardown(), writing and returning data in the returned report.
func (c *continuous[T]) Do() *ContinuousTaskReport[T] {
	ctr := &ContinuousTaskReport[T]{}

	if ctr.ErrSetup = c.steps.Setup(); ctr.ErrSetup != nil {
		return ctr
	}

	// Only create ticker if necessary
	if c.interval > 0 {
		tck := time.NewTicker(c.interval)

		// Only check exitOnErr at top
		if c.exitOnErr {
			for ; c.task.KeepRunning(); <-tck.C {
				if err := c.steps.Do(); err != nil {
					ctr.ErrDo = err
					break
				}
			}
		} else {
			for ; c.task.KeepRunning(); <-tck.C {
				c.steps.Do()
			}
		}
		tck.Stop()

	} else {

		// Only check exitOnErr at top
		if c.exitOnErr {
			for c.task.KeepRunning() {
				if err := c.steps.Do(); err != nil {
					ctr.ErrDo = err
					break
				}
			}
		} else {
			for c.task.KeepRunning() {
				c.steps.Do()
			}
		}
	}

	ctr.Payload, ctr.ErrTeardown = c.steps.Teardown()

	return ctr
}

func NewIntervalTask[T any](steps Stepper[T], interval time.Duration, exitOnErr bool) *Task[*ContinuousTaskReport[T]] {
	ctr := &continuous[T]{
		interval:  interval,
		steps:     steps,
		exitOnErr: exitOnErr,
	}

	tsk := NewTask(ctr)
	ctr.task = tsk

	return tsk
}
