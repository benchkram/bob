package bobtask

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/Benchkram/errz"
)

func (t *Task) Run(ctx context.Context, namePad int) (err error) {
	defer errz.Recover(&err)

	for _, run := range t.cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		errz.Fatal(err)

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

		//litter.Dump(t.env)
		// fmt.Printf("exec %s in dir:%s\n", run, t.dir)
		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(t.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, pw, pw),
		)
		errz.Fatal(err)

		err = r.Run(ctx, p)
		errz.Fatal(err)

		// fmt.Printf("%s succeded \n", t.name)
	}

	return nil
}
