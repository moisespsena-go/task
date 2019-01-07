package restarts

import (
	"os"

	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"
	"github.com/moisespsena-go/task"
)

type SelfRestarter struct {
	*task.Runner
	SelfUpdater fetcher.Interface
}

func (r *SelfRestarter) Run() {
	overseer.Run(overseer.Config{
		Fetcher: r.SelfUpdater,
		Program: func(state overseer.State) {
			if done, err := r.Runner.Run(); err != nil {
				panic(err)
				os.Exit(1)
			} else {
				<-done
			}
		},
	})
}

func New(t ...task.Task) *task.Runner {
	return task.NewRunner(t...).SignalStop()
}

func RunConfig(r *task.Runner, cfg *overseer.Config) {
	if cfg == nil {
		cfg = &overseer.Config{}
	}
	cfg.Program = func(state overseer.State) {
		if done, err := r.Run(); err != nil {
			os.Stderr.Write([]byte(err.Error()))
			os.Exit(1)
		} else {
			<-done
		}
	}
	run(cfg)
}

func RunUpdater(r *task.Runner, fetcher fetcher.Interface) {
	RunConfig(r, &overseer.Config{Fetcher: fetcher})
}

func Run(r *task.Runner) {
	RunConfig(r, nil)
}
