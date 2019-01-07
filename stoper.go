package task

type Stopers []Stoper

func (s Stopers) IsRunning() bool {
	for _, s := range s {
		if s.IsRunning() {
			return true
		}
	}
	return false
}

func (s Stopers) Stop() {
	for _, s := range s {
		if s.IsRunning() {
			s.Stop()
		}
	}
}

type Stoper interface {
	Stop()
	IsRunning() bool
}

type stoper struct {
	stop    func()
	running func() bool
	done    bool
}

func (s *stoper) Stop() {
	s.stop()
}

func (s *stoper) IsRunning() bool {
	return s.running()
}

func NewStoper(stop func(), running func() bool) (s Stoper) {
	return &stoper{stop: stop, running: running}
}

func NewStopDoner(stop func()) (s Stoper, done func()) {
	st := &stoper{stop: stop}
	st.running = func() bool {
		return !st.done
	}
	return st, func() {
		st.done = true
	}
}

type FakeStoper struct{}

func (FakeStoper) Stop() {}

func (FakeStoper) IsRunning() bool { return false }
