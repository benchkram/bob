package execctl

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	ErrCmdAlreadyStarted = errors.New("cmd already started")
)

// Command allows to execute a command and interact with it in real-time
type Command interface {
	Start() error
	Stop() error
	Restart() error
	Wait() error
	Stdout() io.Reader
	Stderr() io.Reader
	Stdin() io.Writer
}

// assert Cmd implements the Command interface
var _ Command = (*Cmd)(nil)

// Cmd allows to control a process started through os.Exec with additional start, stop and restart capabilities, and
// provides readers/writers for the command's outputs and input, respectively.
type Cmd struct {
	mux     sync.Mutex
	cmd     *exec.Cmd
	exe     string
	args    []string
	stdout  pipe
	stderr  pipe
	stdin   pipe
	running bool
	err     chan error
	lastErr error
}

type pipe struct {
	w *os.File
	r *os.File
}

// NewCmd creates a new Cmd, ready to be started
func NewCmd(exe string, args ...string) (*Cmd, error) {
	return &Cmd{
		exe:  exe,
		args: args,
	}, nil
}

// Start starts the command if it's not already running. It will return ErrCmdAlreadyStarted if it is.
func (c *Cmd) Start() error {
	return c.start()
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

	return c.start()
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

	if !running {
		c.mux.Unlock()

		return lastErr
	}

	c.mux.Unlock()

	err := <-errChan

	c.mux.Lock()

	c.lastErr = err

	c.mux.Unlock()

	return err
}

// start creates the command and starts executing it. It also spins up a goroutine that will receive any error occurred
// during the command's exit.
func (c *Cmd) start() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.running {
		return ErrCmdAlreadyStarted
	}

	c.running = true
	c.lastErr = nil

	c.err = make(chan error, 1)

	// create the command with the found executable and the its args
	cmd := exec.Command(c.exe, c.args...)
	c.cmd = cmd

	// create pipes for stdout, stderr and stdin
	var err error
	c.stdout.r, c.stdout.w, err = os.Pipe()
	if err != nil {
		return err
	}

	c.stderr.r, c.stderr.w, err = os.Pipe()
	if err != nil {
		return err
	}

	c.stdin.r, c.stdin.w, err = os.Pipe()
	if err != nil {
		return err
	}

	// assign the pipes to the command
	c.cmd.Stdout = c.stdout.w
	c.cmd.Stderr = c.stderr.w
	c.cmd.Stdin = c.stdin.r

	// start the command
	err = c.cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		err = cmd.Wait()

		c.err <- err

		c.mux.Lock()

		c.running = false
		_ = c.stdout.w.Close()
		_ = c.stderr.w.Close()
		_ = c.stdin.w.Close()

		close(c.err)

		c.mux.Unlock()
	}()

	return nil
}

// stop requests for the command to stop, if it has already started.
func (c *Cmd) stop() error {
	c.mux.Lock()

	running := c.running
	cmd := c.cmd

	c.mux.Unlock()

	if !running {
		return nil
	}

	// send an interrupt signal to the command
	err := cmd.Process.Signal(os.Interrupt)
	if err != nil && !strings.Contains(err.Error(), "os: process already finished") {
		return err
	}

	return nil
}
