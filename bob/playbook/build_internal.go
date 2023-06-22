package playbook

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/processed"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task *bobtask.Task) (pt *processed.Task, err error) {
	defer errz.Recover(&err)

	// if `pt` is `nil` errz.Fatal()
	// returns a nil task which could lead
	// to memory leaks down the line.
	pt = &processed.Task{Task: task}

	// A task is flagged successful before
	var taskSuccessFul bool
	var taskErr error
	defer func() {
		if !taskSuccessFul {
			errr := p.TaskFailed(task.TaskID, taskErr)
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
				_ = p.TaskCanceled(task.TaskID)
			}
		}
	}()

	rebuild, err := p.TaskNeedsRebuild(task.TaskID)
	errz.Fatal(err)
	boblog.Log.V(2).Info(fmt.Sprintf("TaskNeedsRebuild [rebuildRequired: %t] [cause:%s]", rebuild.IsRequired, rebuild.Cause))

	// Task might need a rebuild due to an input change.
	// Could still be possible to load the targets from the artifact store.
	// If a task needs a rebuild due to a dependency change => rebuild.
	if rebuild.IsRequired {
		switch rebuild.Cause {
		case InputNotFoundInBuildInfo:
			hashIn, err := task.HashIn()
			errz.Fatal(err)

			// pull artifact if it exists on the remote. if exists locally will use that one
			err = p.pullArtifact(ctx, hashIn, task, false)
			errz.Fatal(err)

			success, err := task.ArtifactExtract(hashIn, rebuild.VerifyResult.InvalidFiles)
			if err != nil {
				// if local artifact is corrupted due to incomplete previous download, try a fresh download
				if errors.Is(err, io.ErrUnexpectedEOF) {
					err = p.pullArtifact(ctx, hashIn, task, true)
					errz.Fatal(err)
					success, err = task.ArtifactExtract(hashIn, rebuild.VerifyResult.InvalidFiles)
				}
			}

			errz.Fatal(err)
			if success {
				rebuild.IsRequired = false

				// In case an artifact was synced from the remote store no buildinfo exists...
				// To avoid subsequent artifact extraction the Buildinfo is created after
				// extracting the artifact.
				buildInfo, err := p.computeBuildinfo(task.Name())
				errz.Fatal(err)
				err = p.storeBuildInfo(task.Name(), buildInfo)
				errz.Fatal(err)
			}
		case TargetInvalid:
			boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s, extracting artifact", p.namePad, coloredName, rebuild.Cause))
			hashIn, err := task.HashIn()
			errz.Fatal(err)
			success, err := task.ArtifactExtract(hashIn, rebuild.VerifyResult.InvalidFiles)
			errz.Fatal(err)
			if success {
				rebuild.IsRequired = false
			}
		case TargetNotInLocalStore:
		case TaskForcedRebuild:
		case DependencyChanged:
		default:
		}
	}

	if !rebuild.IsRequired {
		status := StateNoRebuildRequired
		boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, status.Short()))
		taskSuccessFul = true
		return pt, p.TaskNoRebuildRequired(task.TaskID)
	}

	err = task.CleanTargetsWithReason(rebuild.VerifyResult.InvalidFiles)
	errz.Fatal(err)

	err = task.Run(ctx, p.namePad)
	if err != nil {
		taskSuccessFul = false
		taskErr = err
	}
	errz.Fatal(err)

	// FIXME: Is this placed correctly?
	// Could also be done after the task completion is
	// done (artifact validation & packaging).
	//
	// What does it do? It prevents the task from beeing
	// flagged as failed in a defered function call.
	taskSuccessFul = true

	err = p.TaskCompleted(task.TaskID)
	if errors.Is(err, ErrFailed) {
		return pt, err
	}
	errz.Log(err)
	errz.Fatal(err)

	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)

	state := taskStatus.State()
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, "..."+state.Short()))

	return pt, nil
}
