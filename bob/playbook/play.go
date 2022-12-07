package playbook

import (
	"errors"
	"fmt"
	"time"

	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/usererror"
)

func (p *Playbook) Play() (err error) {
	return p.play()
}

func (p *Playbook) play() error {

	if p.done {
		return ErrDone
	}

	p.playMutex.Lock()
	defer p.playMutex.Unlock()

	if p.start.IsZero() {
		p.start = time.Now()
	}

	// Walk the task chain and determine the next build task. Send it to the task channel.
	// Returns `taskQueued` when a task has been send to the taskChannel.
	// Returns `taskFailed` when a task has failed.
	// Once it returns `nil` the playbook is done with it's work.
	var taskQueued = fmt.Errorf("task queued")
	var taskFailed = fmt.Errorf("task failed")
	err := p.Tasks.walk(p.root, func(taskname string, task *Status, err error) error {
		if err != nil {
			return err
		}

		// boblog.Log.V(3).Info(fmt.Sprintf("%-*s\t walking", p.namePad, taskname))

		switch task.State() {
		case StatePending:
			// Check if all dependent tasks are completed
			for _, dependentTaskName := range task.Task.DependsOn {
				t, ok := p.Tasks[dependentTaskName]
				if !ok {
					// fmt.Printf("Task %s does not exist", dependentTaskName)
					return usererror.Wrap(boberror.ErrTaskDoesNotExistF(dependentTaskName))
				}

				state := t.State()
				if state != StateCompleted && state != StateCached && state != StateNoRebuildRequired {
					// A dependent task is not completed.
					// So this task is not yet ready to run.
					return nil
				}
			}
		case StateFailed:
			return taskFailed
		case StateCanceled:
			return nil
		case StateCached, StateNoRebuildRequired:
			return nil
		case StateCompleted:
			return nil
		case StateRunning:
			return nil
		default:
		}

		// fmt.Printf("sending task %s to channel\n", task.Task.Name())
		// setting the task start time before passing it to channel
		task.SetStart(time.Now())
		// TODO: for async assure to handle send to a closed channel.
		_ = p.setTaskState(task.Name(), StateRunning, nil)
		p.taskChannel <- task.Task
		return taskQueued
	})

	// taskQueued => return nil (happy path)
	// taskFailed => return PlaybookFailed (ErrFailed)
	// default    => return err
	if err != nil {
		if errors.Is(err, taskQueued) {
			return nil
		}
		if errors.Is(err, taskFailed) {
			return ErrFailed
		}
		return err
	}

	// When arriving here it means that either there is currently
	// no work necessary or that the playbook is done processing all tasks.

	// Finishing the playbook when there is no work left.
	if !p.hasRunningOrPendingTasks() {
		p.Done()
		return ErrDone
	}

	return nil
}
