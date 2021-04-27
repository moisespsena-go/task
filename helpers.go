package task

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type OnDoneEvent struct {
	onDone []func()
}

func (od *OnDoneEvent) OnDone(f ...func()) {
	od.onDone = append(od.onDone, f...)
}

func (od *OnDoneEvent) DoneFuncs() []func() {
	return od.onDone
}

func (od *OnDoneEvent) CallDoneFuncs() {
	for _, f := range od.onDone {
		if f != nil {
			f()
		}
	}
}

type TaskSetupError struct {
	Task  Task
	Index int
	cause error
}

func (se *TaskSetupError) Cause() error {
	return se.cause
}

func (se *TaskSetupError) Error() string {
	return fmt.Sprintf("Task #%d %T setup failed: %v", se.Index, se.Task, se.cause)
}

func setup(appender Appender, add func(t ...Task) error, t ...Task) (err error) {
	if add == nil {
		add = appender.AddTask
	}
	for i, t := range t {
		if t == nil {
			continue
		}
		switch s := t.(type) {
		case TaskSetuper:
			if err = s.Setup(); err != nil {
				return errors.Wrap(&TaskSetupError{t, i, err}, "task setup")
			}
		case TaskSetupAppender:
			if err = s.Setup(appender); err != nil {
				return errors.Wrap(&TaskSetupError{t, i, err}, "task setup")
			}
		}
		if err = add(t); err != nil {
			return errors.Wrapf(err, "add task %d", i)
		}
	}
	return
}

type appenderSetup struct {
	Appender
}

func (this *appenderSetup) AddTask(t ...Task) (err error) {
	return setup(this, this.Appender.AddTask, t...)
}

func MustRun(done func(), t ...Task) error {
	pt, err := Prepare(t...)
	if err != nil {
		return err
	}
	d := make(chan interface{})
	if _, err = pt.Start(func() {
		close(d)
		if done != nil {
			done()
		}
	}); err == nil {
		<-d
	}
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
	if pt, err = Prepare(t...); err != nil {
		return
	}
	var state *State
	if state, err = pt.Start(func() {
		if done != nil {
			done(state)
		}
	}); err != nil {
		return
	}
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
		close(d)
		if done != nil {
			done(state)
		}
	}); err == nil {
		<-d
	}
	return
}

func StopWait(stoper Stoper) {
	if sw, ok := stoper.(StopWaiter); ok {
		sw.StopWait()
		return
	}

	for stoper.IsRunning() {
		stoper.Stop()
		time.Sleep(100 * time.Millisecond)
	}
}
