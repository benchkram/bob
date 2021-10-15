package runctl

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/pkg/ctl"
)

var ErrInProgress = fmt.Errorf("in progress")
var ErrDone = fmt.Errorf("commander is done and can't be started")

type Commander interface {
	Start() error
	Stop() error

	Done() <-chan struct{}
}

// commander allows to manage mutliple controls
type commander struct {
	ctx context.Context

	controls []ctl.Command

	starting *Flag
	stopping *Flag

	done     bool
	doneChan chan struct{}
}

// NewCommander creates a commander object which can be started and stoped
// until shutdown is called, then it becomes noop.
func NewCommander(ctx context.Context, ctls ...ctl.Command) Commander {

	c := &commander{
		ctx:      ctx,
		controls: ctls,

		starting: &Flag{},
		stopping: &Flag{},

		doneChan: make(chan struct{}),
	}

	// Shutdown each control
	// on a canceled context
	go func() {
		<-ctx.Done()
		c.shutdown()
	}()

	return c
}

// Start cmds in inverse order.
// Blocks subsquent calls until the first one is completed.
func (c *commander) Start() (err error) {
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

	for i := len(c.controls) - 1; i >= 0; i-- {
		ctl := c.controls[i]
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

	for _, v := range c.controls {
		if e := v.Stop(); err != nil {
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
	for _, v := range c.controls {
		_ = v.Shutdown()
	}
	c.done = true
	close(c.doneChan)
}
