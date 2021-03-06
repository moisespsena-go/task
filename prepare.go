package task

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

type PreStartCallback interface {
	PreTaskStart(s *State)
}

type PostStartCallback interface {
	PostTaskStart(s *State)
}

type taskStoper struct {
	Task
	Stoper
	key      uint64
	doneChan chan uint64
	postRun  []func()
}

func (ts taskStoper) doneNotify() {
	ts.doneChan <- ts.key
	if pr, ok := ts.Task.(PostRunCallback); ok {
		pr.TaskPosRun()
	}
}

type State struct {
	OnDoneEvent
	Start       time.Time
	End         time.Time
	mu          sync.Mutex
	i           uint64
	taskStopers map[interface{}]*taskStoper
	done        chan uint64
	stopCalled  bool
}

func (s *State) Add(tasks ...Task) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskStopers == nil {
		s.taskStopers = map[interface{}]*taskStoper{}
	}

	var (
		items []*taskStoper
		add   func(tasks ...Task) (err error)
		ta    = &TaskAppender{}
		addt  = func(t Task) {
			items = append(items, &taskStoper{
				Task:     t,
				key:      s.i,
				doneChan: s.done,
			})
			s.i++
		}
	)
	add = func(tasks ...Task) (err error) {
		for _, t := range tasks {
			if f, ok := t.(Factory); ok {
				t = f.Factory()
				if pt, err := Prepare(t); err != nil {
					return err
				} else if err = add(pt.tasks...); err != nil {
					return err
				}
			} else {
				addt(t)
			}
		}
		return
	}

	if err = add(tasks...); err != nil {
		return
	}

	for _, t := range items {
		if pr, ok := t.Task.(PreRunCallback); ok {
			if err = pr.TaskPreRun(ta); err != nil {
				return
			}
		}
	}

	for _, t := range ta.tasks {
		addt(t)
	}

	for _, t := range items {
		if cb, ok := t.Task.(PreStartCallback); ok {
			cb.PreTaskStart(s)
		}

		if stop, err := t.Start(t.doneNotify); err != nil {
			s.stop()
			return err
		} else if validStoper(stop) {
			t.Stoper = stop
			s.taskStopers[t.key] = t

			if cb, ok := t.Task.(PostStartCallback); ok {
				cb.PostTaskStart(s)
			}
		} else {
			s.stop()
			return &ErrNotRunning{t}
		}
	}
	return
}

func (s *State) Wait() {
	for len(s.taskStopers) > 0 {
		key := <-s.done
		delete(s.taskStopers, key)
	}
	s.End = time.Now()
	s.CallDoneFuncs()
}

func (s *State) stop() {
	var newStoppers = map[interface{}]*taskStoper{}
	for k, t := range s.taskStopers {
		if t.IsRunning() {
			t.Stop()
			if t.IsRunning() {
				newStoppers[k] = t
			}
		}
	}
	s.taskStopers = newStoppers
}

func (s *State) Stop() {
	if s.stopCalled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopCalled = true
	if s.taskStopers == nil {
		return
	}
	s.stop()
}

func (s *State) StopWait() {
	if s.stopCalled {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopCalled = true
	if s.taskStopers == nil {
		return
	}
	for len(s.taskStopers) > 0 {
		s.stop()
		time.Sleep(100 * time.Millisecond)
	}
}

func (s State) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskStopers != nil {
		for _, v := range s.taskStopers {
			if v.IsRunning() {
				return true
			}
		}
	}
	return false
}

func (s State) Len() int {
	return len(s.taskStopers)
}

func (s State) Tasks() (tasks Slice) {
	if s.taskStopers != nil {
		for _, t := range s.taskStopers {
			tasks = append(tasks, t)
		}
	}
	return
}

func (s State) Stopers() (stopers Stopers) {
	if s.taskStopers != nil {
		for _, t := range s.taskStopers {
			stopers = append(stopers, t)
		}
	}
	return
}

type PreparedTasks struct {
	tasks Slice
}

func (this *PreparedTasks) Tasks() Slice {
	return this.tasks
}

func Prepare(t ...Task) (pt *PreparedTasks, err error) {
	if len(t) == 0 {
		return nil, errors.Wrap(ErrNoTasks, "prepare")
	}
	appender := appenderSetup{NewAppender()}
	if err = appender.AddTask(t...); err != nil {
		return
	}
	pt = &PreparedTasks{appender.Tasks()}
	return
}

func (this *PreparedTasks) Start(done ...func()) (state *State, err error) {
	state = &State{
		Start: time.Now(),
		done:  make(chan uint64),
	}
	state.OnDone(done...)
	if err = state.Add(this.tasks...); err != nil {
		return
	}
	go state.Wait()
	return
}
