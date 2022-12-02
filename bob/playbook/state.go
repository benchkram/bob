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
	case StateCached:
		return aurora.Green("cached").String() + "  "
	case StateNoRebuildRequired:
		return aurora.Gray(10, "✔").Bold().String() + "       "
	case StateFailed:
		return aurora.Red("failed").String() + "  "
	case StateCanceled:
		return aurora.Faint("canceled").String()
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
	case StateCached:
		return "cached"
	case StateNoRebuildRequired:
		return "not-rebuild"
	case StateFailed:
		return "failed"
	case StateCanceled:
		return "canceled"
	default:
		return ""
	}
}

const (
	StatePending           State = "PENDING"
	StateCompleted         State = "COMPLETED"
	StateCached            State = "CACHED"
	StateNoRebuildRequired State = "NO-REBUILD"
	StateFailed            State = "FAILED"
	StateRunning           State = "RUNNING"
	StateCanceled          State = "CANCELED"
)
