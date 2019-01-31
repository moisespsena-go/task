package task

import (
	"fmt"

	"github.com/moisespsena/go-error-wrap"
)

type TaskSetupError struct {
	Task  Task
	Index int
	Err   error
}

func (se TaskSetupError) Error() string {
	return fmt.Sprintf("Task #%d %s setup failed", se.Index, se.Task)
}

type appenderSetup struct {
	Appender
}

func (ap *appenderSetup) AddTask(t ...Task) (err error) {
	for i, t := range t {
		if err = t.Setup(ap); err != nil {
			return errwrap.Wrap(err, &TaskSetupError{t, i, err})
		}
		if err = ap.Appender.AddTask(t); err != nil {
			return errwrap.Wrap(err, "Add task %d", i)
		}
	}
	return
}

func MustRun(done func(), t ...Task) error {
	pt, err := Prepare(t...)
	if err != nil {
		return err
	}
	d := make(chan interface{})
	if _, err = pt.Start(func() {
		d <- nil
		done()
	}); err == nil {
		<-d
	}
	close(d)
	return err
}

func MustStart(done func(), t ...Task) (s Stoper, err error) {
	var pt *PreparedTasks
	pt, err = Prepare(t...)
	if err != nil {
		return
	}
	return pt.Start(done)
}

func Start(done func(state *State), t ...Task) (s Stoper, err error) {
	var pt *PreparedTasks
	pt, err = Prepare(t...)
	if err != nil {
		return
	}
	var state *State
	state, err = pt.Start(func() {
		done(state)
	})
	s = state
	return
}

func Run(done func(state *State), t ...Task) (err error) {
	var pt *PreparedTasks
	pt, err = Prepare(t...)
	if err != nil {
		return
	}

	var state *State
	d := make(chan interface{})
	if state, err = pt.Start(func() {
		d <- nil
		done(state)
	}); err == nil {
		<-d
	}
	close(d)
	return
}
