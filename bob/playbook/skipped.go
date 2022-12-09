package playbook

import (
	"fmt"
	"os"

	"github.com/benchkram/bob/pkg/boblog"
)

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
