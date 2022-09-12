package bobtask

import (
	"fmt"
	"path/filepath"
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

// isValidFilesystemTarget checks if target is a valid path inside
// the given bob.yaml context.
// Paths resolving to `.` or starting with `/` are considered invalid
func isValidFilesystemTarget(path string) bool {

	cleaned := filepath.Clean(path)

	// the current directory can't be a target.
	if cleaned == "." {
		return false
	}

	// do not leave the context of a directory containing the bob.yaml file.
	if strings.HasPrefix(cleaned, "..") {
		return false
	}

	// root is not allowed as a target.
	if strings.HasPrefix(cleaned, "/") {
		return false
	}

	return true
}

func (t *Task) verifyAfter() (err error) {
	return nil
}
