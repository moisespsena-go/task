// +build !windows

package task

import "syscall"

func init() {
	runnner_signals = append(runnner_signals, syscall.SIGUSR2)
}
