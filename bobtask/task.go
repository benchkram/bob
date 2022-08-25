package bobtask

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v3"

	"github.com/benchkram/bob/bobtask/export"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/pkg/usererror"
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
	// InputAdditionalIgnores is a list of ignores
	// usually the child targets.
	InputAdditionalIgnores []string
	// inputs is filtered by ignored & sanitized
	inputs []string

	CmdDirty string `yaml:"cmd"`
	// The cmds passed to os.Exec
	cmds []string

	// DependsOn are task which must succeed before this task
	// can run.
	DependsOn []string `yaml:"dependsOn"`

	// Target defines the output of a task.
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

	// project this tasks belongs to
	project string

	// dir is the working directory for this task
	dir string

	// env holds key=value pairs passed to the environment
	// when the task is executed.
	env []string

	// hashIn stores the `In` has for reuse
	hashIn *hash.In

	// local store for artifacts
	local store.Store

	// remote store for artifacts
	remote store.Store

	// buildInfoStore stores buildinfos.
	buildInfoStore buildinfostore.Store

	// color is used to color the task's name on the terminal
	color aurora.Color

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	// skippedInputs is a lists of skipped input files
	skippedInputs []string

	// DependenciesDirty read from the bobfile
	DependenciesDirty []string `yaml:"dependencies"`

	// dependencies contain the actual dependencies merged
	// with the global dependencies defined in the Bobfile
	// in the order which they need to be added to PATH
	dependencies []nix.Dependency

	// storePaths contain /nix/store/* paths
	// in the order which they need to be added to PATH
	storePaths []string
}

type TargetEntry interface{}

func Make(opts ...TaskOption) Task {
	t := Task{
		inputs:                 []string{},
		InputAdditionalIgnores: []string{},
		cmds:                   []string{},
		DependsOn:              []string{},
		Exports:                make(export.Map),
		env:                    []string{},
		rebuild:                RebuildOnChange,
		dockerRegistryClient:   dockermobyutil.NewRegistryClient(),
		dependencies:           []nix.Dependency{},
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

func (t *Task) Dependencies() []nix.Dependency {
	return t.dependencies
}
func (t *Task) SetDependencies(dependencies []nix.Dependency) {
	t.dependencies = dependencies
}

func (t *Task) SetStorePaths(storePaths []string) {
	t.storePaths = storePaths
}

func (t *Task) StorePaths() []string {
	return t.storePaths
}

// Project returns the projectname. In case of a non existing projectname the
// tasks local directory is returned.
func (t *Task) Project() string {
	if t.project == "" {
		return t.dir
	}
	return t.project
}

func (t *Task) SetProject(proj string) {
	t.project = proj
}

func (t *Task) SetDockerRegistryClient() {
	t.dockerRegistryClient = dockermobyutil.NewRegistryClient()
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

func (t *Task) WithRemotestore(s store.Store) *Task {
	t.remote = s
	return t
}

func (t *Task) WithBuildinfoStore(s buildinfostore.Store) *Task {
	t.buildInfoStore = s
	return t
}

const EnvironSeparator = "="

func (t *Task) AddEnvironmentVariable(key, value string) {
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

const (
	pathSelector  string = "path"
	imageSelector string = "image"
)

// parseTargets parses target definitions from yaml.
//
// Example yaml input:
//
// target: folder/
//
// target: |-
//	folder/
//	folder1/folder/file
//
// target:
//   path: |-
//		folder/
//		folder1/folder/file
//
// target:
//	image: docker-image-name
//
// target:
//   image: |-
//		docker-image-name
//		docker-image2-name
//
func (t *Task) parseTargets() error {
	targetType := target.DefaultType

	var targets []string
	var err error

	switch td := t.TargetDirty.(type) {
	case string:
		targets, err = parseTargetPath(td)
	case map[string]interface{}:
		targets, targetType, err = parseTargetMap(td)
		if err != nil {
			err = usererror.Wrapm(err, fmt.Sprintf("[task:%s]", t.name))
		}
	}

	if err != nil {
		return err
	}

	if len(targets) > 0 {
		t.target = target.New(
			target.WithType(targetType),
			target.WithTargetPaths(targets),
		)
	}

	return nil
}

func parseTargetMap(tm map[string]interface{}) ([]string, target.Type, error) {

	// check first if both directives are selected
	if keyExists(tm, pathSelector) && keyExists(tm, imageSelector) {
		return nil, target.DefaultType, ErrAmbigousTargetDefinition
	}

	paths, ok := tm[pathSelector]
	if ok {
		targets, err := parseTargetPath(paths.(string))
		if err != nil {
			return nil, target.DefaultType, err
		}

		return targets, target.Path, nil
	}

	images, ok := tm[imageSelector]
	if !ok {
		return nil, target.DefaultType, ErrInvalidTargetDefinition
	}

	return parseTargetImage(images.(string)), target.Docker, nil
}

func parseTargetPath(p string) ([]string, error) {
	targets := []string{}
	if p == "" {
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

func parseTargetImage(p string) []string {
	if p == "" {
		return []string{}
	}

	targetStr := fmt.Sprintf("%v", p)
	targetDirty := split(targetStr)

	return unique(targetDirty)
}

func keyExists(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func (t *Task) UnmarshalYAML(value *yaml.Node) (err error) {
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
	type TmpTask Task
	var tmpTask TmpTask

	err = value.Decode(&tmpTask)
	errz.Fatal(err)

	tmpTask.DependsOn = dependsOn

	*t = Task(tmpTask)

	return nil
}
