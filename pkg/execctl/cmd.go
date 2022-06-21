package execctl

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"
)

var (
	ErrCmdAlreadyStarted = errors.New("cmd already started")
)

// Command allows to execute a command and interact with it in real-time
// type Command interface {
// 	Start() error
// 	Stop() error
// 	Restart() error

// 	Shutdown() error
// 	Done() <-chan struct{}

// 	Stdout() io.Reader
// 	Stderr() io.Reader
// 	Stdin() io.Writer
// }

// assert Cmd implements the Command interface
var _ ctl.Command = (*Cmd)(nil)

// Cmd allows to control a process started through os.Exec with additional start, stop and restart capabilities, and
// provides readers/writers for the command's outputs and input, respectively.
type Cmd struct {
	mux         sync.Mutex
	cmd         *exec.Cmd
	name        string
	exe         string
	args        []string
	stdout      pipe
	stderr      pipe
	stdin       pipe
	running     bool
	interrupted bool
	err         chan error
	lastErr     error
	// path is the $PATH cmd
	path string
}

type pipe struct {
	r *os.File
	w *os.File
}

// NewCmd creates a new Cmd, ready to be started
func NewCmd(name string, exe, path string, args ...string) (c *Cmd, err error) {
	c = &Cmd{
		name: name,
		exe:  exe,
		args: args,
		err:  make(chan error, 1),
	}

	if path != "" {
		c.path = path
	} else {
		c.path = os.Getenv("PATH")
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

	err := os.Setenv("PATH", c.path)
	if err != nil {
		return err
	}

	if c.running {
		return nil
	}

	c.running = true
	c.interrupted = false
	c.lastErr = nil

	// create the command with the found executable and the its args
	cmd := exec.Command(c.exe, c.args...)
	c.cmd = cmd

	// assign the pipes to the command
	c.cmd.Stdout = c.stdout.w
	c.cmd.Stderr = c.stderr.w
	c.cmd.Stdin = c.stdin.r

	// start the command
	err = c.cmd.Start()
	if err != nil {
		return usererror.Wrapm(err, "Command execution failed")
	}

	go func() {
		err = cmd.Wait()

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
	cmd := c.cmd
	c.interrupted = true

	c.mux.Unlock()

	if !running {
		return nil
	}

	if cmd != nil && cmd.Process != nil {
		// send an interrupt signal to the command
		err := cmd.Process.Signal(os.Interrupt)
		if err != nil && !strings.Contains(err.Error(), "os: process already finished") {
			return err
		}
	}

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

// Path returns the $PATH of the command
func (c *Cmd) Path() string {
	return c.path
}
