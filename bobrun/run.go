package bobrun

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/pkg/ctl"
	"github.com/Benchkram/bob/pkg/execctl"
)

var ErrInvalidRunType = fmt.Errorf("Invalid run type")

type Run struct {
	Type RunType

	// ComposePath is the path to a docker-compose file or binary
	// Default filename is used when empty.
	Path string

	// DependsOn run or build tasks
	DependsOn []string

	// didUpdate fires after the run-task
	// did a restart.
	didUpdate chan struct{}

	dir string

	name string
}

func (r *Run) Name() string {
	return r.name
}

func (r *Run) SetName(name string) {
	r.name = name
}

func (r *Run) Dir() string {
	return r.dir
}

func (r *Run) SetDir(dir string) {
	r.dir = dir
}

func New() *Run {
	r := &Run{
		Type:      RunTypeBinary,
		DependsOn: []string{},
		Path:      composeFileDefault,

		didUpdate: make(chan struct{}),
	}
	return r
}

// Run creates run cmds and return a channel to ctl it.
// To shutdown a Run() use a cancable context.
func (r *Run) Run(ctx context.Context) (rc ctl.Command, _ error) {
	fmt.Printf("Creating control for run task [%s]\n", r.name)

	switch r.Type {
	case RunTypeBinary:
		return execctl.NewCmd(r.Path, r.Path)
	case RunTypeCompose:
		return r.composeCommand(ctx)
	default:
		return nil, ErrInvalidRunType
	}
}
