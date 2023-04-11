package bobtask

import (
	"fmt"
	"os"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
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

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		if vb {
			fmt.Printf("Cleaning targets of task %s \n", t.name)
		}

		if t.dir == "" {
			return fmt.Errorf("task dir not set")
		}

		for filename, reasons := range invalidFiles {
			for _, reason := range reasons {
				if reason == target.ReasonCreatedAfterBuild || reason == target.ReasonForcedByNoCache {
					if vb {
						fmt.Printf(" %s ", filename)
					}

					if filename == "/" || filename == homeDir {
						return fmt.Errorf("Cleanup of %s is not allowed", filename)
					}

					err := os.RemoveAll(filename)
					if err != nil {
						if vb {
							fmt.Printf("%s\n", aurora.Red("failed"))
						}
						return err
					}
					if vb {
						fmt.Printf("%s\n", aurora.Green("done"))
					}
				}
			}
		}
	}

	return nil
}
