package playbook

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/errz"
)

func (p *Playbook) BuildTask(ctx context.Context, taskname string) (err error) {
	done := make(chan error)

	go func() {
		// TODO: Run a worker pool so that multiple tasks can run in parallel.
		// https://stackoverflow.com/questions/25306073/always-have-x-number-of-goroutines-running-at-any-time

		c := p.TaskChannel()
		for task := range c {
			err := p.buildSingleTask(ctx, taskname, task)
			if errors.Is(err, context.Canceled) || errors.Is(err, ErrFailed) {
				done <- err
				break
			}
			killOnError(err)
		}

		close(done)
	}()

	_ = p.Play()
	err = <-done
	fmt.Printf("\n\nDone running playbook in %s\n", p.ExecutionTime())
	return err
}

// Run a single task (of potentially a parent) in a playbook.
func (p *Playbook) buildSingleTask(ctx context.Context, taskname string, task bobtask.Task) (err error) {
	// TODO: Run a worker pool so that multiple tasks can run in parallel.
	// https://stackoverflow.com/questions/25306073/always-have-x-number-of-goroutines-running-at-any-time

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				err := p.TaskCanceled(task.Name())
				killOnError(err)
				fmt.Printf("Task %q was canceled\n", task.Name())
			}
		}
	}()

	println()
	println()
	fmt.Printf("Beginning to run task: %q\n", task.Name())

	rebuildRequired, err := p.TaskNeedsRebuild(task.Name())
	killOnError(err)

	if !rebuildRequired {
		err = p.TaskNoRebuildRequired(task.Name())
		killOnError(err)
		fmt.Printf("Task %q doesn't need to be rebuilt\n", task.Name())
		return nil
	}

	err = task.Clean()
	killOnError(err)

	err = task.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return err
	}
	killOnError(err)

	err = task.VerifyAfter()
	killOnError(err)

	// 3: Check whether output is created correctly.
	//    Might be a build error or a configuration
	//    error when the output does not exist.
	// -> task.Target
	succeeded, failedTargets := task.DidSucceede()
	if succeeded {
		err = p.TaskCompleted(task.Name())
		if err != nil {
			errz.Log(err)
		}
		taskStatus, err := p.TaskStatus(task.Name())
		if err != nil {
			errz.Log(err)
		}
		fmt.Printf("Task %q completed in %s\n", task.Name(), taskStatus.ExecutionTime())
	} else {
		for _, target := range failedTargets {
			fmt.Printf("Target %q does not exist", target)
		}
		err = p.TaskFailed(task.Name())
		if err != nil {
			if errors.Is(err, ErrFailed) {
				return err
			}
		}
		killOnError(err)

		fmt.Printf("Task %q failed\n", taskname)
	}

	return nil
}

func killOnError(err error) {
	if err != nil {
		errz.Log(err)
		os.Exit(1)
	}
}
