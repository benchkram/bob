package playbook

import (
	"sync"
	"time"

	"github.com/benchkram/bob/bobtask"
)

// Status holds the state of a task
// inside a playbook.
type Status struct {
	*bobtask.Task

	stateMu sync.RWMutex
	state   State

	startMu sync.RWMutex
	start   time.Time
	endMu   sync.RWMutex
	end     time.Time

	Error error
}

func NewStatus(task *bobtask.Task) *Status {
	return &Status{
		Task:  task,
		state: StatePending,
		start: time.Now(),
	}
}

func (ts *Status) State() State {
	ts.stateMu.RLock()
	s := ts.state
	ts.stateMu.RUnlock()
	return s
}

func (ts *Status) SetState(s State, err error) {
	ts.stateMu.Lock()
	ts.state = s
	ts.Error = err
	ts.stateMu.Unlock()
	//boblog.Log.Info(fmt.Sprintf("Setting taskstate of [task_name: %s] to [state: %s]", ts.Name(), s))
}

func (ts *Status) ExecutionTime() time.Duration {
	ts.startMu.RLock()
	ts.endMu.RLock()
	defer func() {
		ts.startMu.RUnlock()
		ts.endMu.RUnlock()
	}()

	return ts.end.Sub(ts.start)
}

func (ts *Status) Start() time.Time {
	ts.startMu.RLock()
	defer ts.startMu.RUnlock()
	return ts.start
}

func (ts *Status) SetStart(start time.Time) {
	ts.startMu.Lock()
	defer ts.startMu.Unlock()
	ts.start = start
}

func (ts *Status) End() time.Time {
	ts.endMu.RLock()
	defer ts.endMu.RUnlock()
	return ts.end
}

func (ts *Status) SetEnd(end time.Time) {
	ts.endMu.Lock()
	defer ts.endMu.Unlock()
	ts.end = end
}
