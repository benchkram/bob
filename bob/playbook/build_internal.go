package playbook

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task *bobtask.Task) (err error) {
	defer errz.Recover(&err)

	// A task is flagged successful before
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

	// task might need a rebuild due to an input change.
	// Could still be possible to load the targets from the artifact store.
	// If a task needs a rebuild due to a dependency change => rebuild.
	if rebuildRequired {
		switch rebuildCause {
		case InputNotFoundInBuildInfo:
			hashIn, err := task.HashIn()
			errz.Fatal(err)

			// pull artifact if it exists on the remote. if exists locally will use that one
			err = p.pullArtifact(ctx, hashIn, task, false)
			errz.Fatal(err)

			success, err := task.ArtifactExtract(hashIn)
			if err != nil {
				// if local artifact is corrupted due to incomplete previous download, try a fresh download
				if errors.Is(err, io.ErrUnexpectedEOF) {
					err = p.pullArtifact(ctx, hashIn, task, true)
					errz.Fatal(err)
					success, err = task.ArtifactExtract(hashIn)
				}
			}

			errz.Fatal(err)
			if success {
				rebuildRequired = false

				// In case an artifact was synced from the remote store no buildinfo exists...
				// To avoid subsequent artifact extraction the Buildinfo is created after
				// extracting the artifact.
				buildInfo, err := p.computeBuildinfo(task.Name())
				errz.Fatal(err)
				err = p.storeBuildInfo(task.Name(), buildInfo)
				errz.Fatal(err)
			}
		case TargetInvalid:
			boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s, extracting artifact", p.namePad, coloredName, rebuildCause))
			hashIn, err := task.HashIn()
			errz.Fatal(err)
			success, err := task.ArtifactExtract(hashIn)
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
		status := StateCached
		t, _ := task.Target()
		if t == nil {
			status = StateNoRebuildRequired
		}
		boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, status.Short()))
		taskSuccessFul = true
		return p.TaskNoRebuildRequired(task.Name(), status)
	}

	err = task.Clean()
	errz.Fatal(err)

	err = task.Run(ctx, p.namePad, p.nixCache, p.nixShellCache)
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

	err = p.TaskCompleted(task.Name())
	if err != nil {
		if errors.Is(err, ErrFailed) {
			return err
		}
	}
	errz.Fatal(err)

	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)

	state := taskStatus.State()
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, "..."+state.Short()))

	return nil
}
