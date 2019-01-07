// +build !github_com__moisespsena__task_restarter

package restarts

import "github.com/jpillora/overseer"

func run(cfg *overseer.Config) {
	cfg.Program(overseer.State{})
}
