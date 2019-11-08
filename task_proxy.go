package task

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
