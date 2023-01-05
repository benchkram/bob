package processed

import (
	"time"

	"github.com/benchkram/bob/bobtask"
)

type Task struct {
	*bobtask.Task

	FilterInputTook                    time.Duration
	NeddRebuildTook                    time.Duration
	NeedRebuildDidChildtaskChangeTook  time.Duration
	NeedRebuildDidTaskCHangeTook       time.Duration
	NeedRebuildTargetTook              time.Duration
	NeedRebuildTargetVerifyShallowTook time.Duration
	BuildTook                          time.Duration
	CompletionTook                     time.Duration
}
