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
	"github.com/benchkram/bob/pkg/store"
)

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {
	processingErrorsMutex := sync.Mutex{}
	var processingErrors []error

	var processedTasks []*bobtask.Task

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

					// Any error occurred during a build puts the
					// playbook in a done state. This prevents
					// further tasks be queued for execution.

					p.Done()
				}
			}
		}(i + 1)
	}

	// Listen for tasks from the playbook and forward them to the worker pool
	go func() {
		c := p.TaskChannel()
		for t := range c {
			boblog.Log.V(5).Info(fmt.Sprintf("Sending task %s", t.Name()))
			processedTasks = append(processedTasks, t)

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

	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		// sync any newly generated artifacts with the remote store
		syncFromLocalToRemote(ctx, p.localStore, p.remoteStore, getArtifactIds(p, true))
	}

	if len(processingErrors) > 0 {
		// Pass only the very first processing error.
		return processingErrors[0]
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

// getArtifactIds returns the artifact ids of the given playbook (and optionally checks if the target exists first)
func getArtifactIds(pbook *Playbook, checkForTarget bool) []hash.In {
	var artifactIds []hash.In
	for _, t := range pbook.Tasks {
		if checkForTarget && !t.TargetExists() {
			continue
		}

		h, err := t.HashIn()
		if err != nil {
			continue
		}

		artifactIds = append(artifactIds, h)
	}
	return artifactIds
}

// syncFromLocalToRemote syncs the artifacts from the local store to the remote store.
func syncFromLocalToRemote(ctx context.Context, local store.Store, remote store.Store, artifactIds []hash.In) {
	for _, a := range artifactIds {
		err := store.Sync(ctx, local, remote, a.String())
		if errors.Is(err, store.ErrArtifactAlreadyExists) {
			boblog.Log.V(1).Info(fmt.Sprintf("artifact already exists on the remote [artifactId: %s]. skipping...", a.String()))
			continue
		} else if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		// wait for the remote store to finish uploading this artifact. can be moved outside of the for loop but then
		// we don't know which artifacts failed to upload.
		err = remote.Done()
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		boblog.Log.V(1).Info(fmt.Sprintf("synced from local to remote [artifactId: %s]", a.String()))
	}
}
