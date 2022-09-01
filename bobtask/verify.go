package bobtask

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/usererror"
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

	if t.target != nil && t.target.Type() == target.Path {
		for _, path := range t.target.Paths() {
			if verifyTargetPath(path) {
				return usererror.Wrap(fmt.Errorf("invalid target `%s` for task `%s`", path, t.name))
			}
		}
	}

	return nil
}

// verifyTargetPath reports whether the final component of path is "." or ".."
func verifyTargetPath(path string) bool {
	// check for end with one dot
	if path == "." {
		return true
	}
	if len(path) >= 2 && path[len(path)-1] == '.' && os.IsPathSeparator(path[len(path)-2]) {
		return true
	}

	// check for end with two dots
	if len(path) >= 2 && path[len(path)-1] == '.' && path[len(path)-2] == '.' {
		return true
	}

	return false
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
