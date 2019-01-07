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
	tasks  []Task
	stoper Stoper
	log    *logging.Logger
}

func NewRunner(t ...Task) *Runner {
	return &Runner{tasks: t}
}

func (r *Runner) Stop() {
	r.stoper.Stop()
}

func (r *Runner) IsRunning() bool {
	return r.stoper.IsRunning()
}

func (r *Runner) SetLog(log *logging.Logger) *Runner {
	r.log = log
	return r
}

func (r *Runner) Run() (done chan bool, err error) {
	done = make(chan bool, 1)

	if r.stoper, err = Start(func() { done <- true }, r.tasks...); err != nil {
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
			r.log.Debug("received signal:", sig)
		}
		if r.stoper != nil {
			r.stoper.Stop()
		}
	}()
	return r
}
