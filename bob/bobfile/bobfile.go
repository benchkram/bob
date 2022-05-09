package bobfile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/nix"

	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/usererror"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile/project"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/bobrun"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/file"
)

var (
	defaultIgnores = fmt.Sprintf("!%s\n!%s",
		global.BobWorkspaceFile,
		filepath.Join(global.BobCacheDir, "*"),
	)
)

var (
	ErrNotImplemented         = fmt.Errorf("Not implemented")
	ErrBobfileNotFound        = fmt.Errorf("Could not find a Bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("Hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("Task hash does not exist")
	ErrBobfileExists          = fmt.Errorf("Bobfile exists")
	ErrTaskDoesNotExist       = fmt.Errorf("Task does not exist")
	ErrDuplicateTaskName      = fmt.Errorf("duplicate task name")
	ErrInvalidProjectName     = fmt.Errorf("invalid project name")
	ErrSelfReference          = fmt.Errorf("self reference")

	ErrInvalidRunType = fmt.Errorf("Invalid run type")

	ProjectNameFormatHint = "project name should be in the form 'project' or 'registry.com/user/project'"
)

type Bobfile struct {
	// Version is optional, and can be used to
	Version string `yaml:"version,omitempty"`

	// Project uniquely identifies the current project (optional). If supplied,
	// aggregation makes sure the project does not depend on another instance
	// of itself. If not provided, then the project name is set after the path
	// of its bobfile.
	Project string `yaml:"project,omitempty"`

	Variables VariableMap

	// BTasks build tasks
	BTasks bobtask.Map `yaml:"build"`
	// RTasks run tasks
	RTasks bobrun.RunMap `yaml:"run"`

	Dependencies []string `yaml:"dependencies"`

	// UseNix is a flag to indicate if nix is used
	// by any task inside a bobfile
	UseNix bool `yaml:"use-nix"`
	// Nixpkgs specifies an optional nixpkgs source.
	Nixpkgs string `yaml:"nixpkgs"`

	// Parent directory of the Bobfile.
	// Populated through BobfileRead().
	dir string

	bobfiles []*Bobfile
}

func NewBobfile() *Bobfile {
	b := &Bobfile{
		Variables: make(VariableMap),
		BTasks:    make(bobtask.Map),
		RTasks:    make(bobrun.RunMap),
	}
	return b
}

func (b *Bobfile) SetBobfiles(bobs []*Bobfile) {
	b.bobfiles = bobs
}

func (b *Bobfile) Bobfiles() []*Bobfile {
	return b.bobfiles
}

// bobfileRead reads a bobfile and initializes private fields.
func bobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	bobfilePath := filepath.Join(dir, global.BobFileName)

	if !file.Exists(bobfilePath) {
		return nil, ErrBobfileNotFound
	}
	bin, err := ioutil.ReadFile(bobfilePath)
	errz.Fatal(err, "Failed to read config file")

	bobfile := &Bobfile{
		dir: dir,
	}

	err = yaml.Unmarshal(bin, bobfile)
	if err != nil {
		return nil, usererror.Wrapm(err, "YAML unmarshal failed")
	}

	if bobfile.Variables == nil {
		bobfile.Variables = VariableMap{}
	}

	if bobfile.BTasks == nil {
		bobfile.BTasks = bobtask.Map{}
	}

	if bobfile.RTasks == nil {
		bobfile.RTasks = bobrun.RunMap{}
	}

	// Assure tasks are initialized with their defaults
	for key, task := range bobfile.BTasks {
		task.SetDir(bobfile.dir)
		task.SetName(key)

		task.InputDirty = fmt.Sprintf("%s\n%s", task.InputDirty, defaultIgnores)

		// Make sure a task is correctly initialised.
		// TODO: All unitialised must be initialised or get default values.
		// This means switching to pointer types for most members.
		task.SetEnv([]string{})
		task.SetRebuildStrategy(bobtask.RebuildOnChange)

		// initialize docker registry for task
		task.SetDockerRegistryClient()
		task.SetDependencies(initializeDependencies(dir, task, bobfile))
		task.SetUseNix(bobfile.UseNix)

		bobfile.BTasks[key] = task
	}

	// Assure runs are initialized with their defaults
	for key, run := range bobfile.RTasks {
		run.SetDir(bobfile.dir)
		run.SetName(key)

		bobfile.RTasks[key] = run
	}

	return bobfile, nil
}

// initializeDependencies gathers all dependencies for a task(task level and bobfile level)
// and initialize them with bobfile dir and corresponding nixpkgs used
func initializeDependencies(dir string, task bobtask.Task, bobfile *Bobfile) []bobtask.Dependency {
	dependencies := sliceutil.Unique(append(task.DependenciesDirty, bobfile.Dependencies...))
	dependencies = nix.AddDir(dir, dependencies)

	taskDeps := make([]bobtask.Dependency, 0)
	for _, v := range dependencies {
		taskDeps = append(taskDeps, bobtask.Dependency{
			Name:    v,
			Nixpkgs: bobfile.Nixpkgs,
		})
	}

	return bobtask.UniqueDeps(taskDeps)
}

// BobfileRead read from a bobfile.
// Calls sanitize on the result.
func BobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	b, err := bobfileRead(dir)
	errz.Fatal(err)

	err = b.Validate()
	errz.Fatal(err)

	return b, b.BTasks.Sanitize()
}

// BobfileReadPlain reads a bobfile.
// For performance reasons sanitize is not called.
func BobfileReadPlain(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	b, err := bobfileRead(dir)
	errz.Fatal(err)

	err = b.Validate()
	errz.Fatal(err)

	return b, nil
}

// Validate makes sure no task depends on itself (self-reference) or has the same name as another task
func (b *Bobfile) Validate() (err error) {
	if b.Version != "" {
		_, err = version.NewVersion(b.Version)
		if err != nil {
			return fmt.Errorf("invalid version '%s' (%s)", b.Version, b.Dir())
		}
	}

	// validate project name if set
	if b.Project != "" {
		if !project.RestrictedProjectNamePattern.MatchString(b.Project) {
			return usererror.Wrap(errors.WithMessage(ErrInvalidProjectName, ProjectNameFormatHint))
		}

		// test for double slash (do not allow prepended schema)
		if project.ProjectNameDoubleSlashPattern.MatchString(b.Project) {
			return usererror.Wrap(errors.WithMessage(ErrInvalidProjectName, ProjectNameFormatHint))
		}
	}

	// use for duplicate names validation
	names := map[string]bool{}

	for name, task := range b.BTasks {
		// validate no duplicate name
		if names[name] {
			return errors.WithMessage(ErrDuplicateTaskName, name)
		}

		names[name] = true

		// validate no self-reference
		for _, dep := range task.DependsOn {
			if name == dep {
				return errors.WithMessage(ErrSelfReference, name)
			}
		}
	}

	for name, run := range b.RTasks {
		// validate no duplicate name
		if names[name] {
			return errors.WithMessage(ErrDuplicateTaskName, name)
		}

		names[name] = true

		// validate no self-reference
		for _, dep := range run.DependsOn {
			if name == dep {
				return errors.WithMessage(ErrSelfReference, name)
			}
		}
	}

	return nil
}

func (b *Bobfile) BobfileSave(dir, name string) (err error) {
	defer errz.Recover(&err)

	buf := bytes.NewBuffer([]byte{})

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	defer encoder.Close()

	err = encoder.Encode(b)
	errz.Fatal(err)

	return ioutil.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0664)
}

func (b *Bobfile) Dir() string {
	return b.dir
}

func CreateDummyBobfile(dir string, overwrite bool) (err error) {
	// Prevent accidental bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return ErrBobfileExists
	}

	bobfile := NewBobfile()

	bobfile.BTasks[global.DefaultBuildTask] = bobtask.Task{
		InputDirty:  "./main.go",
		CmdDirty:    "go build -o run",
		TargetDirty: "run",
	}
	return bobfile.BobfileSave(dir, global.BobFileName)
}

func IsBobfile(file string) bool {
	return strings.Contains(filepath.Base(file), global.BobFileName)
}
