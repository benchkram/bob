package build

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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
	InputDirty Input `yaml:"input"`
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
	TargetDirty Target `yaml:"target"`
	target      Target

	// Exports other tasks can reuse.
	Exports ExportMap

	// name is the name of the task
	name string
	// dir is the working directory for this task
	dir string

	// env holds key=value pairs passed to the environement
	// when the task is executed.
	env []string
}

func Make(opts ...TaskOption) Task {
	t := Task{
		InputDirty:  MakeInput(),
		TargetDirty: MakeTarget(),

		DependsOn: []string{},
		Exports:   make(ExportMap),
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

func (t *Task) Env() []string {
	return t.env
}

func (t *Task) GetExports() ExportMap {
	return t.Exports
}

func (t *Task) SetName(name string) {
	t.name = name
}

const EnvironSeparator = "="

func (t *Task) AddEnvironment(key, value string) {
	t.env = append(t.env, strings.Join([]string{key, value}, EnvironSeparator))
}
func (t *Task) AddExportPrefix(prefix string) {
	for i, export := range t.Exports {
		t.Exports[i] = Export(filepath.Join(prefix, string(export)))
	}
}

// Hash computes a aggregated hash of all input files.
func (t *Task) Hash() (string, error) {
	aggregatedHashes := bytes.NewBuffer([]byte{})

	// Hash input files
	for _, f := range t.inputs {
		h, err := filehash.Hash(f)
		if err != nil {
			return "", fmt.Errorf("failed to hash file %q: %w", f, err)
		}

		_, err = aggregatedHashes.Write(h)
		if err != nil {
			return "", fmt.Errorf("failed to write file hash to aggregated hash %q: %w", f, err)
		}
	}

	// Hash the public task description
	description, err := yaml.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task: %w", err)
	}
	descriptionHash, err := filehash.HashBytes(bytes.NewBuffer(description))
	if err != nil {
		return "", fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(descriptionHash)
	if err != nil {
		return "", fmt.Errorf("failed to write task description to aggregated hash: %w", err)
	}

	// Hash the environment
	sort.Strings(t.env)
	environment := strings.Join(t.env, ",")
	environmentHash, err := filehash.HashBytes(bytes.NewBufferString(environment))
	if err != nil {
		return "", fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(environmentHash)
	if err != nil {
		return "", fmt.Errorf("failed to write task environment to aggregated hash: %w", err)
	}

	h, err := filehash.HashBytes(aggregatedHashes)
	if err != nil {
		return "", fmt.Errorf("failed to write aggregated hash: %w", err)
	}

	return hex.EncodeToString(h), nil
}
