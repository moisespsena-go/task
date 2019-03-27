package task

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/moisespsena-go/default-logger"
	"github.com/op/go-logging"
)

type CmdTask struct {
	Cmd      *exec.Cmd
	Log      *logging.Logger
	PreStart func(t *CmdTask)
	onDone   []func()
}

func NewCmdTask(cmd *exec.Cmd) *CmdTask {
	return &CmdTask{Cmd: cmd}
}

func (t *CmdTask) OnDone(f ...func()) *CmdTask {
	t.onDone = append(t.onDone, f...)
	return t
}

func (t *CmdTask) SetLog(log *logging.Logger) *CmdTask {
	t.Log = log
	return t
}

func (t *CmdTask) Setup(appender Appender) error {
	return nil
}

func (t *CmdTask) LogInfo() {
	var args = []string{t.Cmd.Args[0]}
	for _, arg := range t.Cmd.Args[1:] {
		args = append(args, strconv.Quote(arg))
	}
	t.Log.Debug("Args: " + strings.Join(args, " "))
	if t.Cmd.Dir != "" {
		t.Log.Debug("Dir: " + t.Cmd.Dir)
	}
}

func (t *CmdTask) preStart() {
	if t.Log == nil {
		t.Log = defaultlogger.NewLogger("CMD: " + t.Cmd.Path)
	}
	if t.PreStart == nil {
		t.LogInfo()
	} else {
		t.PreStart(t)
	}
}

func (t *CmdTask) done() {
	if t.Cmd.ProcessState != nil {
		if s := t.Cmd.ProcessState; s.Success() {
			t.Log.Debug(s.String())
		} else {
			t.Log.Error(s.String())
		}
		for _, f := range t.onDone {
			f()
		}
	}
}
func (t *CmdTask) Run() (err error) {
	t.preStart()
	defer t.done()
	if err = t.Cmd.Start(); err != nil {
		return err
	}
	return t.Cmd.Wait()
}

func (t *CmdTask) Start(done func()) (stop Stoper, err error) {
	t.preStart()
	if err = t.Cmd.Start(); err != nil {
		return nil, err
	}

	var donev bool

	go func() {
		defer func() {
			donev = true
			t.done()
			done()
		}()
		if err := t.Cmd.Wait(); err != nil {
			if t.Log != nil {
				t.Log.Errorf("Wait failed: %v", err)
			} else {
				log.Error("Command %s with %s args wait failed: %v", t.Cmd.Path, t.Cmd.Args, err)
			}
		}
	}()

	return NewStoper(
		func() {
			t.Cmd.Process.Kill()
		},
		func() bool {
			return !donev
		},
	), nil
}
