package task

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/moisespsena/go-error-wrap"
	"github.com/op/go-logging"
)

type Runner struct {
	*State
	tasks Slice
	log   *logging.Logger
	*OnDoneEvent
}

func NewRunner(t ...Task) *Runner {
	return &Runner{tasks: t, OnDoneEvent: &OnDoneEvent{}}
}

func (r *Runner) SetLog(log *logging.Logger) *Runner {
	r.log = log
	return r
}

func (r *Runner) Run() (done chan interface{}, err error) {
	done = make(chan interface{})
	var stop Stoper

	if stop, err = Start(func(s *State) {
		defer func() {
			close(done)
		}()
		layout := "2006-01-02 15:04:05 Z07:00"
		msg := fmt.Sprintf("Done: from `%v` to `%v` with %v duration.", s.Start.Format(layout), s.End.Format(layout), s.End.Sub(s.Start))
		if r.log == nil {
			fmt.Println(msg)
		} else {
			r.log.Notice(msg)
		}
	}, r.tasks...); err != nil {
		return nil, errwrap.Wrap(err, "task start")
	}

	r.State = stop.(*State)
	r.State.OnDone(r.onDone...)
	r.OnDoneEvent = &r.State.OnDoneEvent
	return
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

func (r *Runner) SigStop(sig ...os.Signal) *Runner {
	if len(sig) == 0 {
		sig = append(sig, syscall.SIGINT, syscall.SIGTERM)
		sig = append(sig, runnner_signals...)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig...)
	go func() {
		sig := <-sigs
		if r.log == nil {
			fmt.Println("received signal:", sig)
		} else {
			r.log.Notice("received signal: %v", sig)
		}
		if r.State != nil {
			r.Stop()
		}
	}()
	return r
}

func (r *Runner) SigRun(sig ...os.Signal) (done chan interface{}, err error) {
	if done, err = r.Run(); err != nil {
		return nil, err
	} else {
		r.SigStop(sig...)
	}
	return
}

func (r *Runner) SigRunWait(sig ...os.Signal) (err error) {
	if done, err := r.SigRun(); err != nil {
		return err
	} else {
		<-done
	}
	return
}

func (r *Runner) MustSigRun(sig ...os.Signal) {
	if err := r.SigRunWait(sig...); err != nil {
		log.Fatal(err)
	}
}
