package bobrun

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/composectl"
	"github.com/benchkram/bob/pkg/composeutil"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

const composeFileDefault = "docker-compose.yml"

func (r *Run) composeCommand(ctx context.Context) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	fmt.Println("compose cmd")

	path := r.Path
	if path == "" {
		path = composeFileDefault
	}

	p, err := composeutil.ProjectFromConfig(path)
	errz.Fatal(err)

	//// In case the environment was already running (because of crash during shutdown, for example), shut it down.
	//ctler, err := composectl.New(p, "", "")
	//errz.Fatal(err)
	//
	//err = ctler.Down(ctx)
	//errz.Fatal(err)

	//cfgs := composeutil.ProjectPortConfigs(p)

	//portConflicts := ""
	//portMapping := ""
	//var resolved composeutil.PortConfigs
	//
	//if composeutil.HasPortConflicts(cfgs) {
	//	conflicts := composeutil.PortConflicts(cfgs)
	//
	//	portConflicts = conflicts.String()
	//
	//	// TODO: disable once we also resolve binaries' ports
	//	//errz.Fatal(usererror.Wrap(fmt.Errorf(fmt.Sprint("conflicting ports detected:\n", conflicts))))
	//
	//	resolved, err = composeutil.ResolvePortConflicts(conflicts)
	//	errz.Fatal(err)
	//
	//	portMapping = resolved.String()
	//
	//	// update project's ports
	//}

	//ctler, err := composectl.New()
	//errz.Fatal(err)
	//
	//err = ctler.Down(ctx, p)
	//errz.Fatal(err)
	//
	//portConflicts = ""
	//portMapping = ""
	//if composeutil.HasPortConflicts(cfgs) {
	//	conflicts := composeutil.PortConflicts(cfgs)
	//
	//	portConflicts = conflicts.String()
	//
	//	// TODO: disable once we also resolve binaries' ports
	//	errz.Fatal(usererror.Wrap(fmt.Errorf(fmt.Sprint("conflicting ports detected:\n", conflicts))))
	//
	//	resolved, err := composeutil.ResolvePortConflicts(conflicts)
	//	errz.Fatal(err)
	//
	//	portMapping = resolved.String()
	//
	//	// update project's ports
	//	composeutil.ApplyPortMapping(p, resolved)
	//}
	//
	//w := ctler.StdoutWriter()
	//
	//if portConflicts != "" {
	//	portConflicts = fmt.Sprintf("%s\n%s\n", "Conflicting ports detected:", portConflicts)
	//	_, err = w.Write([]byte(portConflicts))
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//if portMapping != "" {
	//	portMapping = fmt.Sprintf("%s\n%s\n", "Resolved port mapping:", portMapping)
	//	_, err = w.Write([]byte(portMapping))
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//rc := ctl.New(r.name, 1, ctler.Stdout(), ctler.Stderr(), ctler.Stdin())
	//
	//go func() {
	//}()

	ctler, err := composectl.New()
	errz.Fatal(err)

	w := ctler.StdoutWriter()

	ready := make(chan bool)

	go func() {
		err = ctler.Down(ctx, p)
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

		if portConflicts != "" {
			portConflicts = fmt.Sprintf("%s\n%s\n", "Conflicting ports detected:", portConflicts)
			_, err = w.Write([]byte(portConflicts))
			if err != nil {
				errz.Fatal(err)
			}
		}

		if portMapping != "" {
			portMapping = fmt.Sprintf("%s\n%s\n", "Resolved port mapping:", portMapping)
			_, err = w.Write([]byte(portMapping))
			if err != nil {
				errz.Fatal(err)
			}
		}

		ready <- true
	}()

	//cfgs := composeutil.ProjectPortConfigs(p)

	//if composeutil.HasPortConflicts(cfgs) {
	//	conflicts := composeutil.PortConflicts(cfgs)
	//
	//	//portConflicts = conflicts.String()
	//
	//	// TODO: disable once we also resolve binaries' ports
	//	//errz.Fatal(usererror.Wrap(fmt.Errorf(fmt.Sprint("conflicting ports detected:\n", conflicts))))
	//
	//	resolved, err := composeutil.ResolvePortConflicts(conflicts)
	//	errz.Fatal(err)
	//
	//	//portMapping = resolved.String()
	//
	//	// update project's ports
	//	composeutil.ApplyPortMapping(p, resolved)
	//}

	//w := ctler.StdoutWriter()

	rc := ctl.New(r.name, 1, ctler.Stdout(), ctler.Stderr(), ctler.Stdin())

	waitForReady := true

	go func() {
		for {
			switch <-rc.Control() {
			case ctl.Start:
				if waitForReady {
					<-ready
					waitForReady = false
				}

				err = ctler.Up(ctx, p)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStarted()
				}
			case ctl.Stop:
				err = ctler.Down(ctx, p)
				if err != nil {
					rc.EmitError(err)
				} else {
					rc.EmitStopped()
				}
			case ctl.Shutdown:
				// SIGINT takes an extra context to allow
				// a cleanup.
				_ = ctler.Down(ctx, p)
				// TODO: log error to a logger ot emit
				rc.EmitDone()
				return
			}
		}
	}()

	return rc, nil
}
