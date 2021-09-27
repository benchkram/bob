package bobfile

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Benchkram/errz"

	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/bobrun"
	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/bob/pkg/file"
)

var (
	defaultIgnores = []string{
		filepath.Join(global.BuildToolDir, "*"),
		filepath.Join(global.BobCacheDir, "*"),
	}
)

var (
	ErrNotImplemented         = fmt.Errorf("Not implemented")
	ErrBobfileNotFound        = fmt.Errorf("Could not find a Bobfile")
	ErrHashesFileDoesNotExist = fmt.Errorf("Hashes file does not exist")
	ErrTaskHashDoesNotExist   = fmt.Errorf("Task hash does not exist")
	ErrBobfileExists          = fmt.Errorf("Bobfile exists")
	ErrTaskDoesNotExist       = fmt.Errorf("Task does not exist")

	ErrInvalidRunType = fmt.Errorf("Invalid run type")
)

type Bobfile struct {
	Variables VariableMap

	Tasks bobtask.Map

	Runs bobrun.RunMap

	// Parent directory of the Bobfile.
	// Populated through BobfileRead().
	dir string
}

func NewBobfile() *Bobfile {
	b := &Bobfile{
		Variables: make(VariableMap),
		Tasks:     make(bobtask.Map),
		Runs:      make(bobrun.RunMap),
	}
	return b
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
	errz.Fatal(err)

	// Assure tasks are initialized with their defaults
	for key, task := range bobfile.Tasks {
		task.SetDir(bobfile.dir)
		task.SetName(key)
		task.InputDirty.Ignore = append(task.InputDirty.Ignore, defaultIgnores...)

		// Make sure a task is correctly initialised.
		// TODO: All unitialised must be initialised or get default values.
		// This mean switching to pointer types for most members.
		task.SetEnv([]string{})

		bobfile.Tasks[key] = task
	}

	// Assure runs are initialized with their defaults
	for key, run := range bobfile.Runs {
		run.SetDir(bobfile.dir)
		run.SetName(key)

		bobfile.Runs[key] = run
	}

	return bobfile, nil
}

// BobfileRead read from a bobfile.
// Calls sanitize on the result.
func BobfileRead(dir string) (_ *Bobfile, err error) {
	defer errz.Recover(&err)
	b, err := bobfileRead(dir)
	errz.Fatal(err)

	b.Tasks.Sanitize()

	return b, nil
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

	bobfile.Tasks[global.DefaultBuildTask] = bobtask.Task{
		InputDirty: bobtask.Input{
			Inputs: []string{"./main.go"},
			Ignore: []string{},
		},
		CmdDirty: "go build -o run",
		TargetDirty: target.T{
			Paths: []string{"run"},
			Type:  target.File,
		},
	}
	return bobfile.BobfileSave(dir)
}

func IsBobfile(file string) bool {
	return strings.Contains(filepath.Base(file), global.BobFileName)
}
