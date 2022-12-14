package bobtask

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boblog"
)

// Clean the targets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean(invalidFiles map[string][]target.Reason, verbose ...bool) error {
	var vb bool
	if len(verbose) > 0 {
		vb = verbose[0]
	}

	if t.target != nil {
		boblog.Log.V(5).Info(fmt.Sprintf("Cleaning targets for task %s", t.Name()))

		if vb {
			fmt.Printf("Cleaning targets of task %s \n", t.name)
		}

		if t.dir == "" {
			return fmt.Errorf("task dir not set")
		}
		for k, v := range invalidFiles {
			if v[0] == target.ReasonCreatedAfterBuild {
				p := filepath.Join(t.dir, k)
				os.RemoveAll(p)
			}
		}
	}

	return nil
}
