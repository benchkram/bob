package playbook

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// RebuildInfo contains information about a task rebuild: if it's required and the cause for it
type RebuildInfo struct {
	// IsRequired tells if the task requires rebuild again
	IsRequired bool
	// Cause tells why the rebuild is required
	Cause RebuildCause
	// VerifyResult is the result of target filesystem verification
	VerifyResult target.VerifyResult
}

// TaskNeedsRebuild check if a tasks need a rebuild by looking at its hash value
// and its child tasks.
func (p *Playbook) TaskNeedsRebuild(taskName string) (rebuildInfo RebuildInfo, err error) {
	ts, ok := p.Tasks[taskName]
	if !ok {
		return RebuildInfo{}, usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskName))
	}
	task := ts.Task
	coloredName := task.ColoredName()

	// Rebuild strategy set to `always`
	if task.Rebuild() == bobtask.RebuildAlways {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(rebuild set to always)", p.namePad, coloredName))

		// for forced rebuild all task targets are marked as invalid
		invalidFiles := make(map[string][]target.Reason)
		if task.TargetExists() {
			t, err := ts.Target()
			errz.Fatal(err)
			invalidFiles = t.AsInvalidFiles(target.ReasonForcedByNoCache)
		}

		return RebuildInfo{IsRequired: true, Cause: TaskForcedRebuild, VerifyResult: target.VerifyResult{
			TargetIsValid: len(invalidFiles) > 0,
			InvalidFiles:  invalidFiles,
		}}, nil
	}

	// Did a child task change?
	if p.didChildTaskChange(task.Name()) {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(dependecy changed)", p.namePad, coloredName))
		return RebuildInfo{IsRequired: true, Cause: DependencyChanged}, nil
	}

	// Did the current task change?
	// Indicating a cache miss in buildinfostore.
	rebuildRequired, err := task.DidTaskChange()
	errz.Fatal(err)
	if rebuildRequired {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(input changed)", p.namePad, coloredName))

		invalidFiles := make(map[string][]target.Reason)
		if task.TargetExists() {
			t, err := ts.Target()
			errz.Fatal(err)
			invalidFiles = t.AsInvalidFiles(target.ReasonMissing)
		}

		return RebuildInfo{IsRequired: true, Cause: InputNotFoundInBuildInfo, VerifyResult: target.VerifyResult{
			TargetIsValid: len(invalidFiles) > 0,
			InvalidFiles:  invalidFiles,
		}}, nil
	}

	// Check rebuild due to invalidated targets
	target, err := task.Target()
	if err != nil {
		return RebuildInfo{IsRequired: true, Cause: ""}, nil
	}
	if target != nil {
		verifyResult := target.VerifyShallow()
		if !verifyResult.TargetIsValid {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(invalid targets)", p.namePad, coloredName))
			return RebuildInfo{IsRequired: true, Cause: TargetInvalid, VerifyResult: verifyResult}, nil
		}

		// Check if target exists in local store
		hashIn, err := task.HashIn()
		errz.Fatal(err)
		if !task.ArtifactExists(hashIn) {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(target does not exist in localstore)", p.namePad, coloredName))
			return RebuildInfo{IsRequired: true, Cause: TargetNotInLocalStore, VerifyResult: verifyResult}, nil
		}
	}

	return RebuildInfo{IsRequired: false}, err
}

// didChildTaskChange iterates through all child tasks to verify if any of them changed.
func (p *Playbook) didChildTaskChange(taskName string) bool {
	var Done = fmt.Errorf("done")
	err := p.Tasks.walk(taskName, func(tn string, t *Status, err error) error {
		if err != nil {
			return err
		}

		// Ignore the task itself
		if taskName == tn {
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
