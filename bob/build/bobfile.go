package build

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/pkg/multilinecmd"
)

const (
	BobFileName = "bob.yaml"

	DefaultBuildTask = "build"
)

var (
	defaultIgnores = []string{
		filepath.Join(".bob", "*"), // TODO: Use bob.BuildToolDir
		filepath.Join(BobCacheDir, "*"),
	}
)

var (
	ErrBobfileNotFound        = fmt.Errorf("Could not find a Bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("Hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("Task hash does not exist")
	ErrBobfileExists          = fmt.Errorf("Bobfile exists")
	ErrTaskDoesNotExist       = fmt.Errorf("Task does not exist")
)

type Bobfile struct {
	Variables VariableMap

	Tasks TaskMap

	// Parent directory of the Bobfile.
	// Populated through BobfileRead().
	dir string
}

func NewBobfile() *Bobfile {
	b := &Bobfile{
		Variables: make(VariableMap),
		Tasks:     make(TaskMap),
	}
	return b
}

// bobfileRead reads a bobfile and intializes private fields.
func bobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)

	bobfilePath := filepath.Join(dir, BobFileName)

	if !file.Exists(bobfilePath) {
		return nil, ErrBobfileNotFound
	}
	bin, err := ioutil.ReadFile(bobfilePath)
	errz.Fatal(err, "Failed to read config file")

	bobfile := &Bobfile{
		dir: dir,
	}
	err = yaml.Unmarshal(bin, bobfile)
	errz.Fatal(err)

	// Inject dir & taskname into each task.
	for key, task := range bobfile.Tasks {
		task.dir = bobfile.dir
		task.name = key
		task.InputDirty.Ignore = append(task.InputDirty.Ignore, defaultIgnores...)

		// Make sure a task is correctly initialised.
		// TODO: All unitialised must be initialised or get default values.
		// This mean switching to pointer types for most members.
		task.env = []string{}

		bobfile.Tasks[key] = task
	}

	return bobfile, nil
}

// BobfileRead read from a bobfile.
// Calls sanitize on the result.
func BobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)
	b, err := bobfileRead(dir)
	errz.Fatal(err)

	// sanitize
	for key, task := range b.Tasks {
		inputs, err := task.filteredInputs()
		errz.Fatal(err)
		task.inputs = inputs

		sanitizedExports, err := task.sanitizeExports(task.Exports)
		errz.Fatal(err)
		task.Exports = sanitizedExports

		sanitizedTargetPaths, err := task.sanitizeTarget(task.TargetDirty.Paths)
		errz.Fatal(err)
		task.target.Paths = sanitizedTargetPaths
		task.target.Type = task.TargetDirty.Type

		task.cmds = multilinecmd.Split(task.CmdDirty)

		b.Tasks[key] = task
	}

	return b, nil
}

func (b *Bobfile) BobfileSave(dir string) (err error) {
	defer errz.Recover(&err)
	bin, err := yaml.Marshal(b)
	errz.Fatal(err)

	return ioutil.WriteFile(filepath.Join(dir, BobFileName), bin, 0664)
}

func (b *Bobfile) Dir() string {
	return b.dir
}

func CreateDummyBobfile(dir string, overwrite bool) (err error) {
	// Prevent accidential bobfile override
	if file.Exists(BobFileName) && !overwrite {
		return ErrBobfileExists
	}

	bobfile := NewBobfile()

	bobfile.Tasks[DefaultBuildTask] = Task{
		InputDirty: Input{
			Inputs: []string{"./main.go"},
			Ignore: []string{},
		},
		CmdDirty: "go build -o run",
		TargetDirty: Target{
			Paths: []string{"run"},
			Type:  File,
		},
	}
	return bobfile.BobfileSave(dir)
}

func IsBobfile(file string) bool {
	return strings.Contains(filepath.Base(file), BobFileName)
}
