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

func Setup(t ...Task) (task Task, err error) {
	appender := appenderSetup{NewAppender()}
	if err = appender.AddTask(t...); err != nil {
		return
	}
	task = appender.Tasks()
	return
}

func Run(done func(), t ...Task) (s Stoper, err error) {
	if t, err := Setup(t...); err != nil {
		return nil, err
	} else {
		return t.Start(done)
	}
}

func Start(done func(), t ...Task) (s Stoper, err error) {
	if t, err := Setup(t...); err != nil {
		return nil, err
	} else {
		return t.Start(done)
	}
}
