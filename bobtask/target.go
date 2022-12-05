package bobtask

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boblog"
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
			return t.target, t.target.Resolve()
		}
		if errors.Is(err, buildinfostore.ErrBuildInfoInvalid) {
			boblog.Log.V(5).Info(fmt.Sprintf("Invalid buildinfo found for [taskname:%s], ignoring", t.name))
			return t.target, t.target.Resolve()
		}
		return empty, err
	}

	// This indicates the previous build did not contain any targets and therefore it
	// can't  be  compared against.
	// FIXME: Is this necessary? Seems like it rather happens during development.
	if len(buildInfo.Target.Filesystem.Files) == 0 && len(buildInfo.Target.Docker) == 0 {
		return t.target, t.target.Resolve()
	}

	tt := t.target.WithExpected(&buildInfo.Target)
	return tt, tt.Resolve()
}

func (t *Task) TargetExists() bool {
	return t.target != nil
}
