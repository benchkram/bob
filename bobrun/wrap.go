package bobrun

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

// RunWrapper provides init functionality to be executed after
// a command has started.
type RunWrapper struct {
	// inner is the wrapped command
	inner ctl.Command

	// run is the run cmd to get init instructions from.
	run *Run

	// ctx is the outer context to listen to
	ctx context.Context

	// mux a mutex to protect the initRunning flag,
	mux sync.Mutex
	// initRunning set to true when the a init function is running.
	initRunning bool
	// initDone is closed when init is finished.
	initDone chan struct{}
	// initCtxCancel cancels the init cmd.
	// It is recreated each before each init() run.
	initCtxCancel context.CancelFunc

	// done indicates the cmd has shutdown
	done chan struct{}

	stdout pipe
	stderr pipe
}

type pipe struct {
	r *os.File
	w *os.File
}

func (r *Run) WrapCommand(ctx context.Context, rc ctl.Command) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	rw := &RunWrapper{
		inner: rc,
		run:   r,
		ctx:   ctx,

		done: make(chan struct{}),
	}

	// create pipes for stdout, stderr and stdin
	rw.stdout.r, rw.stdout.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	rw.stderr.r, rw.stderr.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	// react to done from inner control
	go func() {
		<-rw.inner.Done()
		<-rw.initDone
		close(rw.done)
	}()

	return rw, nil
}

func (rw *RunWrapper) Name() string {
	return rw.inner.Name()
}

func (rw *RunWrapper) Restart() (err error) {
	return rw.inner.Restart()
}

func (rw *RunWrapper) Start() (err error) {
	defer errz.Recover(&err)

	err = rw.inner.Start()
	errz.Fatal(err)

	// Wait for initial command to have started
	for !rw.inner.Running() {
		time.Sleep(100 * time.Millisecond)
	}

	boblog.Log.Info(fmt.Sprintf("Init [%s] ", rw.inner.Name()))

	return rw.init()
}

func (rw *RunWrapper) Stop() error {
	rw.mux.Lock()
	if rw.initRunning {
		rw.initCtxCancel()
	}
	rw.mux.Unlock()

	// wait for init to shutdown or deadline is reached.
	select {
	case <-rw.initDone:
	case <-time.After(5 * time.Second):
	}

	return rw.inner.Stop()
}

func (rw *RunWrapper) Running() bool {
	return rw.inner.Running()
}

func (rw *RunWrapper) Shutdown() error {
	rw.mux.Lock()
	if rw.initRunning {
		rw.initCtxCancel()
	}
	rw.mux.Unlock()

	return rw.inner.Shutdown()
}

func (rw *RunWrapper) Done() <-chan struct{} {
	return rw.done
}

func (rw *RunWrapper) Stdout() io.Reader {
	return io.MultiReader(rw.inner.Stdout(), rw.stdout.r)
}
func (rw *RunWrapper) Stderr() io.Reader {
	return io.MultiReader(rw.inner.Stderr(), rw.stderr.r)
}
func (rw *RunWrapper) Stdin() io.Writer {
	return rw.inner.Stdin()
}

func (rw *RunWrapper) init() (err error) {
	rw.mux.Lock()
	defer rw.mux.Unlock()

	if rw.initRunning {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	rw.initCtxCancel = cancel
	rw.initRunning = true
	rw.initDone = make(chan struct{})

	// React to cancel from external context
	go func() {
		select {
		case <-rw.ctx.Done():
		case <-rw.initDone: // assures this goroutine exits correctly
		}
		rw.initCtxCancel()
	}()

	go func() {
		for _, run := range rw.run.init {
			p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
			if err != nil {
				boblog.Log.V(1).Error(err, "shell command parse error")
				break
			}

			// FIXME: make run cmds ready for nix integration.
			env := os.Environ()

			pr, pw, err := os.Pipe()
			if err != nil {
				boblog.Log.V(1).Error(err, "creating a pipe")
				break
			}

			s := bufio.NewScanner(pr)
			s.Split(bufio.ScanLines)

			go func() {
				for s.Scan() {
					err := s.Err()
					if err != nil {
						return
					}

					// FIXME: why does printing to Commands stdout not work?
					fmt.Fprintf(rw.stdout.w, "\t%s\n", aurora.Faint(s.Text()))
					boblog.Log.V(1).Info(fmt.Sprintf("\t%s", aurora.Faint(s.Text())))
				}
			}()

			r, err := interp.New(
				interp.Params("-e"),
				interp.Dir(rw.run.dir),

				interp.Env(expand.ListEnviron(env...)),
				interp.StdIO(os.Stdin, pw, pw),
				// FIXME: why does this not work?
				//interp.StdIO(os.Stdin, rw.stdout.w, rw.stderr.w),
			)
			errz.Fatal(err)

			err = r.Run(ctx, p)
			if err != nil {
				boblog.Log.V(1).Error(err, "shell command execute error")
				break
			}
		}
		rw.mux.Lock()
		rw.initRunning = false
		rw.mux.Unlock()

		close(rw.initDone)
	}()

	return nil
}
