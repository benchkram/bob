package shctl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/interp"
	"mvdan.cc/sh/syntax"
)

var (
	ErrCmdAlreadyStarted = errors.New("cmd already started")
)

// assert Cmd implements the Command interface
var _ ctl.Command = (*Cmd)(nil)

// Cmd allows to control a process started through os.Exec with additional start, stop and restart capabilities, and
// provides readers/writers for the command's outputs and input, respectively.
type Cmd struct {
	mux sync.Mutex

	// script to execute
	script string

	// dir to execute the script
	dir string

	// ctx used to cancel script execution
	ctx           context.Context
	ctxCancelFunc context.CancelFunc

	name        string
	args        []string
	stdout      pipe
	stderr      pipe
	stdin       pipe
	running     bool
	interrupted bool
	err         chan error
	lastErr     error
}

type pipe struct {
	r *os.File
	w *os.File
}

// NewCmd creates a new Cmd, ready to be started
func New(name, dir, script string, args ...string) (c *Cmd, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	c = &Cmd{
		name:          name,
		script:        script,
		dir:           dir,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		args:          args,
		err:           make(chan error, 1),
	}

	// create pipes for stdout, stderr and stdin
	c.stdout.r, c.stdout.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	c.stderr.r, c.stderr.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	c.stdin.r, c.stdin.w, err = os.Pipe()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Cmd) sh(ctx context.Context, dir string, cmd string) (err error) {

	defer errz.Recover(&err)

	env := os.Environ()
	// // TODO: warn when overwriting envvar from the environment
	// env = append(env, t.env...)

	// if len(t.storePaths) > 0 && t.useNix {
	// 	for k, v := range env {
	// 		pair := strings.SplitN(v, "=", 2)
	// 		if pair[0] == "PATH" {
	// 			env[k] = "PATH=" + strings.Join(nix.StorePathsBin(t.storePaths), ":")
	// 		}
	// 	}
	// }

	boblog.Log.Info(fmt.Sprintf("sh: running script dir[%s], script[%s]", c.dir, c.script))

	p, err := syntax.NewParser().Parse(strings.NewReader(c.script), "")
	if err != nil {
		return usererror.Wrapm(err, "shell command parse error")
	}

	boblog.Log.Info("sh: parser created")

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	boblog.Log.Info("sh: pipe created")

	s := bufio.NewScanner(pr)
	s.Split(bufio.ScanLines)

	doneReading := make(chan bool)

	go func() {
		for s.Scan() {
			err := s.Err()
			if err != nil {
				return
			}

			//boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t  %s", namePad, t.ColoredName(), aurora.Faint(s.Text())))
			println(s.Text())
		}

		doneReading <- true
	}()

	r, err := interp.New(
		interp.Params("-e"),
		interp.Dir(c.dir),

		interp.Env(expand.ListEnviron(env...)),
		interp.StdIO(c.stdin.r, c.stdout.w, c.stderr.r),
	)

	// TODO: remove
	if err != nil {
		boblog.Log.Error(err, "new interpreter")
	}
	// FIXME: this is not caught
	errz.Fatal(err)

	boblog.Log.Info("sh: calling run")

	err = r.Run(ctx, p)
	if err != nil {
		return usererror.Wrapm(err, "shell command execute error")
	}

	boblog.Log.Info("sh: running")

	// wait for the reader to finish after closing the write pipe
	pw.Close()
	<-doneReading

	return nil

}

func (c *Cmd) Name() string {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.name
}

func (c *Cmd) Running() bool {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.running
}

// Start starts the command if it's not already running. It will be a noop if it is.
// It also spins up a goroutine that will receive any error occurred during the command's exit.
func (c *Cmd) Start() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.running {
		return nil
	}

	c.running = true
	c.interrupted = false
	c.lastErr = nil

	// start the command
	go func() {
		err := c.sh(c.ctx, "dir", "script/cmd")
		if err != nil {
			fmt.Fprintln(c.stderr.w, err.Error())
		}
		c.err <- err

		c.mux.Lock()

		c.running = false

		c.mux.Unlock()
	}()

	return nil
}

// Stop stops the running command with an os.Interrupt signal. It does not return an error if the command has
// already exited gracefully.
func (c *Cmd) Stop() error {
	err := c.stop()
	if err != nil {
		return err
	}

	return c.Wait()
}

// Restart first interrupts the command if it's already running, and then re-runs the command.
func (c *Cmd) Restart() error {
	err := c.stop()
	if err != nil {
		return err
	}

	err = c.Wait()
	if err != nil {
		return err
	}

	return c.Start()
}

// Stdout returns a reader to the command's stdout. The reader will return an io.EOF error if the command exits.
func (c *Cmd) Stdout() io.Reader {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.stdout.r
}

// Stderr returns a reader to the command's stderr. The reader will return an io.EOF error if the command exits.
func (c *Cmd) Stderr() io.Reader {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.stderr.r
}

// Stdin returns a writer to the command's stdin. The writer will be closed if the command has exited by the time this
// function is called.
func (c *Cmd) Stdin() io.Writer {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.stdin.w
}

// Wait awaits for the command to stop running (either to gracefully exit or be interrupted).
// If the command has already finished when Wait is invoked, it returns the error that was returned when the command
// exited, if any.
func (c *Cmd) Wait() error {
	c.mux.Lock()

	running := c.running
	errChan := c.err
	lastErr := c.lastErr
	interrupted := c.interrupted

	if !running {
		c.mux.Unlock()

		return lastErr
	}

	c.mux.Unlock()

	err := <-errChan

	c.mux.Lock()

	if err != nil && interrupted && strings.Contains(err.Error(), "signal: interrupt") {
		err = nil
	} else if err != nil {
		c.lastErr = err
	}

	c.mux.Unlock()

	return err
}

// stop requests for the command to stop, if it has already started.
func (c *Cmd) stop() error {
	c.mux.Lock()

	running := c.running
	c.interrupted = true

	c.mux.Unlock()

	if !running {
		return nil
	}

	c.ctxCancelFunc()

	return nil
}

// Shutdown stops the cmd
func (c *Cmd) Shutdown() error {
	return c.Stop()
}

func (c *Cmd) Done() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		_ = c.Wait()
		close(done)
	}()
	return done
}
