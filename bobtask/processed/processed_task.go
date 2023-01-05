package processed

import (
	"time"

	"github.com/benchkram/bob/bobtask"
)

type Task struct {
	*bobtask.Task

	FilterInputTook time.Duration
	NeddRebuildTook time.Duration
	BuildTook       time.Duration
	CompletionTook  time.Duration
}
