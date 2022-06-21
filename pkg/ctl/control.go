package ctl

import (
	"io"
	"os"
)

type Control interface {
	Command

	Control() <-chan Signal

	EmitDone()
	EmitStopped()
	EmitStarted()
	EmitRestarted()
	EmitError(error)
}

// assert control implements the Command & Control interface
var _ Command = (*control)(nil)
var _ Control = (*control)(nil)

type control struct {
	name string

	stdout io.Reader
	stderr io.Reader
	stdin  io.Writer

	ctl       chan Signal
	done      chan struct{}
	started   chan struct{}
	stopped   chan struct{}
	restarted chan struct{}
	err       chan error
}

func (c *control) Running() bool {
	// TODO: @Tasos that needs to be somehow implemented for the TUI
	// panic("implement me")
	return true

}

// New takes the size of the control channel.
// Usually this should be 0 so that signals are ignored as long
// as a start/stop is in progress.
func New(name string, bufferedctl int, stdout, stderr io.Reader, stdin io.Writer) Control {
	return &control{
		name: name,

		stdout: stdout,
		stderr: stderr,
		stdin:  stdin,

		// ctl recives external signals
		ctl: make(chan Signal, bufferedctl),

		started:   make(chan struct{}, 1),
		stopped:   make(chan struct{}, 1),
		restarted: make(chan struct{}, 1),
		done:      make(chan struct{}),

		err: make(chan error, 1),
	}
}

// Name of the service the runCtrl controls
func (c *control) Name() string {
	return c.name
}

func (c *control) Start() (err error) {
	select {
	case c.ctl <- Start:
	default:
		return nil // Ignoring signal
	}

	select {
	case <-c.started:
	case e := <-c.err:
		return e
	}

	return nil
}

func (c *control) Stop() (err error) {
	select {
	case c.ctl <- Stop:
	default:
		return nil // Ignoring signal
	}

	select {
	case <-c.stopped:
	case e := <-c.err:
		return e
	}

	return nil
}

func (c *control) Shutdown() (err error) {
	select {
	case c.ctl <- Shutdown:
	default:
		return nil
	}

	select {
	case <-c.done:
	case e := <-c.err:
		return e
	}

	return nil
}

func (c *control) Restart() (err error) {
	select {
	case c.ctl <- Restart:
	default:
		return nil
	}

	select {
	case <-c.restarted:
	case e := <-c.err:
		return e
	}

	return nil
}

func (c *control) Control() <-chan Signal {
	return c.ctl
}
func (c *control) Done() <-chan struct{} {
	return c.done
}

func (c *control) Error() <-chan error {
	return c.err
}

// EmitStop signals that the cmd has finished.
func (c *control) EmitDone() {
	close(c.done)
}

// EmitStarted signals that the cmd has been started.
func (c *control) EmitStarted() {
	select {
	case c.started <- struct{}{}:
	default:
	}
}

// EmitRestarted signals that the cmd has been restarted.
func (c *control) EmitRestarted() {
	select {
	case c.restarted <- struct{}{}:
	default:
	}
}

// EmitStop signals that the cmd has stopped
// and can be restarted.
func (c *control) EmitStopped() {
	select {
	case c.stopped <- struct{}{}:
	default:
	}
}

// EmitError emits a error to the error channel.
func (c *control) EmitError(err error) {
	select {
	case c.err <- err:
	default:
	}
}

func (c *control) Stdout() io.Reader {
	return c.stdout
}

func (c *control) Stderr() io.Reader {
	return c.stderr
}

func (c *control) Stdin() io.Writer {
	return c.stdin
}

func (c *control) Path() string {
	return os.Getenv("PATH")
}
