package playbook

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// TaskNeedsRebuild check if a tasks need a rebuild by looking at it's hash value
// and it's child tasks.
func (p *Playbook) TaskNeedsRebuild(taskname string, hashIn hash.In) (rebuildRequired bool, cause RebuildCause, err error) {
	ts, ok := p.Tasks[taskname]
	if !ok {
		return false, "", usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	task := ts.Task
	coloredName := task.ColoredName()

	// rebuild strategy set to `always`
	if task.Rebuild() == bobtask.RebuildAlways {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(rebuild set to always)", p.namePad, coloredName))
		return true, TaskForcedRebuild, nil
	}

	// child task changed
	if p.didChildTaskChange(task.Name(), p.namePad, coloredName) {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(dependecy changed)", p.namePad, coloredName))
		return true, DependencyChanged, nil
	}

	// current task changed // cache miss in buildinfostore
	rebuildRequired, err = task.NeedsRebuild()
	errz.Fatal(err)
	if rebuildRequired {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(input changed)", p.namePad, coloredName))
		return true, TaskInputChanged, nil
	}

	// check rebuild due to invalidated targets
	target, err := task.Target()
	if err != nil {
		return true, "", err
	}
	if target != nil {

		// TODO: simplify verify by check size + modification time of target.
		targetValid := target.VerifyShallow()
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\t 11111 TargetValid is %t", p.namePad, coloredName, targetValid))

		if !targetValid {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(invalid targets)", p.namePad, coloredName))
			return true, TargetInvalid, nil
		}

		// check if target exists in localstore
		hashIn, err := task.HashIn()
		errz.Fatal(err)
		if !task.ArtifactExists(hashIn) {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(target does not exist in localstore)", p.namePad, coloredName))
			return true, TargetNotInLocalStore, nil
		}

		//rebuildRequired = !targetValid

		// // Try to load a target from the store when a rebuild is required.
		// // If not assure the artifact exists in the store.
		// if rebuildRequired {
		// 	boblog.Log.V(2).Info(fmt.Sprintf("[task:%s] trying to get target from store", taskname))
		// 	ok, err := task.ArtifactUnpack(hashIn)
		// 	boblog.Log.Error(err, "Unable to get target from store")

		// 	if ok {
		// 		rebuildRequired = false
		// 	} else {
		// 		boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] failed to get target from store", taskname))
		// 	}
		// } else {
		// 	// Hint: Once there was a time when we created the target in the store
		// 	// in case no rebuild was required and the target doesn't exist.
		// 	// Though this should only be done after the target was really build..
		// 	// If loaded from a remote.. it anyway is synced through the local store.
		// 	if !task.ArtifactExists(hashIn) {
		// 		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(target not in local store)", p.namePad, coloredName))
		// 		return true, TargetInvalid, err
		// 	}
		// }

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
