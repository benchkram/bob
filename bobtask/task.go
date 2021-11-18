package bobtask

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/logrusorgru/aurora"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Benchkram/bob/bobtask/export"
	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/bob/pkg/filehash"

	"gopkg.in/yaml.v3"
)

type Tasker interface {
	Pack() error
	Unpack() error
}

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
	TargetDirty target.T `yaml:"target"`
	target      *target.T

	// Exports other tasks can reuse.
	Exports export.Map `yaml:"exports"`

	// name is the name of the task
	name string
	// dir is the working directory for this task
	dir string

	// env holds key=value pairs passed to the environement
	// when the task is executed.
	env   []string

	// Color is used to color the task's name on the terminal
	Color aurora.Color
}

func Make(opts ...TaskOption) Task {
	t := Task{
		TargetDirty: target.Make(),

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

func (t *Task) ColoredName() string {
	return aurora.Colorize(t.Name(), t.Color).String()
}

func (t *Task) Env() []string {
	return t.env
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

const EnvironSeparator = "="

func (t *Task) AddEnvironment(key, value string) {
	t.env = append(t.env, strings.Join([]string{key, value}, EnvironSeparator))
}
func (t *Task) AddExportPrefix(prefix string) {
	for i, e := range t.Exports {
		t.Exports[i] = export.E(filepath.Join(prefix, string(e)))
	}
}

// Hash computes a aggregated hash of all input files.
func (t *Task) Hash() (computedhash *hash.Task, _ error) {
	aggregatedHashes := bytes.NewBuffer([]byte{})

	// Hash input files
	for _, f := range t.inputs {
		h, err := filehash.Hash(f)
		if err != nil {
			return computedhash, fmt.Errorf("failed to hash file %q: %w", f, err)
		}

		_, err = aggregatedHashes.Write(h)
		if err != nil {
			return computedhash, fmt.Errorf("failed to write file hash to aggregated hash %q: %w", f, err)
		}
	}

	// Hash the public task description
	description, err := yaml.Marshal(t)
	if err != nil {
		return computedhash, fmt.Errorf("failed to marshal task: %w", err)
	}
	descriptionHash, err := filehash.HashBytes(bytes.NewBuffer(description))
	if err != nil {
		return computedhash, fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(descriptionHash)
	if err != nil {
		return computedhash, fmt.Errorf("failed to write task description to aggregated hash: %w", err)
	}

	// Hash the environment
	sort.Strings(t.env)
	environment := strings.Join(t.env, ",")
	environmentHash, err := filehash.HashBytes(bytes.NewBufferString(environment))
	if err != nil {
		return computedhash, fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(environmentHash)
	if err != nil {
		return computedhash, fmt.Errorf("failed to write task environment to aggregated hash: %w", err)
	}

	// Summarize
	h, err := filehash.HashBytes(aggregatedHashes)
	if err != nil {
		return computedhash, fmt.Errorf("failed to write aggregated hash: %w", err)
	}
	return &hash.Task{Input: hex.EncodeToString(h), Targets: make(hash.Targets)}, nil
}
