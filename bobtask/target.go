package bobtask

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/buildinfostore"
)

// Target takes care of populating the targets members correctly.
// It returns a nil in case of a non existing target and a nil error.
func (t *Task) Target() (empty target.Target, _ error) {
	if t.target == nil {
		return empty, nil
	}

	hashIn, err := t.HashIn()
	if err != nil {
		if errors.Is(err, ErrHashInDoesNotExist) {
			return t.target.WithDir(t.dir), nil
		}
		return empty, err
	}

	buildInfo, err := t.ReadBuildinfo()
	if err != nil {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			return t.target.WithDir(t.dir), nil
		}
		return empty, err
	}

	expectedTargetHash, ok := buildInfo.Targets[hashIn]
	if !ok {
		return t.target.WithDir(t.dir), nil
	}

	return t.target.WithDir(t.dir).WithExpectedHash(expectedTargetHash), nil
}

func (t *Task) TargetExists() bool {
	return t.target != nil
}

// Clean the targets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean() error {
	if t.target != nil {
		for _, f := range t.target.Paths {
			if t.dir == "" {
				return fmt.Errorf("task dir not set")
			}
			p := filepath.Join(t.dir, f)
			if p == "/" {
				return fmt.Errorf("root cleanup is not allowed")
			}

			//fmt.Printf("Cleaning %s ", p)
			err := os.RemoveAll(p)
			if err != nil {
				//fmt.Printf("%s\n", aurora.Red("failed"))
				return err
			}
			//fmt.Printf("%s\n", aurora.Green("done"))
		}
	}

	return nil
}
