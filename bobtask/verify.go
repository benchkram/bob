package bobtask

import (
	"fmt"
	"os"
	"strings"

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
			if !isValidFilesystemTarget(path) {
				return usererror.Wrap(fmt.Errorf("invalid target `%s` for task `%s`", path, t.name))
			}
		}
	}

	return nil
}

// isValidFilesystemTarget reports whether the final component of path is "." or ".."
func isValidFilesystemTarget(path string) bool {
	// check for end with one dot
	if path == "." {
		return false
	}

	if strings.Contains(path, "..") {
		return false
	}

	if len(path) >= 2 && path[len(path)-1] == '.' && os.IsPathSeparator(path[len(path)-2]) {
		return false
	}

	return true
}

func (t *Task) verifyAfter() (err error) {
	return nil
}
