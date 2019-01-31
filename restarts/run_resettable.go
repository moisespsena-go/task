// +build resettable

package restarts

import "github.com/jpillora/overseer"

func run(cfg *overseer.Config) {
	overseer.Run(*cfg)
}
