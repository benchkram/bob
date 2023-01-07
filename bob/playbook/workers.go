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

	workerStateMutex sync.Mutex
	workerState      []string

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
		workerState:    []string{},

		errors:    []error{},
		processed: []*processed.Task{},
	}
	return s
}

func (wm *workerManager) stopWorkers() {
	//boblog.Log.V(1).Info("stopping workers")
	wm.shutdownMutext.Lock()
	wm.shutdown = true
	wm.shutdownMutext.Unlock()
}

func (wm *workerManager) closeWorkloadQueues() {
	//boblog.Log.V(1).Info("closing workload queues")
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
	//boblog.Log.V(1).Info("adding error to worker manager")
	wm.errorsMutex.Lock()
	wm.errors = append(wm.errors, err)
	wm.errorsMutex.Unlock()
}

func (wm *workerManager) addProcessedTask(t *processed.Task) {
	wm.processedMutex.Lock()
	wm.processed = append(wm.processed, t)
	wm.processedMutex.Unlock()
}

// func (wm *workerManager) setWorkerState(workerID int, state string) {
// 	wm.workerStateMutex.Lock()
// 	wm.workerState[workerID] = state
// 	wm.workerStateMutex.Unlock()
// }
func (wm *workerManager) printWorkerState() {
	wm.workerStateMutex.Lock()
	for i, s := range wm.workerState {
		fmt.Printf("worker %d is in state %s\n", i, s)
	}

	wm.workerStateMutex.Unlock()
}

// startWorkers and return a state to interact with the workers.
func (p *Playbook) startWorkers(ctx context.Context, workers int) *workerManager {

	wm := newWorkerManager()

	// initially set all workers to idle
	// for i := 0; i < workers; i++ {
	// 	wm.workerState = append(wm.workerState, "starting")
	// }

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
			//wm.setWorkerState(workerID, "idle")
			// signal availability to receive workload
			wm.idleChan <- workerID

			for t := range queue {
				t.SetStart(time.Now())
				_ = p.setTaskState(t.Task.TaskID, StateRunning, nil)
				//wm.setWorkerState(workerID, "running")

				// check if a shutdown is required.
				if wm.canShutdown() {
					break
				}

				//boblog.Log.V(1).Info(fmt.Sprintf("RUNNING task %s on worker %d", t.Name(), workerID))

				processedTask, err := p.build(ctx, t.Task)
				if err != nil {
					wm.addError(fmt.Errorf("(worker) [task: %s], %w", t.Name(), err))

					// stopp workers asap.
					wm.stopWorkers()
					//shutdownAvailabilityQueue()

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
				wm.addProcessedTask(processedTask)

				//wm.setWorkerState(workerID, "idle")

				// done with processing. signal availability.
				wm.idleChan <- workerID
			}
			//fmt.Printf("worker %d is shutting down\n", workerID)
			//wm.setWorkerState(workerID, "ended")
			wm.workerWG.Done()
		}(i)
	}

	return wm

}
