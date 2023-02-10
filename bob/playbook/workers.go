package playbook

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benchkram/bob/bobtask/processed"
)

type workerManager struct {
	// workerWG wait for all workers to be stopped.
	workerWG sync.WaitGroup

	// workloadQueues to send workload to a worker
	workloadQueues []chan *Status

	// workers report themself as idle
	// by putting their id on this channel.
	idleChan chan int

	errorsMutex sync.Mutex
	errors      []error

	processedMutex sync.Mutex
	processed      []*processed.Task

	shutdownMutext sync.Mutex
	shutdown       bool
}

func newWorkerManager() *workerManager {
	s := &workerManager{
		workloadQueues: []chan *Status{},
		idleChan:       make(chan int, 1000),

		errors:    []error{},
		processed: []*processed.Task{},
	}
	return s
}

func (wm *workerManager) stopWorkers() {
	wm.shutdownMutext.Lock()
	wm.shutdown = true
	wm.shutdownMutext.Unlock()
}

func (wm *workerManager) closeWorkloadQueues() {
	for _, q := range wm.workloadQueues {
		close(q)
	}
}

func (wm *workerManager) canShutdown() bool {
	wm.shutdownMutext.Lock()
	defer wm.shutdownMutext.Unlock()
	return wm.shutdown

}

func (wm *workerManager) addError(err error) {
	wm.errorsMutex.Lock()
	wm.errors = append(wm.errors, err)
	wm.errorsMutex.Unlock()
}

func (wm *workerManager) addProcessedTask(t *processed.Task) {
	wm.processedMutex.Lock()
	wm.processed = append(wm.processed, t)
	wm.processedMutex.Unlock()
}

// startWorkers and return a state to interact with the workers.
func (p *Playbook) startWorkers(ctx context.Context, workers int) *workerManager {

	wm := newWorkerManager()

	// Start the workers which listen on task queue
	for i := 0; i < workers; i++ {

		// create workload queue for this worker
		// A non-blocking queue with a too big size allows to shutdown gracefully in
		// case of error. The workload pump might pump more tasks to the queues
		// and to avoid blocking behaviour the queue size is not nil.
		queue := make(chan *Status, 1000)
		wm.workloadQueues = append(wm.workloadQueues, queue)

		wm.workerWG.Add(1)
		go func(workerID int) {
			// signal availability to receive workload
			wm.idleChan <- workerID

			for t := range queue {
				t.SetStart(time.Now())
				_ = p.setTaskState(t.Task.TaskID, StateRunning, nil)

				// check if a shutdown is required.
				if wm.canShutdown() {
					break
				}

				processedTask, err := p.build(ctx, t.Task)
				if err != nil {
					wm.addError(fmt.Errorf("(worker) [task: %s], %w", t.Name(), err))

					// stopp workers asap.
					wm.stopWorkers()
				}
				wm.addProcessedTask(processedTask)

				// done with processing. signal availability.
				wm.idleChan <- workerID
			}
			wm.workerWG.Done()
		}(i)
	}

	return wm

}
