package playbook

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// TaskNeedsRebuild check if a tasks need a rebuild by looking at its hash value
// and its child tasks.
func (p *Playbook) TaskNeedsRebuild(taskname string) (rebuildRequired bool, cause RebuildCause, err error) {
	ts, ok := p.Tasks[taskname]
	if !ok {
		return false, "", usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	task := ts.Task
	coloredName := task.ColoredName()

	// Rebuild strategy set to `always`
	if task.Rebuild() == bobtask.RebuildAlways {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(rebuild set to always)", p.namePad, coloredName))
		return true, TaskForcedRebuild, nil
	}

	// Did a child task change?
	if p.didChildTaskChange(task.Name(), p.namePad, coloredName) {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(dependecy changed)", p.namePad, coloredName))
		return true, DependencyChanged, nil
	}

	// Did the current task change?
	// Indicating a cache miss in buildinfostore.
	rebuildRequired, err = task.DidTaskChange()
	errz.Fatal(err)
	if rebuildRequired {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(input changed)", p.namePad, coloredName))
		return true, InputNotFoundInBuildInfo, nil
	}

	// Check rebuild due to invalidated targets
	target, err := task.Target()
	if err != nil {
		return true, "", err
	}
	if target != nil {
		targetValid := target.VerifyShallow()
		if !targetValid {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(invalid targets)", p.namePad, coloredName))
			return true, TargetInvalid, nil
		}

		// Check if target exists in localstore
		hashIn, err := task.HashIn()
		errz.Fatal(err)
		if !task.ArtifactExists(hashIn) {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(target does not exist in localstore)", p.namePad, coloredName))
			return true, TargetNotInLocalStore, nil
		}
	}

	return false, "", err
}

// didChildTaskChange iterates through all child tasks to verify if any of them changed.
func (p *Playbook) didChildTaskChange(taskname string, namePad int, coloredName string) bool {
	var Done = fmt.Errorf("done")
	err := p.Tasks.walk(taskname, func(tn string, t *Status, err error) error {
		if err != nil {
			return err
		}

		// Ignore the task itself
		if taskname == tn {
			return nil
		}

		// Check if child task changed
		if t.State() != StateNoRebuildRequired {
			return Done
		}

		return nil
	})

	return errors.Is(err, Done)
}
