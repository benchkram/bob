package bobtask

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/Benchkram/errz"
)

func (t *Task) Run(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	taskstr := fmt.Sprintf("%-*s", namePad, t.ColoredName())

	return t.runCmds(ctx, taskstr, t.cmds)
}

func (t *Task) PreRun(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	if len(t.precmds) > 0 {
		// boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t  %s", namePad, t.ColoredName(), aurora.Faint("Pre run Command are running")))
		taskstr := fmt.Sprintf("%-*s", namePad, t.ColoredNameWithSuffix(" (pre)"))
		err = t.runCmds(ctx, taskstr, t.precmds)
		if err != nil {
			return usererror.Wrapm(err, "Failed while running the Pre-run commands")
		}
	}

	return nil
}

func (t *Task) PostRun(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	if len(t.postcmds) > 0 {
		taskstr := fmt.Sprintf("%-*s", namePad, t.ColoredNameWithSuffix(" (post)"))
		err = t.runCmds(ctx, taskstr, t.postcmds)
		if err != nil {
			return usererror.Wrapm(err, "Failed while running the Post-run commands")
		}
	}

	return nil
}

// runCmds, runs all the commands from the cmdilst prints the output on different go routines
// under single waitgroup. wait until all the task under wait group finished before return.
func (t *Task) runCmds(ctx context.Context, taskstr string, cmdlist []string) error {
	var wg sync.WaitGroup

	for _, run := range cmdlist {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		if err != nil {
			return usererror.Wrapm(err, "shell command parse error")
		}

		env := os.Environ()
		// TODO: warn when overwriting envvar from the environment
		env = append(env, t.env...)

		pr, pw, err := os.Pipe()
		if err != nil {
			return err
		}

		s := bufio.NewScanner(pr)
		s.Split(bufio.ScanLines)

		wg.Add(1)

		go func() {
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				boblog.Log.V(1).Info(fmt.Sprintf("%s\t  %s", taskstr, aurora.Faint(s.Text())))
			}
			wg.Done()
		}()

		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(t.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, pw, pw),
		)
		errz.Fatal(err)

		err = r.Run(ctx, p)
		if err != nil {
			return usererror.Wrapm(err, "shell commands execution error")
		}

		pw.Close()
	}

	wg.Wait()
	return nil
}
