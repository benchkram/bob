package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
)

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {
	done := make(chan error)

	go func() {
		// TODO: Run a worker pool so that multiple tasks can run in parallel.
		// https://stackoverflow.com/questions/25306073/always-have-x-number-of-goroutines-running-at-any-time

		c := p.TaskChannel()
		for task := range c {
			err := p.build(ctx, task)
			if err != nil {
				//if errors.Is(err, context.Canceled) || errors.Is(err, ErrFailed) {
				done <- err
				break
			}
		}

		close(done)
	}()

	_ = p.Play()
	err = <-done
	// fmt.Printf("\n\nDone running playbook in %s\n", p.ExecutionTime())
	return err
}

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task bobtask.Task) (err error) {
	defer errz.Recover(&err)

	// TODO: Run a worker pool so that multiple tasks can run in parallel.
	// https://stackoverflow.com/questions/25306073/always-have-x-number-of-goroutines-running-at-any-time

	boblog.Log.V(2).Info(fmt.Sprintf("Building [task: %s]", task.Name()))

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				boblog.Log.V(2).Info(fmt.Sprintf("Task %q was canceled", task.Name()))
				_ = p.TaskCanceled(task.Name())
			}
		}
	}()

	rebuildRequired, err := p.TaskNeedsRebuild(task.Name())
	errz.Fatal(err)

	if !rebuildRequired {
		boblog.Log.V(2).Info(fmt.Sprintf("Task %q doesn't need to be rebuilt", task.Name()))
		return p.TaskNoRebuildRequired(task.Name())
	}

	err = task.Clean()
	errz.Fatal(err)

	err = task.Run(ctx)
	errz.Fatal(err)

	err = task.VerifyAfter()
	errz.Fatal(err)

	target, err := task.Target()
	if err != nil {
		errz.Fatal(err)
	}

	// Check targets are created correctly.
	// On success the target hash is computed
	// inside TaskCompleted().
	if target != nil {
		if !target.Exists() {
			boblog.Log.V(2).Info(fmt.Sprintf("Task %q failed due to invalid targets", task.Name()))
			err = p.TaskFailed(task.Name())
			if err != nil {
				if errors.Is(err, ErrFailed) {
					return err
				}
			}
		}
	}

	err = p.TaskCompleted(task.Name())
	errz.Fatal(err)
	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)
	boblog.Log.V(2).Info(fmt.Sprintf("Task %q completed in %s", task.Name(), taskStatus.ExecutionTime()))

	return nil
}
