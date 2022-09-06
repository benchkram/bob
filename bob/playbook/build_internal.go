package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// didWriteBuildOutput assures that a new line is added
// before writing state or logs of a task to stdout.
var didWriteBuildOutput bool

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task *bobtask.Task) (err error) {
	defer errz.Recover(&err)

	// A task id flagged succesful before
	var taskSuccessFul bool
	var taskErr error
	defer func() {
		if !taskSuccessFul {
			errr := p.TaskFailed(task.Name(), taskErr)
			if errr != nil {
				boblog.Log.Error(errr, "Setting the task state to failed, failed.")
			}
		}
	}()

	coloredName := task.ColoredName()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, StateCanceled))
				_ = p.TaskCanceled(task.Name())
			}
		}
	}()

	rebuildRequired, rebuildCause, err := p.TaskNeedsRebuild(task.Name())
	errz.Fatal(err)
	boblog.Log.V(2).Info(fmt.Sprintf("TaskNeedsRebuild [rebuildRequired: %t] [cause:%s]", rebuildRequired, rebuildCause))

	// task might need a rebuild due to a input change.
	// Could still be possible to load the targets from the artifact store.
	// If a task needs a rebuild due to a dependency change => rebuild.
	if rebuildRequired {
		switch rebuildCause {
		case TaskInputChanged:
			hashIn, err := task.HashIn()
			errz.Fatal(err)
			success, err := task.ArtifactUnpack(hashIn)
			errz.Fatal(err)
			if success {
				rebuildRequired = false

				// In case a artifact was synced from the remote store no buildinfo exists...
				// To avaoid subsequent artifact extraction the Buildinfo is created after
				// unpacking the artifact.
				buildInfo, err := p.computeBuildinfo(task.Name())
				errz.Fatal(err)
				err = p.storeBuildInfo(task.Name(), buildInfo)
				errz.Fatal(err)
			}
		case TargetInvalid:
			boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s, unpacking artifact", p.namePad, coloredName, rebuildCause))
			hashIn, err := task.HashIn()
			errz.Fatal(err)
			success, err := task.ArtifactUnpack(hashIn)
			errz.Fatal(err)
			if success {
				rebuildRequired = false
			}
		case TargetNotInLocalStore:
		case TaskForcedRebuild:
		case DependencyChanged:
		default:
		}
	}

	if !rebuildRequired {
		status := StateNoRebuildRequired
		boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, status.Short()))
		taskSuccessFul = true
		return p.TaskNoRebuildRequired(task.Name())
	}

	if !didWriteBuildOutput {
		boblog.Log.V(1).Info("")
		didWriteBuildOutput = true
	}
	err = task.Clean()
	errz.Fatal(err)

	err = task.Run(ctx, p.namePad)
	if err != nil {
		taskSuccessFul = false
		taskErr = err
	}
	errz.Fatal(err)

	// TODO: think about if this is correctlty placed.
	// Could also be done after the task completed correctly?
	// Ass target verification is no don einside TaskCompleted()
	taskSuccessFul = true

	// err = task.VerifyAfter()
	// errz.Fatal(err)

	// target, err := task.Target()
	// if err != nil {
	// 	errz.Fatal(err)
	// }

	// Check targets are created correctly.
	// On success the target hash is computed
	// inside TaskCompleted().
	// if target != nil {
	// 	if !target.Exists() {
	// 		boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s\t(invalid targets)", p.namePad, coloredName, StateFailed))
	// 		err = p.TaskFailed(task.Name(), fmt.Errorf("targets not created"))
	// 		if err != nil {
	// 			if errors.Is(err, ErrFailed) {
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// TODO: check if the error handling works as intended, in case of a not created target inside `cmd:`
	// the user should get a correct error message that the cmd does not create the target as expected.
	err = p.TaskCompleted(task.Name())
	if err != nil {
		if err != nil {
			if errors.Is(err, ErrFailed) {
				return err
			}
		}
	}
	errz.Fatal(err)

	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)

	state := taskStatus.State()
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, "..."+state.Short()))

	return nil
}

const maxSkippedInputs = 5
