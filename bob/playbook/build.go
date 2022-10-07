package playbook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
)

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {
	processingErrorsMutex := sync.Mutex{}
	processingErrors := []error{}

	processedTasks := []*bobtask.Task{}

	p.pickTaskColors()

	// Setup worker pool and queue
	parallelTasks := p.maxParallel
	queue := make(chan *bobtask.Task)

	boblog.Log.Info(fmt.Sprintf("Using %d workers", parallelTasks))

	for i := 0; i < parallelTasks; i++ {
		go func(workerID int) {
			for t := range queue {
				boblog.Log.V(5).Info(fmt.Sprintf("RUNNING task %s on worker  %d ", t.Name(), workerID))
				err := p.build(ctx, t)
				if err != nil {
					processingErrorsMutex.Lock()
					processingErrors = append(processingErrors, fmt.Errorf("[task: %s], %w", t.Name(), err))
					processingErrorsMutex.Unlock()

					// Any error occured during a build puts the
					// playbook in a done state. This prevents
					// further tasks be queued for execution.

					p.Done()
				}

				processedTasks = append(processedTasks, t)
			}
		}(i + 1)
	}

	// Listen for tasks from the playbook and forward them to the worker pool
	go func() {
		c := p.TaskChannel()
		for t := range c {
			boblog.Log.V(5).Info(fmt.Sprintf("Sending task %s", t.Name()))

			// blocks till a worker is available
			queue <- t

			// initiate another playbook run
			// as there might be workers without assigned tasks left.
			err := p.Play()
			if err != nil {
				if !errors.Is(err, ErrDone) {
					processingErrorsMutex.Lock()
					processingErrors = append(processingErrors, fmt.Errorf("[task: %s], %w", t.Name(), err))
					processingErrorsMutex.Unlock()
				}
				break
			}
		}
	}()

	err = p.Play()
	if err != nil {
		return err
	}

	<-p.DoneChan()

	close(queue)

	// iterate through tasks and logs
	// skipped input files.
	var skippedInputs int
	for _, task := range processedTasks {
		skippedInputs = logSkippedInputs(
			skippedInputs,
			task.ColoredName(),
			task.LogSkippedInput(),
		)
	}

	p.summary(processedTasks)

	if len(processingErrors) > 0 {
		// Pass only the very first processing error.
		return processingErrors[0]
	}

	return nil
}

// logSkippedInputs until max is reached
func logSkippedInputs(count int, taskname string, skippedInputs []string) int {
	if len(skippedInputs) == 0 {
		return count
	}
	if count >= maxSkippedInputs {
		return maxSkippedInputs
	}

	for _, f := range skippedInputs {
		count = count + 1
		boblog.Log.V(1).Info(fmt.Sprintf("skipped %s '%s' %s", taskname, f, os.ErrPermission))

		if count >= maxSkippedInputs {
			boblog.Log.V(1).Info(fmt.Sprintf("skipped %s %s", taskname, "& more..."))
			break
		}
	}

	return count
}
