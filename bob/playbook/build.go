package playbook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
)

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {
	processingErrorsMutex := sync.Mutex{}
	var processingErrors []error

	var processedTasks []*bobtask.Task

	p.pickTaskColors()

	// Setup worker pool and queue.
	workers := p.maxParallel
	queue := make(chan *bobtask.Task)

	boblog.Log.Info(fmt.Sprintf("Using %d workers", workers))

	processing := sync.WaitGroup{}

	// Start the workers which listen on task queue
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			for t := range queue {
				processing.Add(1)
				boblog.Log.V(5).Info(fmt.Sprintf("RUNNING task %s on worker  %d ", t.Name(), workerID))
				err := p.build(ctx, t)
				if err != nil {
					processingErrorsMutex.Lock()
					processingErrors = append(processingErrors, fmt.Errorf("(build loop) [task: %s], %w", t.Name(), err))
					processingErrorsMutex.Unlock()

					// Any error occurred during a build puts the
					// playbook in a done state. This prevents
					// further tasks be queued for execution.
					p.Done()
				}

				processedTasks = append(processedTasks, t)
				processing.Done()
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

			// initiate another playbook run,
			// as there might be workers without
			// assigned tasks left.
			err := p.Play()
			if err != nil {
				if !errors.Is(err, ErrDone) {
					processingErrorsMutex.Lock()
					processingErrors = append(processingErrors, fmt.Errorf("(playbook exit) [task: %s], %w", t.Name(), err))
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
	processing.Wait()

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

	// sync any newly generated artifacts with the remote store
	if p.enablePush {
		for taskName, artifact := range p.inputHashes(true) {
			err = p.pushArtifact(ctx, artifact, taskName)
			if err != nil {
				return usererror.Wrap(err)
			}
		}
	}

	return nil
}

const maxSkippedInputs = 5

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

// inputHashes returns and array of input hashes of the playbook,
// optionally filters tasks without targets.
func (p *Playbook) inputHashes(filterTarget bool) map[string]hash.In {
	artifactIds := make(map[string]hash.In)

	for _, t := range p.Tasks {
		if filterTarget && !t.TargetExists() {
			continue
		}

		h, err := t.HashIn()
		if err != nil {
			continue
		}

		artifactIds[t.Name()] = h
	}
	return artifactIds
}
