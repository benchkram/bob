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
	"github.com/Benchkram/bob/pkg/dockermoby"
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
	TargetDirty TargetEntry `yaml:"target,omitempty"`
	target      *target.T

	// Exports other tasks can reuse.
	Exports export.Map `yaml:"exports"`

	// defines the rebuild strategy
	RebuildDirty string `yaml:"rebuild,omitempty"`
	rebuild      RebuildType

	// name is the name of the task
	// TODO: Make this public to allow yaml.Marshal to add this to the task hash?!?
	name string

	// builder is the project who trigered the build
	builder string

	// project this tasks belongs to
	// TODO: todoproject: Currently it's the path.. later
	// we need globaly unique identifiers when using remote caching.
	project string

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

	// Handle all the request with docker local registry
	dockerRegistry dockermoby.RegistryHandler

	// skippedInputs is a lists of skipped input files
	skippedInputs []string
}

type TargetEntry interface{}

func Make(opts ...TaskOption) Task {
	t := Task{
		inputs:         []string{},
		cmds:           []string{},
		DependsOn:      []string{},
		Exports:        make(export.Map),
		env:            []string{},
		rebuild:        RebuildOnChange,
		dockerRegistry: dockermoby.New(),
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

func (t *Task) SetProject(proj string) {
	t.project = proj
}

func (t *Task) SetBuilder(builder string) {
	t.builder = builder
}

func (t *Task) SetDockerRegistry() {
	t.dockerRegistry = dockermoby.New()
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

// AddToSkippedInputs add filenames with permission issues to the task's
// skippedInput list. returns without appending if file
// already exists, thus maintain uniqueness
func (t *Task) addToSkippedInputs(f string) {
	for _, si := range t.skippedInputs {
		if si == f {
			return
		}
	}
	t.skippedInputs = append(t.skippedInputs, f)
}

// LogSkippedInput skipped input files from the task.
// prints nothing if there is not skipped input.
func (t *Task) LogSkippedInput() []string {
	return t.skippedInputs
}

func (t *Task) parseTargets() error {
	targetType := target.DefaultType // DefaultType set to Path currently

	var targets []string
	var err error

	switch t.TargetDirty.(type) {
	case string:
		targets, err = parseTargetPath(t.TargetDirty)
	case map[string]interface{}:
		targets, targetType, err = parseTargetMap(t.TargetDirty)
	default:
		targets, err = parseTargetPath(t.TargetDirty)
	}

	if err != nil {
		return err
	}

	if len(targets) > 0 {
		t.target = target.New()
		t.target.Paths = targets
		t.target.Type = targetType
	}

	return nil
}

func parseTargetMap(t interface{}) ([]string, target.TargetType, error) {
	mapped := t.(map[string]interface{})
	pathsI, ok := mapped["paths"]
	if !ok {
		return nil, target.DefaultType, fmt.Errorf("Can't find 'paths' on Target properties")
	}

	targets, err := parseTargetPath(pathsI)
	if err != nil {
		return nil, target.DefaultType, err
	}

	targetType := target.DefaultType
	typeI, ok := mapped["type"]
	if ok {
		typeStr := fmt.Sprintf("%v", typeI)
		targetType, err = target.ParseType(typeStr)
		if err != nil {
			return targets, target.DefaultType, err
		}
	}

	return targets, targetType, nil
}

func parseTargetPath(p interface{}) ([]string, error) {
	targets := []string{}
	if p == nil {
		return targets, nil
	}

	targetStr := fmt.Sprintf("%v", p)
	targetDirty := split(targetStr)

	for _, targetPath := range unique(targetDirty) {
		if strings.Contains(targetPath, "../") {
			return targets, fmt.Errorf("'../' not allowed in file path %q", targetPath)
		}

		targets = append(targets, targetPath)
	}

	return targets, nil
}
