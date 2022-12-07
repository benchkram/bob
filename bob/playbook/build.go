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
	workerQueues := []chan *bobtask.Task{}
	workerAvailabilityQueue := make(chan int, 1000)

	boblog.Log.Info(fmt.Sprintf("Using %d workers", workers))

	runningWorkers := sync.WaitGroup{}

	var once sync.Once
	shutdownWorkers := func() {
		once.Do(func() {
			// Intitate gracefull shutdown of all workers
			// by closing their queues.
			for _, wq := range workerQueues {
				close(wq)
			}
			// clearing the queue
			workerQueues = []chan *bobtask.Task{}
		})
	}

	// Start the workers which listen on task queue
	for i := 0; i < workers; i++ {
		queue := make(chan *bobtask.Task)
		workerQueues = append(workerQueues, queue)
		runningWorkers.Add(1)
		go func(workerID int) {
			// initially signal availability to receive workload
			workerAvailabilityQueue <- workerID

			for t := range queue {

				boblog.Log.V(5).Info(fmt.Sprintf("RUNNING task %s on worker  %d ", t.Name(), workerID))

				err := p.build(ctx, t)
				if err != nil {
					processingErrorsMutex.Lock()
					processingErrors = append(processingErrors, fmt.Errorf("(worker) [task: %s], %w", t.Name(), err))
					processingErrorsMutex.Unlock()

					shutdownWorkers()

					// Any error occurred during a build puts the
					// playbook in a done state. This prevents
					// further tasks be queued for execution.
					// p.Done()

					// TODO: shutdown gracefully.
					// close(workerAvailabilityQueue)
					// for _, wq := range workerQeues {
					//  close(wq)
					// }

				}

				processedTasks = append(processedTasks, t)

				// done with processing. signal availability.
				select {
				case workerAvailabilityQueue <- workerID:
				default:
				}
			}
			runningWorkers.Done()
		}(i + 1)
	}

	// listen for available workers
	go func() {
		// A buffer for workers which have
		// no workload assigned.
		workerBuffer := []int{}

		for workerID := range workerAvailabilityQueue {
			task, err := p.Next()
			if err != nil {
				if errors.Is(err, ErrDone) {
					shutdownWorkers()

					// exit
					return
				}

				processingErrorsMutex.Lock()
				processingErrors = append(processingErrors, fmt.Errorf("worker-availability-queue: unexpected error comming from Next(): %w", err))
				processingErrorsMutex.Unlock()
				return
			}

			// Push workload to the worker or store the worker for later.
			if task != nil {
				// Send workload to worker
				workerQueues[workerID-1] <- task

				// There might be more workload left.
				// Reqeuing a workler from the buffer.
				if len(workerBuffer) > 0 {
					wID := workerBuffer[len(workerBuffer)-1]
					workerBuffer = workerBuffer[:len(workerBuffer)-1]

					// requeue a buffered worker
					workerAvailabilityQueue <- wID
				}
			} else {
				// No task yet ready to be worked on butt the playbook is not done yet.
				// Therfore the worker is stored in a buffer and is requeued on
				// the next change to the playbook.
				workerBuffer = append(workerBuffer, workerID)
			}
		}

	}()

	runningWorkers.Wait()
	close(workerAvailabilityQueue)

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

	//p.summary(processedTasks)

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
