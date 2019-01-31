package task

import (
	"sync"
	"time"
)

type State struct {
	tasks   Slice
	stopers Stopers
	Start   time.Time
	End     time.Time
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
		ts    Slice
		stop  Stopers
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

	for _, t = range items {
		if s, err = t.Start(wg.Done); err != nil {
			for _, s = range stop {
				s.Stop()
			}
			return
		} else if s == nil {
			wg.Done()
		} else {
			ts = append(ts, t)
			stop = append(stop, s)
		}
	}

	p = &State{
		tasks:   items,
		stopers: stop,
		Start:   now,
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

func (t State) Stop() {
	t.stopers.Stop()
}

func (t State) IsRunning() bool {
	return t.stopers.IsRunning()
}

func (t State) Tasks() Slice {
	return t.tasks
}

func (t State) Stopers() Stopers {
	return t.stopers
}
