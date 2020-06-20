package task

import (
	"os"

	"github.com/moisespsena-go/signald"
)

func (r *Runner) SigStop() *Runner {
	signald.AutoBindInterface(r)
	return r
}

func (r *Runner) SigRun(sig ...os.Signal) (done chan interface{}, err error) {
	if done, err = r.Run(); err != nil {
		return nil, err
	} else {
		r.SigStop()
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
