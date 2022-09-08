package bobtask

import (
	"fmt"
	"os"

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
	if t.target != nil {
		for _, path := range t.target.FilesystemEntriesRawPlain() {
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
	return nil
}
