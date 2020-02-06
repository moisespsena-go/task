package task

type Factory interface {
	Task
	Factory() Task
}

type TaskFactory struct {
	FactoryFunc func() Task
	task        Task
}

func (f *TaskFactory) getTask() Task {
	if f.task == nil {
		f.task = f.Factory()
	}
	return f.task
}

func (f *TaskFactory) Setup(appender Appender) (err error) {
	return setup(appender, func(t ...Task) error {
		return nil
	}, f.getTask())
}

func (f *TaskFactory) Run() (err error) {
	t := f.getTask()
	if runer, ok := t.(TaskRunner); ok {
		return runer.Run()
	}
	done := make(chan struct{})
	if _, err = t.Start(func() {
		close(done)
	}); err != nil {
		return
	}
	<-done
	return
}

func (f *TaskFactory) Start(done func()) (stop Stoper, err error) {
	return f.getTask().Start(done)
}

func (f *TaskFactory) Factory() Task {
	return f.FactoryFunc()
}

type FactoryFunc func() Task

func (f FactoryFunc) Factory() Task {
	return f()
}

func (f FactoryFunc) Setup(appender Appender) (err error) {
	return
}

func (f FactoryFunc) Run() error {
	return nil
}

func (f FactoryFunc) Start(done func()) (stop Stoper, err error) {
	return
}
