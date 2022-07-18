package bobrun

import (
	"context"
	"fmt"
	"strings"

	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/execctl"
	"github.com/benchkram/errz"
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

	// didUpdate fires after the run task
	// did a restart.
	didUpdate chan struct{}

	dir string

	// env holds key=value pairs passed to the environment
	// when the task is executed.
	env []string

	name string

	// flag if its bobfile has Nix enabled
	useNix bool
}

func (r *Run) Name() string {
	return r.name
}

func (r *Run) SetEnv(env []string) {
	r.env = env
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

const EnvironSeparator = "="

func (r *Run) AddEnvironmentVariable(key, value string) {
	r.env = append(r.env, strings.Join([]string{key, value}, EnvironSeparator))
}

func (r *Run) AddEnvironment(env []string) {
	r.env = append(r.env, env...)
}

func (r *Run) SetUseNix(useNix bool) {
	r.useNix = useNix
}

func (r *Run) UseNix() bool {
	return r.useNix
}

// Command creates a run cmd and returns a Command interface to control it.
// To shutdown a Command() use a cancelable context.
func (r *Run) Command(ctx context.Context) (rc ctl.Command, err error) {
	defer errz.Recover(&err)
	fmt.Printf("Creating control for run task [%s]\n", r.name)

	switch r.Type {
	case RunTypeBinary:
		rc, err = execctl.NewCmd(r.name, r.Path)
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
