package task

import (
	"sync"

	"github.com/op/go-logging"
)

type Slice []Task

func (tasks Slice) Setup(appender Appender) (err error) {
	for _, t := range tasks {
		if err = t.Setup(appender); err != nil {
			return
		}
	}
	return nil
}

func (tasks Slice) Prepare() (ts Slice, stop Stopers, wg *sync.WaitGroup, postRun func(), err error) {
	if len(tasks) == 0 {
		return
	}

	wg = &sync.WaitGroup{}

	var (
		ta       = &TaskAppender{}
		postRunS PostRunSlice
		s        Stoper
		i        int
		t        Task
	)

	for i, t = range tasks {
		if pr, ok := t.(PreRunCallback); ok {
			if err = pr.TaskPreRun(ta); err != nil {
				return
			}
		}
		if pr, ok := t.(PostRunCallback); ok {
			postRunS = append(postRunS, pr.TaskPosRun)
		}
	}

	postRun = func() {
		for _, pr := range postRunS {
			pr()
		}
	}

	tasks = append(tasks, ta.tasks...)
	stop = make(Stopers, len(tasks))

	wg.Add(len(tasks))

	for i, t = range tasks {
		if s, err = t.Start(wg.Done); err != nil {
			for _, s = range stop[0:i] {
				s.Stop()
			}
			break
		} else {
			stop[i] = s
		}
	}

	ts = tasks
	return
}

func (tasks Slice) Run() (err error) {
	_, _, wg, postRun, err := tasks.Prepare()
	if err != nil {
		return
	}
	defer postRun()
	wg.Wait()
	return
}

func (tasks Slice) StartOnly(wg *sync.WaitGroup, postRun, done func()) {
	if len(tasks) == 0 {
		done()
		return
	}

	go func() {
		defer func() {
			done()
		}()
		defer postRun()
		wg.Wait()
	}()
}

func (tasks Slice) Start(done func()) (s Stoper, err error) {
	if len(tasks) == 0 {
		done()
		return &FakeStoper{}, nil
	}

	tasks, stop, wg, postRun, err := tasks.Prepare()
	if err != nil {
		return nil, err
	}

	tasks.StartOnly(wg, postRun, done)
	return stop, nil
}

type PreRunSlice []func(ap Appender) error
type PostRunSlice []func()

type Appender interface {
	AddTask(t ...Task) error
	PostRun(f ...func())
	Tasks() Slice
}

type PreRunCallback interface {
	TaskPreRun(ta Appender) error
}

type PostRunCallback interface {
	TaskPosRun()
}

type Tasks struct {
	TaskAppender
	log          *logging.Logger
	preRun       []func(ta Appender) error
	preRunCalled bool
}

func (tasks *Tasks) Log() *logging.Logger {
	return tasks.log
}

func (tasks *Tasks) SetLog(log *logging.Logger) {
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
			tasks.log.Error(err)
		} else {
			log.Error(err)
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
}

func (tasks *TaskAppender) Tasks() Slice {
	return tasks.tasks
}

func (ta *TaskAppender) AddSetup(s ...func(ta Appender) error) {
	ta.setup = append(ta.setup, s...)
}

func (ta *TaskAppender) Setup(tar Appender) (err error) {
	for _, s := range ta.setup {
		if err = s(tar); err != nil {
			return
		}
	}
	for _, t := range ta.tasks {
		if err = t.Setup(tar); err != nil {
			return
		}
	}
	return nil
}

func (ta *TaskAppender) AddTask(t ...Task) error {
	ta.tasks = append(ta.tasks, t...)
	return nil
}

func (ta *TaskAppender) PostRun(f ...func()) {
	ta.postRun = append(ta.postRun, f...)
}

func (ta *TaskAppender) TaskPostRun() {
	for _, pr := range ta.postRun {
		pr()
	}
}

func NewAppender() Appender {
	return &TaskAppender{}
}
