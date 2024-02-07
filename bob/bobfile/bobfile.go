package bobfile

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/nix"
	storeclient "github.com/benchkram/bob/pkg/store-client"

	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/pkg/store/remotestore"
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
	ErrNotImplemented         = fmt.Errorf("Not implemented")
	ErrBobfileNotFound        = fmt.Errorf("Could not find a bob.yaml")
	ErrHashesFileDoesNotExist = fmt.Errorf("Hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("Task hash does not exist")
	ErrBobfileExists          = fmt.Errorf("Bobfile exists")
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

	Imports []string `yaml:"import,omitempty"`

	// Variables is a map of variables that can be used in the tasks.
	Variables VariableMap

	// BTasks build tasks
	BTasks bobtask.Map `yaml:"build"`
	// RTasks run tasks
	RTasks bobrun.RunMap `yaml:"run"`

	// Dependencies are nix packages used on a global scope.
	// Mutually exclusive with Shell. ??Overwrites task based dependencies.??
	Dependencies []string `yaml:"dependencies"`

	// Shell specifies a shell.nix file as usually used by nix-shell.
	// This is mutualy exclusive with Dependencies.
	ShellDotNix string `yaml:"shell"`

	// Nixpkgs specifies an optional nixpkgs source.
	Nixpkgs string `yaml:"nixpkgs"`

	// Parent directory of the Bobfile.
	// Populated through BobfileRead().
	dir string

	bobfiles []*Bobfile

	RemoteStoreHost string
	remotestore     store.Store
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

func (b *Bobfile) SetRemotestore(remote store.Store) {
	b.remotestore = remote
}

func (b *Bobfile) Remotestore() store.Store {
	return b.remotestore
}

// bobfileRead reads a bobfile and initializes private fields.
func bobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	bobfilePath := filepath.Join(dir, global.BobFileName)

	if !file.Exists(bobfilePath) {
		return nil, usererror.Wrap(ErrBobfileNotFound)
	}
	bin, err := os.ReadFile(bobfilePath)
	errz.Fatal(err)

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
		task.InputAdditionalIgnores = []string{}

		// Make sure a task is correctly initialised.
		// TODO: All unitialised must be initialised or get default values.
		// This means switching to pointer types for most members.
		task.SetEnv([]string{})
		task.SetRebuildStrategy(bobtask.RebuildOnChange)

		// initialize docker registry for task
		task.SetDependencies(initializeDependencies(dir, task.DependenciesDirty, bobfile))

		bobfile.BTasks[key] = task
	}

	// Assure runs are initialized with their defaults
	for key, run := range bobfile.RTasks {
		run.SetDir(bobfile.dir)
		run.SetName(key)
		run.SetEnv([]string{})

		run.SetDependencies(initializeDependencies(dir, run.DependenciesDirty, bobfile))

		bobfile.RTasks[key] = run
	}

	// // Initialize remote store in case of a valid remote url /  projectname.
	// if bobfile.Project != "" {
	//	projectname, err := project.Parse(bobfile.Project)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	switch projectname.Type() {
	//	case project.Local:
	//		// Do nothing
	//	case project.Remote:
	//		// Initialize remote store
	//		url, err := projectname.Remote()
	//		if err != nil {
	//			return nil, err
	//		}
	//
	//		boblog.Log.V(1).Info(fmt.Sprintf("Using remote store: %s", url.String()))
	//
	//		bobfile.remotestore = NewRemotestore(url)
	//	}
	// } else {
	//	bobfile.Project = bobfile.dir
	// }

	return bobfile, nil
}

// initializeDependencies gathers all dependencies for a task(task level and bobfile level)
// and initialize them with bobfile dir and corresponding nixpkgs used
func initializeDependencies(dir string, taskDependencies []string, bobfile *Bobfile) []nix.Dependency {
	dependencies := sliceutil.Unique(append(taskDependencies, bobfile.Dependencies...))
	dependencies = nix.AddDir(dir, dependencies)

	taskDeps := make([]nix.Dependency, 0)
	for _, v := range dependencies {
		taskDeps = append(taskDeps, nix.Dependency{
			Name:    v,
			Nixpkgs: bobfile.Nixpkgs,
		})
	}

	return nix.UniqueDeps(taskDeps)
}

func NewRemotestore(endpoint *url.URL, allowInsecure bool, token string) (s store.Store) {
	const sep = "/"

	parts := strings.Split(strings.TrimLeft(endpoint.Path, sep), sep)

	username := parts[0]
	proj := strings.Join(parts[1:], sep)

	protocol := "https://"
	if allowInsecure {
		protocol = "http://"
	}

	s = remotestore.New(
		username,
		proj,

		remotestore.WithClient(
			storeclient.New(protocol+endpoint.Host, token),
		),
	)
	return s
}

// BobfileRead read from a bobfile.
// Calls sanitize on the result.
func BobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	b, err := bobfileRead(dir)
	errz.Fatal(err)

	err = b.Validate()
	errz.Fatal(err)

	err = b.BTasks.Sanitize()
	errz.Fatal(err)

	err = b.RTasks.Sanitize()
	errz.Fatal(err)

	return b, nil
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

	return os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0664)
}

func (b *Bobfile) Dir() string {
	return b.dir
}

// Vars returns the bobfile variables in the form "key=value"
// based on its Variables
func (b *Bobfile) Vars() []string {
	var env []string
	for key, value := range b.Variables {
		env = append(env, strings.Join([]string{key, value}, "="))
	}
	return env
}
