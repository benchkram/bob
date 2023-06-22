package playbook

import "github.com/logrusorgru/aurora"

type State string

// Summary state indicators.
// The nbsp are intended to align on the cli.
func (s *State) Summary() string {
	switch *s {
	case StatePending:
		return "⌛       "
	case StateCompleted:
		return aurora.Green("✔").Bold().String() + "       "
	case StateNoRebuildRequired:
		return aurora.Green("cached").String() + "  "
	case StateFailed:
		return aurora.Red("failed").String() + "  "
	case StateCanceled:
		return aurora.Faint("canceled").String()
	case StateQueued:
		return aurora.Faint("queued").String()
	default:
		return ""
	}
}

func (s *State) Short() string {
	switch *s {
	case StatePending:
		return "pending"
	case StateCompleted:
		return "done"
	case StateNoRebuildRequired:
		return "cached"
	case StateFailed:
		return "failed"
	case StateCanceled:
		return "canceled"
	case StateQueued:
		return "queued"
	default:
		return ""
	}
}

const (
	StatePending           State = "PENDING"
	StateCompleted         State = "COMPLETED"
	StateNoRebuildRequired State = "CACHED"
	StateFailed            State = "FAILED"
	StateRunning           State = "RUNNING"
	StateCanceled          State = "CANCELED"
	StateQueued            State = "QUEUED"
)
