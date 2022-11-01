package bobrun

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/benchkram/errz"
	"gopkg.in/yaml.v3"

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
	DependsOn []string `yaml:"dependsOn"`

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

	nixpkgs string

	dir string

	// env holds key=value pairs passed to the environment
	// when the task is executed.
	env []string

	name string
}

func (r *Run) Name() string {
	return r.name
}

func (r *Run) SetEnv(env []string) {
	r.env = env
}

func (r *Run) SetNixpkgs(nixpkgs string) {
	r.nixpkgs = nixpkgs
}

func (r *Run) Env() []string {
	return r.env
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

func (r *Run) Dependencies() []nix.Dependency {
	return r.dependencies
}
func (r *Run) SetDependencies(dependencies []nix.Dependency) {
	r.dependencies = dependencies
}

func (r *Run) UnmarshalYAML(value *yaml.Node) (err error) {
	defer errz.Recover(&err)

	var values struct {
		Lowercase []string `yaml:"dependson"`
		Camelcase []string `yaml:"dependsOn"`
	}

	err = value.Decode(&values)
	errz.Fatal(err)

	if len(values.Lowercase) > 0 && len(values.Camelcase) > 0 {
		errz.Fatal(errors.New("both `dependson` and `dependsOn` nodes detected near line " + strconv.Itoa(value.Line)))
	}

	dependsOn := make([]string, 0)
	if values.Lowercase != nil && len(values.Lowercase) > 0 {
		dependsOn = values.Lowercase
	}
	if values.Camelcase != nil && len(values.Camelcase) > 0 {
		dependsOn = values.Camelcase
	}

	// new type needed to avoid infinite loop
	type TmpRun Run
	var tmpRun TmpRun

	err = value.Decode(&tmpRun)
	errz.Fatal(err)

	tmpRun.DependsOn = dependsOn

	*r = Run(tmpRun)

	return nil
}

// Command creates a run cmd and returns a Command interface to control it.
// To shutdown a Command() use a cancelable context.
func (r *Run) Command(ctx context.Context) (rc ctl.Command, err error) {
	defer errz.Recover(&err)
	fmt.Printf("Creating control for run task [%s]\n", r.name)

	switch r.Type {
	case RunTypeBinary:
		rc, err = execctl.NewCmd(
			r.name,
			r.Path,
			execctl.WithEnv(r.Env()),
		)
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
