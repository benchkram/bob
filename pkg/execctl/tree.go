package execctl

import (
	"context"
	"github.com/Benchkram/bob/pkg/ctl"
	"golang.org/x/sync/errgroup"
)

type CommandTree interface {
	ctl.Command

	Subcommands() []CommandTree
}

// assert cmdTree implements the CommandTree interface
var _ CommandTree = (*cmdTree)(nil)

type cmdTree struct {
	*Cmd

	subcommands []*cmdTree
}

func NewCmdTree(root *Cmd, subcommands ...*Cmd) *cmdTree {
	cmds := make([]*cmdTree, 0)

	for _, cmd := range subcommands {
		cmds = append(cmds, NewCmdTree(cmd))
	}

	return &cmdTree{Cmd: root, subcommands: cmds}
}

func (c *cmdTree) Subcommands() []CommandTree {
	subs := make([]CommandTree, 0)

	for _, cmd := range c.subcommands {
		subs = append(subs, CommandTree(cmd))
	}

	return subs
}

// Start starts the command if it's not already running. It will return ErrCmdAlreadyStarted if it is.
func (c *cmdTree) Start() error {
	// use errgroup to speed up startup of the subcommands (run startup of them in parallel)
	g, _ := errgroup.WithContext(context.Background())

	for _, sub := range c.subcommands {
		// required or ref is replaced within loop
		subsub := sub

		g.Go(func() error {
			return subsub.Start()
		})
	}

	err := g.Wait()
	if err != nil {
		return err
	}

	return c.Cmd.Start()
}

// Stop stops the running command with an os.Interrupt signal. It does not return an error if the command has
// already exited gracefully.
func (c *cmdTree) Stop() error {
	err := c.Cmd.Stop()
	if err != nil {
		return err
	}

	g, _ := errgroup.WithContext(context.Background())

	for _, sub := range c.subcommands {
		// required or ref is replaced within loop
		subsub := sub

		g.Go(func() error {
			return subsub.Stop()
		})
	}

	return g.Wait()
}

// Restart first interrupts the command if it's already running, and then re-runs the command.
func (c *cmdTree) Restart() error {
	err := c.Stop()
	if err != nil {
		return err
	}

	return c.Start()
}
