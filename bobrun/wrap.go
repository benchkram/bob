package bobrun

import (
	"fmt"
	"time"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/execctl"
	"github.com/benchkram/errz"
)

type RunWrapper struct {
	ctl.Command
	initCommand ctl.Command
}

func (r *Run) WrapCommand(rc ctl.Command) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	name := fmt.Sprintf("%s_%s", r.name, "init")

	var initCommand ctl.Command
	if len(r.init) != 0 {
		initCommand, err = execctl.NewCmd(name, "bash", r.init...)
		errz.Fatal(err)
	}

	rw := RunWrapper{
		Command:     rc,
		initCommand: initCommand,
	}
	return &rw, nil
}

func (rw *RunWrapper) Start() (err error) {
	defer errz.Recover(&err)
	boblog.Log.Info(fmt.Sprintf("Start [%s] Pre", rw.Name()))
	// TODO: run rw.Pre
	err = rw.Command.Start()
	errz.Fatal(err)

	boblog.Log.Info(fmt.Sprintf("Started [%s] running: %+v", rw.Name(), rw.Running()))

	boblog.Log.Info(fmt.Sprintf("Start [%s] Post", rw.Name()))

	return nil
}

func (rw *RunWrapper) Stop() (err error) {
	defer errz.Recover(&err)
	err = rw.Command.Stop()
	errz.Fatal(err)

	boblog.Log.Info(fmt.Sprintf("Stop [%s] Post", rw.Name()))
	return nil
}

func (rw *RunWrapper) Restart() error {
	boblog.Log.Info(fmt.Sprintf("Restart [%s] Pre", rw.Name()))
	return rw.Command.Restart()
}

func (rw *RunWrapper) Init() (err error) {
	defer errz.Recover(&err)

	if rw.initCommand == nil {
		return nil
	}

	// Wait for initial command to have started
	go func() {
		for !rw.Running() {
			time.Sleep(100 * time.Millisecond)
		}

	}()

	boblog.Log.Info(fmt.Sprintf("Init [%s] ", rw.Name()))

	err = rw.initCommand.Start()
	errz.Fatal(err)

	// At last call Init on the command itself
	return rw.Command.Init()
}
