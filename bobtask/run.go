package bobtask

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/benchkram/errz"
)

func (t *Task) Run(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	for _, run := range t.cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		if err != nil {
			return usererror.Wrapm(err, "shell command parse error")
		}

		err = updatePath(ctx)
		if err != nil {
			return err
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

		go func() {
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t  %s", namePad, t.ColoredName(), aurora.Faint(s.Text())))
			}
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
	}

	return nil
}

func updatePath(ctx context.Context) error {
	if ctx.Value("newPath") == nil {
		return nil
	}

	newPath := ctx.Value("newPath").(string)
	fmt.Printf("Updating $PATH to: %s\n", newPath)

	return os.Setenv("PATH", newPath)
}
