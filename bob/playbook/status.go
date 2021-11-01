package playbook

import (
	"sync"
	"time"

	"github.com/Benchkram/bob/bobtask"
)

// Status holds the state of a task
// inside a playbook.
type Status struct {
	bobtask.Task

	stateMu sync.RWMutex
	state   string

	Start time.Time
	End   time.Time
}

func NewStatus(task bobtask.Task) *Status {
	return &Status{
		Task:  task,
		state: StatePending,
		Start: time.Now(),
	}
}

func (ts *Status) State() string {
	ts.stateMu.RLock()
	defer ts.stateMu.RUnlock()
	return ts.state
}

func (ts *Status) SetState(s string) {
	ts.stateMu.Lock()
	defer ts.stateMu.Unlock()
	ts.state = s
}

func (ts *Status) ExecutionTime() time.Duration {
	return ts.End.Sub(ts.Start)
}
