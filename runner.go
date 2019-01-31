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
	Stoper
	tasks Slice
	log   *logging.Logger
}

func NewRunner(t ...Task) *Runner {
	return &Runner{tasks: t}
}

func (r *Runner) SetLog(log *logging.Logger) *Runner {
	r.log = log
	return r
}

func (r *Runner) Run() (done chan bool, err error) {
	done = make(chan bool, 1)

	if r.Stoper, err = Start(func(s *State) {
		defer func() {
			done <- true
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
	return
}

func (r *Runner) SignalStop(sig ...os.Signal) *Runner {
	if len(sig) == 0 {
		sig = append(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, sig...)
	go func() {
		sig := <-sigs
		if r.log == nil {
			fmt.Println("received signal:", sig)
		} else {
			r.log.Notice("received signal:", sig)
		}
		if r.Stoper != nil {
			r.Stoper.Stop()
		}
	}()
	return r
}
