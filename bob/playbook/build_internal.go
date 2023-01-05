package playbook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/processed"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task *bobtask.Task) (_ processed.Task, err error) {
	defer errz.Recover(&err)

	pt := processed.Task{Task: task}

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

	// Filter inputs populates the task input member by reading and validating
	// inputs with the filesystem.
	start := time.Now()
	// err = task.FilterInputs()
	// errz.Fatal(err)
	pt.FilterInputTook = time.Since(start)

	start = time.Now()
	rebuildRequired, rebuildCause, err := p.TaskNeedsRebuild(task.TaskID, &pt)
	errz.Fatal(err)
	pt.NeddRebuildTook = time.Since(start)
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
		status := StateNoRebuildRequired
		boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, status.Short()))
		taskSuccessFul = true
		return pt, p.TaskNoRebuildRequired(task.TaskID)
	}

	err = task.Clean()
	errz.Fatal(err)

	start = time.Now()
	err = task.Run(ctx, p.namePad)
	if err != nil {
		taskSuccessFul = false
		taskErr = err
	}
	errz.Fatal(err)
	pt.BuildTook = time.Since(start)

	// FIXME: Is this placed correctly?
	// Could also be done after the task completion is
	// done (artifact validation & packaging).
	//
	// What does it do? It prevents the task from beeing
	// flagged as failed in a defered function call.
	taskSuccessFul = true

	start = time.Now()
	err = p.TaskCompleted(task.TaskID)
	if err != nil {
		if errors.Is(err, ErrFailed) {
			pt.CompletionTook = time.Since(start)
			return pt, err
		}
	}
	errz.Fatal(err)
	pt.CompletionTook = time.Since(start)

	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)

	state := taskStatus.State()
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, "..."+state.Short()))

	return pt, nil
}
