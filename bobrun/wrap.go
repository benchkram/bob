package bobrun

import (
	"fmt"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/errz"
)

type RunWrapper struct {
	ctl.Command
}

func (r *Run) WrapCommand(rc ctl.Command) ctl.Command {
	rw := RunWrapper{
		Command: rc,
	}
	return &rw
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
