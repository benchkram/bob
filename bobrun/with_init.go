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

// WrapCommand takes a ctl to add init functionality defined in the run task.
func (r *Run) WrapWithInit(ctx context.Context, rc ctl.Command) (_ ctl.Command, err error) {
	defer errz.Recover(&err)

	rw := &WithInit{
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

func (rw *WithInit) Name() string {
	return rw.inner.Name()
}

func (rw *WithInit) Restart() (err error) {
	return rw.inner.Restart()
}

func (rw *WithInit) Start() (err error) {
	defer errz.Recover(&err)

	err = rw.inner.Start()
	errz.Fatal(err)

	// Wait for initial command to have started
	for !rw.inner.Running() {
		time.Sleep(100 * time.Millisecond)
	}

	return rw.init()
}

func (rw *WithInit) Stop() error {
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

func (rw *WithInit) Running() bool {
	return rw.inner.Running()
}

func (rw *WithInit) Shutdown() (err error) {
	defer errz.Recover(&err)
	rw.mux.Lock()
	if rw.initRunning {
		rw.initCtxCancel()
	}
	rw.mux.Unlock()

	return rw.inner.Shutdown()
}

func (rw *WithInit) Done() <-chan struct{} {
	return rw.done
}

func (rw *WithInit) Stdout() io.Reader {
	return io.MultiReader(rw.inner.Stdout(), rw.stdout.r)
}
func (rw *WithInit) Stderr() io.Reader {
	return io.MultiReader(rw.inner.Stderr(), rw.stderr.r)
}
func (rw *WithInit) Stdin() io.Writer {
	return rw.inner.Stdin()
}

// init runs the `initOnce` and `init` cmds.
// Reacts to the external context and takes care of setting
// the state of the control.
func (rw *WithInit) init() (err error) {

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

		defer func() {
			rw.mux.Lock()
			rw.initRunning = false
			rw.mux.Unlock()
			close(rw.initDone)
		}()

		// InitOnce.
		if len(rw.run.InitOnce()) > 0 {
			var onceErr error
			rw.once.Do(
				func() {
					boblog.Log.Info(fmt.Sprintf("InitOnce [%s] ", rw.inner.Name()))
					onceErr = rw.shexec(ctx, rw.run.InitOnce())
				},
			)
			if onceErr != nil {
				boblog.Log.V(1).Error(onceErr, "")
				return
			}
		}

		// Init.
		if len(rw.run.Init()) > 0 {
			boblog.Log.Info(fmt.Sprintf("Init [%s] ", rw.inner.Name()))
			err = rw.shexec(ctx, rw.run.Init())
			if err != nil {
				boblog.Log.V(1).Error(err, "")
				return
			}
		}
	}()

	return nil
}

func (rw *WithInit) shexec(ctx context.Context, cmds []string) (err error) {
	defer errz.Recover(&err)

	for _, run := range cmds {
		p, err := syntax.NewParser().Parse(strings.NewReader(run), "")
		errz.Fatal(err)

		env := rw.run.Env()
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
				// fmt.Fprintf(rw.stdout.w, "\t%s\n", aurora.Faint(s.Text()))
				boblog.Log.V(1).Info(fmt.Sprintf("\t%s", aurora.Faint(s.Text())))
			}
		}()

		r, err := interp.New(
			interp.Params("-e"),
			interp.Dir(rw.run.dir),

			interp.Env(expand.ListEnviron(env...)),
			interp.StdIO(os.Stdin, pw, pw),
			// FIXME: why does this not work?
			// interp.StdIO(os.Stdin, rw.stdout.w, rw.stderr.w),
		)
		errz.Fatal(err)

		err = r.Run(ctx, p)
		errz.Fatal(err)
	}

	return nil
}
