package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
)

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {

	// Setup worker pool and queue.
	workers := p.maxParallel
	boblog.Log.Info(fmt.Sprintf("Using %d workers", workers))

	p.pickTaskColors()

	wm := p.startWorkers(ctx, workers)

	// listen for idle workers
	go func() {
		// A buffer for workers which have
		// no workload assigned.
		workerBuffer := []int{}

		for workerID := range wm.idleChan {

			// boblog.Log.V(1).Info("Calling Next")
			task, err := p.Next()
			if err != nil {

				if errors.Is(err, ErrDone) {
					//boblog.Log.V(1).Info("Done")
					wm.stopWorkers()
					// exit
					break
				}

				wm.addError(fmt.Errorf("worker-availability-queue: unexpected error comming from Next(): %w", err))
				wm.stopWorkers()
				break
			}

			// Push workload to the worker or store the worker for later.
			if task != nil {
				//boblog.Log.V(1).Info(fmt.Sprintf("Sending task %s to worker", task.Name()))
				// Send workload to worker
				wm.workloadQueues[workerID] <- task

				// There might be more workload left.
				// Reqeuing a worker from the buffer.
				if len(workerBuffer) > 0 {
					wID := workerBuffer[len(workerBuffer)-1]
					workerBuffer = workerBuffer[:len(workerBuffer)-1]

					//	boblog.Log.V(1).Info("Requeue Worker")
					// requeue a buffered worker
					wm.idleChan <- wID
				}
			} else {

				// No task yet ready to be worked on but the playbook is not done yet.
				// Therfore the worker is stored in a buffer and is requeued on
				// the next change to the playbook.
				workerBuffer = append(workerBuffer, workerID)
				//boblog.Log.V(1).Info(fmt.Sprintf("Buffering Worker [worker_id:%d ] [%s:%d]", workerID, "buffer_size", len(workerBuffer)))

			}

			//wm.printWorkerState()
		}

		// to assure even idling workers will be shutdown.
		wm.closeWorkloadQueues()
	}()

	wm.workerWG.Wait()

	// iterate through tasks and logs
	// skipped input files.
	var skippedInputs int
	for _, task := range wm.processed {
		skippedInputs = logSkippedInputs(
			skippedInputs,
			task.ColoredName(),
			task.LogSkippedInput(),
		)
	}

	//p.summary(processedTasks)

	if len(wm.errors) > 0 {
		// Pass only the very first processing error.
		return wm.errors[0]
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
