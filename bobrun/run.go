package bobrun

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/Benchkram/bob/pkg/composectl"
	"github.com/Benchkram/bob/pkg/composeutil"
	"github.com/Benchkram/bob/pkg/runctl"
	"github.com/Benchkram/errz"
)

var ErrInvalidRunType = fmt.Errorf("Invalid run type")

type Run struct {
	Type RunType

	// ComposePath is the path to a docker-compose file or binary
	// Default filename is used when empty.
	Path string

	// DependsOn run or build tasks
	DependsOn []string

	// didUpdate fires after the run-task
	// did a restart.
	didUpdate chan struct{}

	dir string

	name string
}

func (r *Run) Name() string {
	return r.name
}

func (r *Run) SetName(name string) {
	r.name = name
}

func (r *Run) Dir() string {
	return r.dir
}

func (r *Run) SetDir(dir string) {
	r.dir = dir
}

const composeFileDefault = "docker-compose.yml"

func New() *Run {
	r := &Run{
		Type:      RunTypeBinary,
		DependsOn: []string{},
		Path:      composeFileDefault,

		didUpdate: make(chan struct{}),
	}
	return r
}

// Run starts run cmds and return a channel to ctl the run. It also returns a
// `stoppped` channel which is closed when the run cmd finished it's work.
//
// To shutdown a Run() use a cancable context.
func (r *Run) Run(ctx context.Context) (rc runctl.Control, _ error) {
	fmt.Printf("Starting run task [%s]\n", r.name)

	switch r.Type {
	case RunTypeBinary:
		return r.runBinary(ctx)
	case RunTypeCompose:
		return r.runCompose(ctx)
	default:
		return nil, ErrInvalidRunType
	}
}

const (
	RestartSignal = syscall.Signal(0x70)
)

func (r *Run) runBinary(ctx context.Context) (_ runctl.Control, err error) {

	rc := runctl.New()

	go func() {
		defer func() {
			rc.EmitStop()
		}()

		innerCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cmd := exec.CommandContext(innerCtx, r.Path)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		done := make(chan struct{}, 1)

		err = cmd.Start()
		if err != nil {
			errz.Log(err)
			close(done)
		} else {
			go func() {
				rc.EmitStarted()
				err = cmd.Wait()
				errz.Log(err)
				close(done)
			}()
		}

		for {
			select {
			case <-done:
				fmt.Println("done with cmd")
				return
			case <-ctx.Done():
				fmt.Println("done with context")
				return
			case s := <-rc.Control():
				switch s {
				case RestartSignal:
					err = cmd.Process.Signal(syscall.SIGTERM)
					errz.Log(err)

				}
			}
		}
	}()

	return rc, nil
}

// func (r *Run) binary(ctx context.Context) (_ RunControl, err error) {

// 	rc := NewRunControl()

// 	innerCtx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	cmd := exec.CommandContext(innerCtx, r.Path)
// 	cmd.Stderr = os.Stderr
// 	cmd.Stdout = os.Stdout

// 	done := make(chan struct{}, 1)

// 	err = cmd.Start()
// 	if err != nil {
// 		errz.Log(err)
// 		close(done)
// 	} else {
// 		go func() {
// 			rc.EmitStarted()
// 			err = cmd.Wait()
// 			errz.Log(err)
// 			close(done)
// 		}()
// 	}
// }

func (r *Run) runCompose(ctx context.Context) (_ runctl.Control, err error) {
	defer errz.Recover(&err)

	rc := runctl.New()

	go func() {
		defer func() {
			rc.EmitStop()
		}()

		path := r.Path
		if path == "" {
			path = composeFileDefault
		}
		p, err := composeutil.ProjectFromConfig(path)
		errz.Fatal(err)

		configs := composeutil.PortConfigs(p)

		hasPortConflict := composeutil.HasPortConflicts(configs)

		if hasPortConflict {
			composeutil.PrintPortConflicts(configs)

			resolved, err := composeutil.ResolvePortConflicts(p, configs)
			if err != nil {
				errz.Fatal(err)
			}

			composeutil.PrintNewPortMappings(resolved)
		}

		innerCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ctl, err := composectl.New(innerCtx)
		errz.Fatal(err)

		fmt.Println()
		err = ctl.Up(p)
		errz.Fatal(err)

		defer func() {
			fmt.Print("\n\n")
			err := ctl.Down()
			errz.Log(err)
			fmt.Println("\nEnvironment down.")
		}()

		rc.EmitStarted()
		fmt.Print("\nEnvironment up.\n\n")

		// Blockign signal handling when doing
		// a restart. A shutdown signal is only handled
		// after a restart is done. Idealy we can shutdown
		// while restarting but it makes it pretty complicated.
		//
		// Bug: Handling restart in a goroutine might end up calling `up`
		// and `down` at the same time.
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-rc.Control():
				switch s {
				case RestartSignal:
					err = ctl.Down()
					errz.Log(err)
					err = ctl.Up(p)
					errz.Log(err)

					rc.EmitStarted()
				}
			}
		}
	}()

	return rc, nil
}
