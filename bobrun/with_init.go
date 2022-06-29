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

	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
)

// WithInit wraps a run-task to provide init functionality executed after
// the task started.
type WithInit struct {
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

	// once assures that the init is executed only once if defined so by the run-task.
	once sync.Once

	stdout pipe
	stderr pipe
}

type pipe struct {
	r *os.File
	w *os.File
}

// WrapWithInit takes a ctl to add init functionality defined in the run task.
func (r *Run) WrapWithInit(ctx context.Context, rc ctl.Command) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	wi := &WithInit{
		inner: rc,
		run:   r,
		ctx:   ctx,

		done: make(chan struct{}),
	}

	// create pipes for stdout, stderr
	wi.stdout.r, wi.stdout.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	wi.stderr.r, wi.stderr.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	// react to done from inner control
	go func() {
		<-wi.inner.Done()
		<-wi.initDone
		close(wi.done)
	}()

	return wi, nil
}

func (w *WithInit) Name() string {
	return w.inner.Name()
}

func (w *WithInit) Restart() (err error) {
	// wait for init to shutdown or deadline is reached.
	select {
	case <-w.initDone:
	case <-time.After(15 * time.Second): // FIXME we need a consistent deadline for all WithInit
	}

	err = w.inner.Restart()
	if err != nil {
		return err
	}
	return w.init()
}

func (w *WithInit) Start() (err error) {
	defer errz.Recover(&err)

	err = w.inner.Start()
	errz.Fatal(err)

	// Wait for initial command to have started
	for !w.inner.Running() {
		time.Sleep(100 * time.Millisecond)
	}

	return w.init()
}

func (w *WithInit) Stop() error {
	w.mux.Lock()
	if w.initRunning {
		w.initCtxCancel()
	}
	w.mux.Unlock()

	// wait for init to shutdown or deadline is reached.
	select {
	case <-w.initDone:
	case <-time.After(5 * time.Second):
	}

	return w.inner.Stop()
}

func (w *WithInit) Running() bool {
	return w.inner.Running()
}

func (w *WithInit) Shutdown() (err error) {
	defer errz.Recover(&err)
	w.mux.Lock()
	if w.initRunning {
		w.initCtxCancel()
	}
	w.mux.Unlock()

	return w.inner.Shutdown()
}

func (w *WithInit) Done() <-chan struct{} {
	return w.done
}

func (w *WithInit) Stdout() io.Reader {
	return io.MultiReader(w.stdout.r, w.inner.Stdout())
}
func (w *WithInit) Stderr() io.Reader {
	return io.MultiReader(w.stderr.r, w.inner.Stderr())
}
func (w *WithInit) Stdin() io.Writer {
	return w.inner.Stdin()
}

// init runs the `initOnce` and `init` cmds.
// Reacts to the external context and takes care of setting
// the state of the control.
func (w *WithInit) init() (err error) {

	w.mux.Lock()
	defer w.mux.Unlock()

	if w.initRunning {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.initCtxCancel = cancel
	w.initRunning = true
	w.initDone = make(chan struct{})

	// React to cancel from external context
	go func() {
		select {
		case <-w.ctx.Done():
		case <-w.initDone: // assures this goroutine exits correctly
		}
		w.initCtxCancel()
	}()

	go func() {
		defer func() {
			w.mux.Lock()
			w.initRunning = false
			w.mux.Unlock()
			close(w.initDone)
		}()

		// InitOnce.
		if len(w.run.InitOnce()) > 0 {
			var onceErr error
			w.once.Do(
				func() {
					boblog.Log.Info(fmt.Sprintf("InitOnce [%s] ", w.inner.Name()))
					onceErr = w.shexec(ctx, w.run.InitOnce())
				},
			)
			if onceErr != nil {
				boblog.Log.V(1).Error(onceErr, "")
				return
			}
		}

		// Init.
		if len(w.run.Init()) > 0 {
			boblog.Log.Info(fmt.Sprintf("Init [%s] ", w.inner.Name()))
			err = w.shexec(ctx, w.run.Init())
			if err != nil {
				boblog.Log.V(1).Error(err, "")
				return
			}
		}
	}()

	return nil
}

func (w *WithInit) shexec(ctx context.Context, cmds []string) (err error) {
	defer errz.Recover(&err)

	for _, run := range cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		errz.Fatal(err)

		// FIXME: make run cmds ready for nix integration.
		env := os.Environ()

		pr, pw, err := os.Pipe()
		errz.Fatal(err)

		s := bufio.NewScanner(pr)
		s.Split(bufio.ScanLines)

		go func() {
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				// FIXME: why does printing Commands to stdout not work?
				// fmt.Fprintf(w.stdout.w, "\t%s\n", aurora.Faint(s.Text()))
				boblog.Log.V(1).Info(fmt.Sprintf("\t%s", aurora.Faint(s.Text())))
			}
		}()

		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(w.run.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, pw, pw),
			// FIXME: why does this not work?
			// interp.StdIO(os.Stdin, w.stdout.w, w.stderr.w),
		)
		errz.Fatal(err)

		err = r.Run(ctx, p)
		errz.Fatal(err)
	}

	return nil
}
