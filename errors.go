package task

import (
	"errors"
	"fmt"
)

type ErrNotRunning struct {
	Task Task
}

func (this ErrNotRunning) Error() string {
	return fmt.Sprintf("task %T not running", this.Task)
}

var ErrNoTasks = errors.New("no tasks")


