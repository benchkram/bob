package bobtask

import (
	"fmt"
	"os"
	"path/filepath"
)

// Clean the targets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean() error {
	if t.target != nil {
		fmt.Printf("Cleaning targets of task %s \n", t.name)
		for _, f := range t.target.FilesystemEntriesRawPlain() {
			if t.dir == "" {
				return fmt.Errorf("task dir not set")
			}
			p := filepath.Join(t.dir, f)
			if p == "/" {
				return fmt.Errorf("root cleanup is not allowed")
			}

			fmt.Printf("  %s \n", p)
			err := os.RemoveAll(p)
			if err != nil {
				//fmt.Printf("%s\n", aurora.Red("failed"))
				return err
			}
			//fmt.Printf("%s\n", aurora.Green("done"))
		}
	}

	return nil
}
