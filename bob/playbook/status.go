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

	Start time.Time
	End   time.Time

	Error error
}

func NewStatus(task *bobtask.Task) *Status {
	return &Status{
		Task:  task,
		state: StatePending,
		Start: time.Now(),
	}
}

func (ts *Status) State() State {
	ts.stateMu.RLock()
	defer ts.stateMu.RUnlock()
	return ts.state
}

func (ts *Status) SetState(s State, err error) {
	ts.stateMu.Lock()
	defer ts.stateMu.Unlock()
	ts.state = s
	ts.Error = err
}

func (ts *Status) ExecutionTime() time.Duration {
	return ts.End.Sub(ts.Start)
}
