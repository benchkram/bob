package bobtask

import (
	"fmt"
	"github.com/benchkram/bob/pkg/nix"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"

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

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	// skippedInputs is a lists of skipped input files
	skippedInputs []string

	Dependencies []string `yaml:"dependencies"`

	// Contains Dependencies + its bobfile Dependencies
	AllDependencies []string

	// Paths from /nix/store computed from all its dependencies
	StorePaths []string
	UseNix     bool
	Nixpkgs    string
}

type TargetEntry interface{}

func Make(opts ...TaskOption) Task {
	t := Task{
		inputs:               []string{},
		cmds:                 []string{},
		DependsOn:            []string{},
		Exports:              make(export.Map),
		env:                  []string{},
		rebuild:              RebuildOnChange,
		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
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
		t.target = target.New()
		t.target.Paths = targets
		t.target.Type = targetType
	}

	return nil
}

func parseTargetMap(tm map[string]interface{}) ([]string, target.TargetType, error) {

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

func (t *Task) PopulateStorePaths() error {
	var storePaths []string

	if t.UseNix && !nix.IsInstalled() {
		return fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl())
	}

	if t.UseNix && len(t.AllDependencies) > 0 {
		fmt.Println("Building nix dependencies...")
		storePathsFromPackages, err := nix.BuildPackages(nix.FilterPackageNames(t.AllDependencies), t.Nixpkgs)
		if err != nil {
			return err
		}
		fmt.Println("Building .nix files...")
		storePathsFromFiles, err := nix.BuildFiles(nix.FilterNixFiles(t.AllDependencies), t.Nixpkgs)
		if err != nil {
			return err
		}
		storePaths = append(storePathsFromPackages, storePathsFromFiles...)
	}

	t.StorePaths = storePaths
	return nil
}
