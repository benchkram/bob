package bobtask

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/benchkram/errz"
)

func (t *Task) Run(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	env := t.Env()
	if len(t.storePaths) > 0 {
		env = nix.AddPATH(t.storePaths, env)
	}

	for _, run := range t.cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		if err != nil {
			return usererror.Wrapm(err, "shell command parse error")
		}

		pr, pw, err := os.Pipe()
		if err != nil {
			return err
		}

		s := bufio.NewScanner(pr)
		s.Split(bufio.ScanLines)

		done := make(chan bool)

		go func() {
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t  %s", namePad, t.ColoredName(), aurora.Faint(s.Text())))
			}

			done <- true
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
			return usererror.Wrapm(err, "shell command execute error")
		}

		// wait for the reader to finish after closing the write pipe
		pw.Close()
		<-done
	}

	return nil
}
