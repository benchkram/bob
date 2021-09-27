package bobtask

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/Benchkram/errz"
)

func (t *Task) Run(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	fmt.Printf("Running %s\n", t.name)
	for _, run := range t.cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		errz.Fatal(err)

		env := os.Environ()
		// TODO: warn when overwriting envvar from the environment
		env = append(env, t.env...)

		//litter.Dump(t.env)
		// fmt.Printf("exec %s in dir:%s\n", run, t.dir)
		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(t.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, os.Stdout, os.Stderr),
		)
		errz.Fatal(err)

		err = r.Run(ctx, p)
		errz.Fatal(err)

		// fmt.Printf("%s succeded \n", t.name)
	}

	return nil
}
