package task

import (
	"sync"
	"time"
)

type taskStoper struct {
	Task
	Stoper
	stopFunc func()
}

func (ts taskStoper) Stop() {
	defer ts.stopFunc()
	ts.Stoper.Stop()
}

type State struct {
	Start       time.Time
	End         time.Time
	mu          sync.Mutex
	taskStopers map[interface{}]taskStoper
}

func (s *State) AddTaskStoper(key interface{}, t Task, stop Stoper) (newStoper Stoper) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskStopers == nil {
		s.taskStopers = map[interface{}]taskStoper{}
	}
	s.taskStopers[key] = taskStoper{t, s, func() {
		delete(s.taskStopers, key)
	}}

	return s.taskStopers[key]
}

func (s State) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.taskStopers == nil {
		return
	}
	for _, t := range s.taskStopers {
		t.Stop()
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

func (pt *PreparedTasks) Tasks() Slice {
	return pt.tasks
}

func Prepare(t ...Task) (pt *PreparedTasks, err error) {
	appender := appenderSetup{NewAppender()}
	if err = appender.AddTask(t...); err != nil {
		return
	}
	pt = &PreparedTasks{appender.Tasks()}
	return
}

func (pt *PreparedTasks) Start(doneFuncs ...func()) (p *State, err error) {
	if len(pt.tasks) == 0 {
		return
	}

	doneFuncs = append([]func(){}, doneFuncs...)

	var (
		wg    sync.WaitGroup
		ta    = &TaskAppender{}
		s     Stoper
		t     Task
		items Slice
		now   = time.Now()

		done = func() {
			for _, d := range doneFuncs {
				if d != nil {
					d()
				}
			}
		}

		add func(tasks ...Task) (err error)
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
				items = append(items, t)
			}
		}
		return
	}

	if err = add(pt.tasks...); err != nil {
		return
	}

	for _, t = range items {
		if pr, ok := t.(PreRunCallback); ok {
			if err = pr.TaskPreRun(ta); err != nil {
				return
			}
		}
		if pr, ok := t.(PostRunCallback); ok {
			doneFuncs = append(doneFuncs, pr.TaskPosRun)
		}
	}

	items = append(items, ta.tasks...)
	wg.Add(len(items))

	p = &State{
		Start: now,
	}

	for i, t := range items {
		if s, err = t.Start(wg.Done); err != nil {
			p.Stop()
			return
		} else if s == nil {
			wg.Done()
		} else {
			p.AddTaskStoper(i, t, s)
		}
	}

	if len(items) == 0 {
		p.End = time.Now()
		go done()
		return
	}
	go func() {
		defer func() {
			p.End = time.Now()
			done()
		}()
		wg.Wait()
	}()
	return
}
