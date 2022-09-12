package bobtask

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/logrusorgru/aurora"
)

// Clean the targets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean(verbose ...bool) error {
	var vb bool
	if len(verbose) > 0 {
		vb = verbose[0]
	}

	if t.target != nil {

		boblog.Log.V(5).Info(fmt.Sprintf("Cleaning targets for task %s", t.Name()))

		if vb {
			fmt.Printf("Cleaning targets of task %s \n", t.name)
		}

		for _, f := range t.target.FilesystemEntriesRawPlain() {
			if t.dir == "" {
				return fmt.Errorf("task dir not set")
			}
			p := filepath.Join(t.dir, f)
			if p == "/" {
				return fmt.Errorf("root cleanup is not allowed")
			}

			if vb {
				fmt.Printf("  %s ", p)
			}

			err := os.RemoveAll(p)
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

	return nil
}
