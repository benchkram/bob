package bobrun

import (
	"context"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/composectl"
	"github.com/benchkram/bob/pkg/composeutil"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"
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

	cfgs := composeutil.ProjectPortConfigs(p)

	portConflicts := ""
	portMapping := ""
	if composeutil.HasPortConflicts(cfgs) {
		conflicts := composeutil.PortConflicts(cfgs)

		portConflicts = conflicts.String()

		// TODO: disable once we also resolve binaries' ports
		errz.Fatal(usererror.Wrap(fmt.Errorf(fmt.Sprint("conflicting ports detected:\n", conflicts))))

		resolved, err := composeutil.ResolvePortConflicts(conflicts)
		errz.Fatal(err)

		portMapping = resolved.String()

		// update project's ports
		composeutil.ApplyPortMapping(p, resolved)
	}

	ctler, err := composectl.New(p, portConflicts, portMapping)
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
