package bobrun

import (
	"context"

	"github.com/Benchkram/bob/pkg/composectl"
	"github.com/Benchkram/bob/pkg/composeutil"
	"github.com/Benchkram/bob/pkg/ctl"
	"github.com/Benchkram/errz"
)

const composeFileDefault = "docker-compose.yml"

func (r *Run) composeCommand(ctx context.Context) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	path := r.Path
	if path == "" {
		path = composeFileDefault
	}

	p, err := composeutil.ProjectFromConfig(path)
	errz.Fatal(err)

	configs := composeutil.PortConfigs(p)

	hasPortConflict := composeutil.HasPortConflicts(configs)

	mappings := ""
	conflicts := ""
	if hasPortConflict {
		conflicts = composeutil.GetPortConflicts(configs)

		resolved, err := composeutil.ResolvePortConflicts(p, configs)
		if err != nil {
			errz.Fatal(err)
		}

		mappings = composeutil.GetNewPortMappings(resolved)
	}

	ctler, err := composectl.New(p, conflicts, mappings)
	errz.Fatal(err)

	rc := ctl.New(r.name, 1, ctler.Stdout(), ctler.Stderr(), ctler.Stdin())

	go func() {
		for {
			switch <-rc.Control() {
			case ctl.Start:
				err = ctler.Up(ctx)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStarted()
				}
			case ctl.Stop:
				err = ctler.Down(ctx)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStopped()
				}
			case ctl.Shutdown:
				// SIGINT takes an extra context to allow
				// a cleanup.
				_ = ctler.Down(ctx)
				// TODO: log error to a logger ot emit
				rc.EmitDone()
				return
			}
		}
	}()

	return rc, nil
}
