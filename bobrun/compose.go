package bobrun

import (
	"context"

	"github.com/Benchkram/bob/pkg/composectl"
	"github.com/Benchkram/bob/pkg/composeutil"
	"github.com/Benchkram/bob/pkg/ctl"
	"github.com/Benchkram/bob/pkg/runctl"
	"github.com/Benchkram/errz"
)

const composeFileDefault = "docker-compose.yml"

func (r *Run) composeCommand(ctx context.Context) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	initialized := make(chan bool)
	rc := runctl.New(r.name, 1)

	go func() {

		path := r.Path
		if path == "" {
			path = composeFileDefault
		}
		project, err := composeutil.ProjectFromConfig(path)
		errz.Fatal(err)

		configs := composeutil.PortConfigs(project)

		hasPortConflict := composeutil.HasPortConflicts(configs)

		if hasPortConflict {
			composeutil.PrintPortConflicts(configs)

			resolved, err := composeutil.ResolvePortConflicts(project, configs)
			if err != nil {
				errz.Fatal(err)
			}

			composeutil.PrintNewPortMappings(resolved)
		}

		ctl, err := composectl.New()
		errz.Fatal(err)

		close(initialized)

		for {
			switch <-rc.Control() {
			case runctl.Start:
				err = ctl.Up(ctx, project)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStarted()
				}
			case runctl.Stop:
				err = ctl.Down(ctx)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStopped()
				}
			case runctl.Shutdown:
				// SIGINT takes an extra context to allow
				// a cleanup.
				err = ctl.Down(context.Background())
				errz.Log(err)
				rc.EmitDone()
				return
			}
		}
	}()

	<-initialized
	return rc, nil
}
