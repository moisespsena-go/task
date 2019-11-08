package task

import "sync"

type startableTask struct {
	start       func(done ...func()) (err error)
	stop        func()
	RunningFunc func() bool
	running     bool
	c           chan struct{}
	mu          sync.Mutex
}

func (this *startableTask) Stop() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.stop()
	if this.RunningFunc == nil {
		this.running = false
	}
	close(this.c)
}

func (this *startableTask) IsRunning() bool {
	if this.RunningFunc != nil {
		return this.RunningFunc()
	}
	return this.running
}

func (this *startableTask) Setup(appender Appender) error {
	return nil
}

func (this *startableTask) Run() (err error) {
	if _, err = this.Start(nil); err != nil {
		return
	}
	<-this.c
	return
}

func (this *startableTask) Start(done func()) (stop Stoper, err error) {
	if err = this.start(func() {
		defer func() {
			this.mu.Lock()
			defer this.mu.Unlock()
			if this.running {
				close(this.c)
				this.running = false
			}
		}()
		if done != nil {
			done()
		}
	}); err != nil {
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()
	this.c = make(chan struct{})
	this.running = true
	return this, nil
}

func NewStartableTask(start func(done ...func()) (err error), stop func()) *startableTask {
	return &startableTask{
		start: start,
		stop:  stop,
	}
}
