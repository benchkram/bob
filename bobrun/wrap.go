package bobrun

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

type RunWrapper struct {
	ctl.Command
	run *Run
	ctx context.Context
}

func (r *Run) WrapCommand(ctx context.Context, rc ctl.Command) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	// name := fmt.Sprintf("%s_%s", r.name, "init")

	// var initCommand ctl.Command
	// if len(r.init) != 0 {
	// 	initCommand, err = execctl.NewCmd(name, "bash", r.init...)
	// 	errz.Fatal(err)
	// }

	rw := RunWrapper{
		Command: rc,
		run:     r,
		ctx:     ctx,
	}
	return &rw, nil
}

// func (rw *RunWrapper) Start() (err error) {
// 	defer errz.Recover(&err)
// 	boblog.Log.Info(fmt.Sprintf("Start [%s] Pre", rw.Name()))
// 	// TODO: run rw.Pre
// 	err = rw.Command.Start()
// 	errz.Fatal(err)

// 	boblog.Log.Info(fmt.Sprintf("Started [%s] running: %+v", rw.Name(), rw.Running()))

// 	boblog.Log.Info(fmt.Sprintf("Start [%s] Post", rw.Name()))

// 	return nil
// }

// func (rw *RunWrapper) Stop() (err error) {
// 	defer errz.Recover(&err)
// 	err = rw.Command.Stop()
// 	errz.Fatal(err)

// 	boblog.Log.Info(fmt.Sprintf("Stop [%s] Post", rw.Name()))
// 	return nil
// }

// func (rw *RunWrapper) Restart() error {
// 	boblog.Log.Info(fmt.Sprintf("Restart [%s] Pre", rw.Name()))
// 	return rw.Command.Restart()
// }

func (rw *RunWrapper) Init() (err error) {
	defer errz.Recover(&err)

	boblog.Log.V(1).Info("Init called")

	// no init command
	if len(rw.run.init) == 0 {
		return nil
	}

	// Wait for initial command to have started
	// go func() {
	// for !rw.Running() {
	// 	time.Sleep(100 * time.Millisecond)
	// }

	// }()

	boblog.Log.Info(fmt.Sprintf("Init [%s] ", rw.Name()))

	err = rw.startInit()
	errz.Fatal(err)

	// At last call Init on the command itself
	return rw.Command.Init()
}

func (rw *RunWrapper) startInit() (err error) {

	boblog.Log.Info("startInit")

	// namePad := fmt.Sprintf("%s_init", rw.Name())

	for _, run := range rw.run.init {
		// break
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		if err != nil {
			return usererror.Wrapm(err, "shell command parse error")
		}

		env := os.Environ()
		// TODO: warn when overwriting envvar from the environment
		// env = append(env, rw.env...)

		pr, pw, err := os.Pipe()
		if err != nil {
			return err
		}

		s := bufio.NewScanner(pr)
		s.Split(bufio.ScanLines)

		go func() {
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				boblog.Log.V(1).Info(fmt.Sprintf("\t%s", aurora.Faint(s.Text())))
			}
		}()

		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(rw.run.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, pw, pw),
		)
		errz.Fatal(err)

		err = r.Run(rw.ctx, p)
		if err != nil {
			return usererror.Wrapm(err, "shell command execute error")
		}
	}

	return nil
}

func (rw *RunWrapper) stopInit() error {

	return nil
}
