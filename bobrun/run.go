package bobrun

import (
	"context"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/execctl"
	"github.com/benchkram/bob/pkg/nix"
)

var ErrInvalidRunType = fmt.Errorf("invalid run type")

type Run struct {
	Type RunType

	// ComposePath is the path to a docker-compose file or binary
	// Default filename is used when empty.
	Path string

	// DependsOn run or build tasks
	DependsOn []string

	// InitDirty runs run after this task has started and `initOnce`conpleted.
	InitDirty string `yaml:"init"`
	// init see InitDirty
	init []string

	// InitOnceDirty runs once during the lifetime of a run
	// after the actual task has started.
	InitOnceDirty string `yaml:"initOnce"`
	// initOnce see InitOnceDirty
	initOnce []string

	// DependenciesDirty read from the bobfile
	DependenciesDirty []string `yaml:"dependencies"`

	// dependencies contain the actual dependencies merged
	// with the global dependencies defined in the Bobfile
	// in the order which they need to be added to PATH
	dependencies []nix.Dependency

	// storePaths contain /nix/store/* paths
	// in the order which they need to be added to PATH
	storePaths []string

	// flag if its bobfile has Nix enabled
	useNix bool

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

func (r *Run) SetUseNix(useNix bool) {
	r.useNix = useNix
}

func (r *Run) UseNix() bool {
	return r.useNix
}

func (r *Run) Dependencies() []nix.Dependency {
	return r.dependencies
}
func (r *Run) SetDependencies(dependencies []nix.Dependency) {
	r.dependencies = dependencies
}

func (r *Run) SetStorePaths(storePaths []string) {
	r.storePaths = storePaths
}

// HasNixStorePaths checks if the run has /nix/store paths
// to be added to $PATH
func (r *Run) HasNixStorePaths() bool {
	return len(r.storePaths) > 0 && r.UseNix()
}

// Command creates a run cmd and returns a Command interface to control it.
// To shutdown a Command() use a cancelable context.
func (r *Run) Command(ctx context.Context) (rc ctl.Command, err error) {
	defer errz.Recover(&err)
	fmt.Printf("Creating control for run task [%s]\n", r.name)

	switch r.Type {
	case RunTypeBinary:
		rc, err = execctl.NewCmd(r.name, r.Path, r.storePaths, r.UseNix())
		errz.Fatal(err)
	case RunTypeCompose:
		rc, err = r.composeCommand(ctx)
		errz.Fatal(err)
	default:
		return nil, ErrInvalidRunType
	}

	rc, err = r.WrapWithInit(ctx, rc)
	errz.Fatal(err)

	return rc, nil
}
