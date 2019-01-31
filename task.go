package task

import (
	"errors"

	"github.com/moisespsena/go-default-logger"
	"github.com/moisespsena/go-path-helpers"
)

var (
	pkg            = path_helpers.GetCalledDir()
	ErrTaskRunning = errors.New("Task is running")
	log            = defaultlogger.NewLogger(pkg)
)

type Task interface {
	Setup(appender Appender) error
	Run() error
	Start(done func()) (stop Stoper, err error)
}

type task struct {
	RunFunc   func() (err error)
	StopFunc  func()
	SetupFunc func(appender Appender) error
	running   bool
}

func (t *task) Run() (err error) {
	defer func() {
		t.running = false
	}()
	t.running = true
	return t.RunFunc()
}

func (t *task) SetSetup(s func(appender Appender) error) {
	t.SetupFunc = s
}

func (t *task) GetSetup() func(appender Appender) error {
	return t.SetupFunc
}

func (t *task) Start(done func()) (stop Stoper, err error) {
	if t.running {
		return nil, ErrTaskRunning
	}
	go func() {
		defer done()
		t.RunFunc()
	}()
	t.running = true
	return t, nil
}

func (t *task) Setup(appender Appender) (err error) {
	if t.SetupFunc != nil {
		return t.SetupFunc(appender)
	}
	return nil
}

func (t *task) Stop() {
	if t.running && t.StopFunc != nil {
		t.StopFunc()
	}
}

func (t *task) IsRunning() bool {
	return t.running
}

func NewTask(run func() (err error), stop ...func()) *task {
	var stopf func()
	if len(stop) > 0 {
		stopf = stop[0]
	}
	return &task{RunFunc: run, StopFunc: stopf}
}

type TaskProxy struct {
	T         Task
	SetupFunc func(appender Appender) error
	RunFunc   func() error
	StartFunc func(done func()) (stop Stoper, err error)
}

func (px *TaskProxy) Setup(appender Appender) error {
	if px.SetupFunc != nil {
		return px.SetupFunc(appender)
	}
	return nil
}

func (px *TaskProxy) Run() error {
	if px.RunFunc != nil {
		return px.RunFunc()
	}
	return nil
}

func (px *TaskProxy) Start(done func()) (stop Stoper, err error) {
	if px.StartFunc != nil {
		return px.StartFunc(done)
	}
	return
}

type TaskSetup func(appender Appender) error

func (tf TaskSetup) Setup(appender Appender) error {
	return tf(appender)
}

func (TaskSetup) Run() error {
	return nil
}

func (TaskSetup) Start(done func()) (stop Stoper, err error) {
	return nil, nil
}
