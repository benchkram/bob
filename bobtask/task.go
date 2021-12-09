package bobtask

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"

	"github.com/Benchkram/bob/bobtask/export"
	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/bob/pkg/buildinfostore"
	"github.com/Benchkram/bob/pkg/store"
)

type RebuildType string

const (
	RebuildAlways   RebuildType = "always"
	RebuildOnChange RebuildType = "on-change"
)

type Task struct {
	// Inputs are directorys or files
	// the task monitors for a rebuild.

	// InputDirty is the representation read from a bobfile.
	InputDirty string `yaml:"input"`
	// inputs is filtered by ignored & sanitized
	inputs []string

	CmdDirty string `yaml:"cmd"`
	// The cmds passed to os.Exec
	cmds []string

	// DependsOn are task which must succeede before this task
	// can run.
	DependsOn []string

	// TODO: Shall we add a optional environment?
	// Like a docker image which can be used to build a target.
	// It's more effort but allows for more or less fixed build tool
	// versions acros systems.
	//
	// Another options would be to provide versions for a
	// task and build tool.. But each build tool needs manual
	// handling to figure out it's version.
	//
	// !!Needs Decission!!
	Environment string

	// Target defines the output of a task.
	//
	// ??? (unsure)
	// Binary or Directory:
	// Can be a internal file or directory.
	// Parent tasks can take the files and copy them
	// to a place they like to
	// ???
	TargetDirty string `yaml:"target,omitempty"`
	target      *target.T

	// Exports other tasks can reuse.
	Exports export.Map `yaml:"exports"`

	// defines the rebuild strategy
	RebuildDirty string `yaml:"rebuild,omitempty"`
	rebuild      RebuildType

	// name is the name of the task
	// TODO: Make this public to allow yaml.Marshal to add this to the task hash?!?
	name string
	// dir is the working directory for this task
	dir string

	// env holds key=value pairs passed to the environement
	// when the task is executed.
	env []string

	// hashIn stores the `In` has for reuse
	hashIn *hash.In

	// local store for artifacts
	local store.Store

	// buildInfoStore stores buildinfos.
	buildInfoStore buildinfostore.Store

	// color is used to color the task's name on the terminal
	color aurora.Color
}

func Make(opts ...TaskOption) Task {
	t := Task{
		DependsOn: []string{},
		Exports:   make(export.Map),
		env:       []string{},
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&t)
	}

	return t
}

func (t *Task) Dir() string {
	return t.dir
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) ShortName() string {
	_, name := filepath.Split(t.name)
	return name
}

func (t *Task) SetColor(color aurora.Color) {
	t.color = color
}

func (t *Task) ColoredName() string {
	return aurora.Colorize(t.Name(), t.color).String()
}

func (t *Task) Env() []string {
	return t.env
}

func (t *Task) Rebuild() RebuildType {
	return t.rebuild
}

func (t *Task) GetExports() export.Map {
	return t.Exports
}

func (t *Task) SetDir(dir string) {
	t.dir = dir
}

func (t *Task) SetName(name string) {
	t.name = name
}

func (t *Task) SetEnv(env []string) {
	t.env = env
}

// Set the rebuild strategy for the task
// defaults to `on-change`.
func (t *Task) SetRebuildStrategy(rebuild RebuildType) {
	t.rebuild = rebuild
}

func (t *Task) WithLocalstore(s store.Store) *Task {
	t.local = s
	return t
}

func (t *Task) WithBuildinfoStore(s buildinfostore.Store) *Task {
	t.buildInfoStore = s
	return t
}

const EnvironSeparator = "="

func (t *Task) AddEnvironment(key, value string) {
	t.env = append(t.env, strings.Join([]string{key, value}, EnvironSeparator))
}
func (t *Task) AddExportPrefix(prefix string) {
	for i, e := range t.Exports {
		t.Exports[i] = export.E(filepath.Join(prefix, string(e)))
	}
}

func (t *Task) parseTargets() error {
	targetDirty := split(t.TargetDirty)
	targets := []string{}

	for _, targetPath := range unique(targetDirty) {
		if strings.Contains(targetPath, "..") {
			return fmt.Errorf("'..' not allowed in file path %q", targetPath)
		}

		targets = append(targets, targetPath)
	}

	if len(targets) > 0 {
		t.target = target.New()
		t.target.Paths = targets
	}

	return nil
}
