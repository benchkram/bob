package playbook

import (
	"fmt"
)

func (p *Playbook) Next() (_ *Status, err error) {
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

	// Walk the task chain and determine the next build task. Send it to the task channel.
	// Returns `taskQueued` when a task has been send to the taskChannel.
	// Returns `taskFailed` when a task has failed.
	// Once it returns `nil` the playbook is done with it's work.
	var taskQueued = fmt.Errorf("task queued")
	var taskFailed = fmt.Errorf("task failed")
	//var noTaskReadyToRun = fmt.Errorf("no task ready to run")

	type result struct {
		t     *Status
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
				output <- result{t: task, state: "failed"}
				return taskFailed
			case StateCanceled:
				output <- result{t: task, state: "canceled"}
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

			// TODO: for async assure to handle send to a closed channel.
			_ = p.setTaskState(task.TaskID, StateQueued, nil)
			output <- result{t: task, state: "queued"}
			return taskQueued
		})

		if didAllTaskComplete {
			output <- result{t: nil, state: "playbook-done"}
		}
		close(output)

		// the goroutine can outlive the Next() run.
		// avoiding concurrrent runs by only unlocking at the
		// end of the walk.
		//p.playMutex.Unlock()

	}(c)

	for r := range c {
		switch r.state {
		case "queued":
			//fmt.Printf("received task %s and returning\n", r.t.Name())
			return r.t, nil
		case "failed":
			//fmt.Println("failed")
			fallthrough
		case "canceled":
			//fmt.Println("failed")
			fallthrough
		case "playbook-done":
			//fmt.Println("playbook-done")
			p.done = true
			return nil, ErrDone
		}
	}

	//fmt.Printf("returning without a task\n")
	return nil, nil

}
