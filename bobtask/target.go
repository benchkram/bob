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

	// ReadBuildInfo is dependent on the inputHash of the task.
	// For this reason we cannot read build info on target creation,
	// as this happens right after parsing the config.
	// Computing the input must be avoided till the task is actually
	// passed to the worker.

	buildInfo, err := t.ReadBuildInfo()
	if err != nil {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			tt := t.target.WithDir(t.dir)
			return tt, tt.Resolve()
		}
		return empty, err
	}

	// This indicates the previous build did not contain any targets.
	// TODO: Is this necessary? Document why this is necessary.
	if len(buildInfo.Target.Filesystem.Files) == 0 && len(buildInfo.Target.Docker) == 0 {
		tt := t.target.WithDir(t.dir)
		return tt, tt.Resolve()
	}

	tt := t.target.WithDir(t.dir).WithExpected(&buildInfo.Target)
	return tt, tt.Resolve()
}

func (t *Task) TargetExists() bool {
	return t.target != nil
}

// Clean the targets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean() error {
	if t.target != nil {
		for _, f := range t.target.FilesystemEntriesRawPlain() {
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
