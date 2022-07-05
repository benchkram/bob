package ctl

import (
	"context"
	"fmt"
	"io"

	"github.com/benchkram/errz"
)

var ErrInProgress = fmt.Errorf("in progress")
var ErrDone = fmt.Errorf("commander is done and can't be started")

type Commander interface {
	CommandTree
}

// commander allows managing multiple controls
type commander struct {
	// ctx to listen for execution interruption
	ctx context.Context

	// builder can trigger a rebuild.
	builder Builder

	// control is used to control the commander.
	control Control

	// commands are the children the commander controls.
	commands []Command

	// starting  blocks subssequent starting requests.
	starting *Flag
	// stopping blocks subssequent stopping requests.
	stopping *Flag
	// restarting blocks subssequent stopping requests.
	restarting *Flag

	// done indicates that the commander becomes noop.
	done bool

	// doneChan emits when the commander becomes noop.
	doneChan chan struct{}
}

type Builder interface {
	Build(context.Context) error
}

// NewCommander creates a commander object which can be started and stopped
// until shutdown is called, then it becomes noop.
//
// The commander allows it to control multiple commands while taking
// orders from a higher level instance like a TUI.
//
// TODO: Could be beneficial for a TUI to directly control the commands.
//       That needs somehow blocking of a starting/stopping of the whole commander
//       while a child is doing some work. This is currently not implemented.
//       It is possible to control the underlying commands directly through
//       `Subcommands()` but that could probably lead to nasty start/stop loops.
//  ___________             ___________             ___________
// |           | Command() |           | Command() |           |
// | n*command | *-------1 | commander |1--------1 |    tui    |
// |           |           |           |           |           |
// |___________|           |___________|           |___________|
//
func NewCommander(ctx context.Context, builder Builder, ctls ...Command) Commander {
	c := &commander{
		ctx: ctx,

		builder: builder,

		control:  New("commander", 0, nil, nil, nil),
		commands: ctls,

		starting:   &Flag{},
		stopping:   &Flag{},
		restarting: &Flag{},

		doneChan: make(chan struct{}),
	}

	// Listen on the control for external cmds
	go func() {
		<-ctx.Done()
		// wait till all cmds are done
		<-c.Done()
		c.control.EmitDone()
	}()

	// Shutdown each control
	// on a canceled context
	go func() {
		<-ctx.Done()
		c.shutdown()
	}()

	return c
}

// Subcommands allows direct access to the underlying commands.
// !!!Should used with care!!!
// See the comment from `NewCommander()`
func (c *commander) Subcommands() []Command {
	return c.commands
}

// Start cmds in inverse order.
// Blocks subsequent calls until the first one is completed.
func (c *commander) Start() (err error) {
	defer errz.Recover(&err)

	err = c.builder.Build(c.ctx)
	errz.Fatal(err)

	return c.start()
}

func (c *commander) start() (err error) {
	if c.done {
		return ErrDone
	}

	done, err := c.starting.InProgress()
	if err != nil {
		return err
	}
	defer done()

	for i := len(c.commands) - 1; i >= 0; i-- {
		ctl := c.commands[i]
		err = ctl.Start()
		if err != nil {
			return err
		}
	}

	return err
}

// Stop children from top to bottom.
// Blocks subsquent calls until the first one is completed.
func (c *commander) Stop() (err error) {
	return c.stop()
}

// stop children, starting from top.
func (c *commander) stop() (err error) {
	done, err := c.stopping.InProgress()
	if err != nil {
		return err
	}
	defer done()

	for _, cmd := range c.commands {
		if e := cmd.Stop(); err != nil {
			err = stackErrors(err, e)
		}
	}

	return err
}

func (c *commander) Done() <-chan struct{} {
	return c.doneChan
}

// shutdown forwards the signal to the children.
func (c *commander) shutdown() {
	for _, v := range c.commands {
		_ = v.Shutdown()
	}
	c.done = true
	close(c.doneChan)
}

func (c *commander) Name() string {
	return c.control.Name()
}
func (c *commander) Restart() error {
	done, err := c.restarting.InProgress()
	if err != nil {
		return err
	}
	defer done()

	// Trigger a rebuild.
	err = c.builder.Build(c.ctx)
	if err != nil {
		return err
	}

	for i := 0; i < len(c.commands); i++ {
		ctl := c.commands[i]
		err = ctl.Restart()
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *commander) Running() bool {
	return c.control.Running()
}
func (c *commander) Shutdown() error {
	return c.control.Shutdown()
}
func (c *commander) Stdout() io.Reader {
	return c.control.Stdout()
}
func (c *commander) Stderr() io.Reader {
	return c.control.Stderr()
}
func (c *commander) Stdin() io.Writer {
	return c.control.Stdin()
}
