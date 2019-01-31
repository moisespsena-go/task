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

func (f *TaskFactory) Setup(appender Appender) error {
	return f.getTask().Setup(appender)
}

func (f *TaskFactory) Run() error {
	return f.getTask().Run()
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
