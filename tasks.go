package task

import (
	"context"

	"github.com/moisespsena-go/logging"
)

type Slice []Task

func (tasks Slice) Setup(appender Appender) (err error) {
	for _, t := range tasks {
		if s, ok := t.(TaskSetupAppender); ok {
			if err = s.Setup(appender); err != nil {
				return
			}
		} else if s, ok := t.(TaskSetuper); ok {
			if err = s.Setup(); err != nil {
				return
			}
		}
	}
	return nil
}

func (tasks Slice) Run() (err error) {
	return MustRun(nil, tasks...)
}

func (tasks Slice) Start(done func()) (s Stoper, err error) {
	return MustStart(done, tasks...)
}

type PreRunSlice []func(ap Appender) error
type PostRunSlice []func()

type Appender interface {
	SliceGetter
	AddTask(t ...Task) error
	PostRun(f ...func())
	Context() context.Context
	WithContext(ctx context.Context) (done func())
}

type PreRunCallback interface {
	TaskPreRun(ta Appender) error
}

type PostRunCallback interface {
	TaskPosRun()
}

type Tasks struct {
	TaskAppender
	log          logging.Logger
	preRun       []func(ta Appender) error
	preRunCalled bool
}

func (tasks *Tasks) Log() logging.Logger {
	return tasks.log
}

func (tasks *Tasks) SetLog(log logging.Logger) {
	tasks.log = log
}

func (tasks *Tasks) PreRun(f ...func(ta Appender) error) {
	tasks.preRun = append(tasks.preRun, f...)
}

func (tasks *Tasks) copyAppender() *TaskAppender {
	return &TaskAppender{
		tasks:   append([]Task{}, tasks.tasks...),
		postRun: tasks.postRun,
	}
}

func (tasks *Tasks) GetPreRun() []func(ta Appender) error {
	return tasks.preRun
}

func (tasks *Tasks) Run() (err error) {
	ts := tasks.copyAppender()
	if err = tasks.TaskPreRun(ts); err != nil {
		return
	}
	defer tasks.TaskPostRun()
	return ts.tasks.Run()
}

func (tasks *Tasks) Start(done func()) (stop Stoper, err error) {
	ts := tasks.copyAppender()
	if err = tasks.TaskPreRun(ts); err != nil {
		return
	}

	if stop, err = ts.tasks.Start(func() {
		ts.TaskPostRun()
		done()
	}); err != nil {
		if tasks.log != nil {
			tasks.log.Error(err.Error())
		} else {
			log.Error(err.Error())
		}
	}
	return
}

func (tasks *Tasks) TaskPreRun(ts *TaskAppender) (err error) {
	if !tasks.preRunCalled {
		tasks.preRunCalled = true
		for _, pr := range tasks.preRun {
			if err = pr(ts); err != nil {
				return err
			}
		}
	}

	return nil
}

type TaskAppender struct {
	tasks   Slice
	setup   []func(ta Appender) error
	postRun []func()
	context context.Context
}

func (this *TaskAppender) Context() context.Context {
	return this.context
}

func (this *TaskAppender) WithContext(ctx context.Context) (done func()) {
	old := this.context
	this.context = ctx
	return func() {
		this.context = old
	}
}

func (this *TaskAppender) Tasks() Slice {
	return this.tasks
}

func (this *TaskAppender) AddSetup(s ...func(ta Appender) error) {
	this.setup = append(this.setup, s...)
}

func (this *TaskAppender) Setup(tar Appender) (err error) {
	for _, s := range this.setup {
		if err = s(tar); err != nil {
			return
		}
	}
	return setup(this, func(t ...Task) error {
		return nil
	}, this.tasks...)
}

func (this *TaskAppender) AddTask(t ...Task) (err error) {
	for _, t := range t {
		switch tt := t.(type) {
		case Slice:
			err = this.AddTask(tt...)
		case SliceGetter:
			err = this.AddTask(tt.Tasks()...)
		default:
			this.tasks = append(this.tasks, t)
		}
		if err != nil {
			return
		}
	}
	return
}

func (this *TaskAppender) PostRun(f ...func()) {
	this.postRun = append(this.postRun, f...)
}

func (this *TaskAppender) TaskPostRun() {
	for _, pr := range this.postRun {
		pr()
	}
}

func NewAppender() Appender {
	return &TaskAppender{}
}

type SliceGetter interface {
	Tasks() Slice
}
