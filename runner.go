package task

import (
	"os"

	errwrap "github.com/moisespsena-go/error-wrap"
	"github.com/moisespsena-go/signald"

	"github.com/moisespsena-go/logging"
)

type Runner struct {
	*State
	tasks Slice
	log   logging.Logger
	*OnDoneEvent
	Signal os.Signal
}

func NewRunner(t ...Task) *Runner {
	return &Runner{
		tasks:       t,
		log:         logging.WithPrefix(log, "Runner"),
		OnDoneEvent: &OnDoneEvent{},
	}
}

func (r *Runner) SetLog(log logging.Logger) *Runner {
	r.log = log
	return r
}

func (this *Runner) GetLogger() logging.Logger {
	return this.log
}

func (r *Runner) Run() (done chan interface{}, err error) {
	done = make(chan interface{})
	var stop Stoper

	if stop, err = Start(func(s *State) {
		defer func() {
			close(done)
		}()
		layout := "2006-01-02 15:04:05 Z07:00"
		r.log.Noticef("Done: from `%v` to `%v` with %v duration.",
			s.Start.Format(layout), s.End.Format(layout), s.End.Sub(s.Start))
	}, r.tasks...); err != nil {
		return nil, errwrap.Wrap(err, "task start")
	}
	r.State = stop.(*State)
	r.State.OnDone(r.onDone...)
	r.OnDoneEvent = &r.State.OnDoneEvent
	return
}

func (r *Runner) Add(t ...Task) (err error) {
	var prepared *PreparedTasks
	if prepared, err = Prepare(t...); err != nil {
		return
	}
	return r.State.Add(prepared.tasks...)
}

func (r *Runner) RunWait() (err error) {
	if done, err := r.Run(); err != nil {
		return err
	} else {
		<-done
	}
	return
}

func (r *Runner) MustRunWait() {
	if err := r.RunWait(); err != nil {
		log.Fatal(err)
	}
}

func (this *Runner) SignalBinder() signald.Binder {
	return signald.Binder{
		Callback: func(signal os.Signal) {
			this.StopWait()
		},
		Unbind: func(f func()) {
			this.OnDone(f)
		},
	}
}
