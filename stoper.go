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
		if s != nil && s.IsRunning() {
			s.Stop()
		}
	}
}

func (s *Stopers) IsRunningOrRemove() (ok bool) {
	var new Stopers
	for _, s := range *s {
		if !s.IsRunning() {
			continue
		}
		new = append(new, s)
	}
	*s = new
	return len(new) > 0
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
