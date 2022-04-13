package bobfile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/store"
	storeclient "github.com/benchkram/bob/pkg/store-client"
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
	ErrSelfReference          = fmt.Errorf("self reference")

	ErrInvalidRunType = fmt.Errorf("Invalid run type")
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

	// Parent directory of the Bobfile.
	// Populated through BobfileRead().
	dir string

	bobfiles []*Bobfile

	remotestore store.Store
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
func (b *Bobfile) Remotestore() store.Store {
	return b.remotestore
}

// bobfileRead reads a bobfile and intializes private fields.
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

		bobfile.BTasks[key] = task
	}

	// Assure runs are initialized with their defaults
	for key, run := range bobfile.RTasks {
		run.SetDir(bobfile.dir)
		run.SetName(key)

		bobfile.RTasks[key] = run
	}

	// Initialize remote store in case of a valid remote url /  projectname.
	if bobfile.Project != "" {
		projectname, err := project.Parse(bobfile.Project)
		if err != nil {
			return nil, err
		}

		switch projectname.Type() {
		case project.Local:
			// Do nothing
		case project.Remote:
			// Initialize remote store
			url, err := projectname.Remote()
			if err != nil {
				return nil, err
			}
			println("using remote store")
			println(url.String())
			bobfile.remotestore = newRemotestore(url)
		}
	} else {
		bobfile.Project = bobfile.dir
	}

	return bobfile, nil
}

func newRemotestore(endpoint *url.URL) (s store.Store) {
	const sep = "/"

	parts := strings.Split(endpoint.Path, sep)

	username := parts[0]
	project := strings.Join(parts[1:], sep)
	s = remotestore.New(
		username,
		project,
		remotestore.WithClient(
			storeclient.New("http://"+endpoint.Host),
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

	return b, b.BTasks.Sanitize()
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
	_, err = project.Parse(b.Project)
	if err != nil {
		return err
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

func (b *Bobfile) BobfileSave(dir string) (err error) {
	defer errz.Recover(&err)

	buf := bytes.NewBuffer([]byte{})

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	defer encoder.Close()

	err = encoder.Encode(b)
	errz.Fatal(err)

	return ioutil.WriteFile(filepath.Join(dir, global.BobFileName), buf.Bytes(), 0664)
}

func (b *Bobfile) Dir() string {
	return b.dir
}

func CreateDummyBobfile(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(global.BobFileName) && !overwrite {
		return ErrBobfileExists
	}

	bobfile := NewBobfile()

	bobfile.BTasks[global.DefaultBuildTask] = bobtask.Task{
		InputDirty:  "./main.go",
		CmdDirty:    "go build -o run",
		TargetDirty: "run",
	}
	return bobfile.BobfileSave(dir)
}

func IsBobfile(file string) bool {
	return strings.Contains(filepath.Base(file), global.BobFileName)
}
