package bobtask

import (
	"fmt"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/file"
)

// Verify a bobfile before task runner.
func (t *Task) Verify() error {
	return t.verifyBefore()
}

// VerifyBefore a bobfile before task runner.
func (t *Task) VerifyBefore() error {
	return t.verifyBefore()
}

// VerifyAfter a bobfile after task runner.
func (t *Task) VerifyAfter() error {
	return t.verifyAfter()
}

func (t *Task) verifyBefore() (err error) {
	// A task without commands always needs valid exports
	if len(t.cmds) == 0 {
		for _, export := range t.Exports {
			target := filepath.Join(t.dir, string(export))
			if !file.Exists(target) {
				return fmt.Errorf("%s exports %s but the path does not exist.", t.name, export)
			}
		}
	}
	return nil
}

func (t *Task) verifyAfter() (err error) {
	// Verify all exports
	for _, export := range t.Exports {
		target := filepath.Join(t.dir, string(export))
		if !file.Exists(target) {
			return fmt.Errorf("%s exports %s but the path does not exist.", t.name, export)
		}
	}

	return nil
}
