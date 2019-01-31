// +build !resettable

package restarts

import "github.com/jpillora/overseer"

func run(cfg *overseer.Config) {
	cfg.Program(overseer.State{})
}
