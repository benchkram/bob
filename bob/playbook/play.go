package playbook

import (
	"errors"
	"fmt"
	"time"

	"github.com/benchkram/bob/bobtask"
)

func (p *Playbook) Play() (err error) {

	return p.playOnce()
}

func (p *Playbook) Next() (_ *bobtask.Task, err error) {
	if p.done {
		return nil, ErrDone
	}
	p.oncePrepareOptimizedAccess.Do(func() {
		_ = p.Tasks.walk(p.root, func(taskname string, task *Status, _ error) error {
			for _, dependentTaskName := range task.DependsOn {
				t := p.Tasks[dependentTaskName]
				task.DependsOnIDs = append(task.DependsOnIDs, t.TaskID)
			}
			return nil
		})
	})

	// Required?
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
	//var noTaskReadyToRun = fmt.Errorf("no task ready to run")

	type result struct {
		t     *bobtask.Task
		state string // queued, playbook-done, failed
	}
	c := make(chan result, 1)

	// Starting the walk function in a goroutine to be able
	// to return  a ready to be processed task immeadiately
	// from Next().
	go func(output chan result) {
		didAllTaskComplete := true
		_ = p.TasksOptimized.walkBottomFirst(p.rootID, func(taskID int, task *Status, err error) error {
			if err != nil {
				return err
			}

			//boblog.Log.V(3).Info(fmt.Sprintf("%-*d\t walking", p.namePad, taskID))

			switch task.State() {
			case StatePending:
				didAllTaskComplete = false
				// Check if all dependent tasks are completed
				for _, dependentTaskID := range task.Task.DependsOnIDs {
					t := p.TasksOptimized[dependentTaskID]

					state := t.State()
					if state != StateCompleted && state != StateNoRebuildRequired {
						// A dependent task is not completed.
						// So this task is not yet ready to run.
						return nil
					}
				}
			case StateFailed:
				//output <- result{t: task.Task, state: "failed"}
				return taskFailed
			case StateCanceled:
				//output <- result{t: task.Task, state: "canceled"}
				return nil
			case StateNoRebuildRequired:
				return nil
			case StateCompleted:
				return nil
			case StateRunning:
				didAllTaskComplete = false
				return nil
			case StateQueued:
				didAllTaskComplete = false
				return nil
			default:
			}

			//fmt.Printf("sending task %s to channel\n", task.Task.Name())
			// setting the task start time before passing it to channel
			task.SetStart(time.Now())
			// TODO: for async assure to handle send to a closed channel.
			_ = p.setTaskState(task.Name(), StateRunning, nil)
			output <- result{t: task.Task, state: "queued"}
			return taskQueued
		})

		if didAllTaskComplete {
			output <- result{t: nil, state: "playbook-done"}
		}
		close(output)
	}(c)

	for r := range c {
		switch r.state {
		case "queued":
			//fmt.Printf("received task %s and returning\n", r.t.Name())
			return r.t, nil
		case "failed":
			fallthrough
		case "playbook-done":
			p.done = true
			return nil, ErrDone
		}
	}

	//fmt.Printf("returning without a task\n")
	return nil, nil
	// // wait
	// select {
	// case cc := <-c:
	// 	// task ready
	// 	fmt.Printf("received task %s and returning\n", cc.Name())
	// 	return cc, nil
	// default:
	// 	// no task ready
	// 	fmt.Printf("returning without a task\n")
	// 	return nil, nil
	// }

	// // taskQueued => return nil (happy path)
	// // taskFailed => return PlaybookFailed (ErrFailed)
	// // default    => return err
	// if err != nil {
	// 	if errors.Is(err, taskQueued) {
	// 		return nil, nil
	// 	}
	// 	if errors.Is(err, taskFailed) {
	// 		return nil, ErrFailed
	// 	}
	// 	return nil, err
	// }

	// // When arriving here it means that either there is currently
	// // no work necessary or that the playbook is done processing all tasks.

	// // Finishing the playbook when there is no work left.
	// if !p.hasRunningOrPendingTasks() {
	// 	p.Done()
	// 	return nil, ErrDone
	// }

}

func (p *Playbook) play() error {

	if p.done {
		return ErrDone
	}
	p.oncePrepareOptimizedAccess.Do(func() {
		_ = p.Tasks.walk(p.root, func(taskname string, task *Status, _ error) error {
			for _, dependentTaskName := range task.DependsOn {
				t := p.Tasks[dependentTaskName]
				task.DependsOnIDs = append(task.DependsOnIDs, t.TaskID)
			}
			return nil
		})
	})

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
	//var noTaskReadyToRun = fmt.Errorf("no task ready to run")
	err := p.TasksOptimized.walk(p.rootID, func(taskID int, task *Status, err error) error {
		if err != nil {
			return err
		}

		//boblog.Log.V(3).Info(fmt.Sprintf("%-*s\t walking", p.namePad, taskname))

		switch task.State() {
		case StatePending:
			// Check if all dependent tasks are completed
			for _, dependentTaskID := range task.Task.DependsOnIDs {
				t := p.TasksOptimized[dependentTaskID]

				state := t.State()
				if state != StateCompleted && state != StateNoRebuildRequired {
					// A dependent task is not completed.
					// So this task is not yet ready to run.
					return nil
				}
			}
		case StateFailed:
			return taskFailed
		case StateCanceled:
			return nil
		case StateNoRebuildRequired:
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

func (p *Playbook) playOnce() error {

	if p.done {
		return ErrDone
	}
	p.oncePrepareOptimizedAccess.Do(func() {
		_ = p.Tasks.walk(p.root, func(taskname string, task *Status, _ error) error {
			for _, dependentTaskName := range task.DependsOn {
				t := p.Tasks[dependentTaskName]
				task.DependsOnIDs = append(task.DependsOnIDs, t.TaskID)
			}
			return nil
		})
	})

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
	//var noTaskReadyToRun = fmt.Errorf("no task ready to run")
	err := p.TasksOptimized.walkBottomFirst(p.rootID, func(taskID int, task *Status, err error) error {
		if err != nil {
			return err
		}

		//boblog.Log.V(3).Info(fmt.Sprintf("%-*s\t walking", p.namePad, taskname))

		switch task.State() {
		case StatePending:
			// Queue task
		case StateFailed:
			return taskFailed
		case StateCanceled:
			return nil
		case StateNoRebuildRequired:
			return nil
		case StateCompleted:
			return nil
		case StateRunning:
			return nil
		case StateQueued:
			return nil
		default:
		}

		// fmt.Printf("sending task %s to channel\n", task.Task.Name())
		// setting the task start time before passing it to channel
		task.SetStart(time.Now())
		// TODO: for async assure to handle send to a closed channel.
		_ = p.setTaskState(task.Name(), StateRunning, nil)
		p.taskChannel <- task.Task
		return nil
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

	// // When arriving here it means that either there is currently
	// // no work necessary or that the playbook is done processing all tasks.

	// // Finishing the playbook when there is no work left.
	// if !p.hasRunningOrPendingTasks() {
	// 	p.Done()
	// 	return ErrDone
	// }

	return nil
}
