package build

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/mholt/archiver/v3"

	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/errz"
)

type Target struct {
	Paths []string
	Type  TargetType
}

func MakeTarget() Target {
	return Target{
		Paths: []string{},
		Type:  File,
	}
}

type TargetType string

const (
	File   TargetType = "file"
	Docker TargetType = "docker"
)

func (t *Task) Target() (target *Target) {
	return &t.target
}

// Clean the tragets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean() error {
	for _, f := range t.target.Paths {
		if t.dir == "" {
			panic("Task dir not set")
		}
		p := filepath.Join(t.dir, f)
		if p == "/" {
			panic("Root cleanup is permitted")
		}

		fmt.Printf("Cleaning %s ", p)
		err := os.RemoveAll(p)
		if err != nil {
			fmt.Printf("%s\n", aurora.Red("failed"))
			return err
		}
		fmt.Printf("%s\n", aurora.Green("done"))
	}
	return nil
}

// DidSucceede returns false if any of the targets is missing.
func (t *Task) DidSucceede() (_ bool, failedTargets []string) {
	failedTargets = []string{}
	for _, f := range t.target.Paths {
		if !file.Exists(filepath.Join(t.dir)) {
			failedTargets = append(failedTargets, f)
		}
	}
	return len(failedTargets) == 0, failedTargets
}

// Pack creates a archive for a target
func (t *Task) Pack(hash string) (err error) {
	defer errz.Recover(&err)

	paths := []string{}
	for _, path := range t.target.Paths {
		paths = append(paths, filepath.Join(t.dir, path))
	}

	archive := filepath.Join(t.dir, BobCacheDir, hash+".tar.br")
	err = os.RemoveAll(archive)
	errz.Fatal(err)
	err = archiver.Archive(paths, archive)
	errz.Fatal(err)

	return nil
}

// Unpack targets from a given archive
func (t *Task) Unpack() error {
	return nil
}
