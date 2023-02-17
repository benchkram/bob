package bobtask

import (
	"fmt"
	"strings"
)

// sanitizeInputs assures that inputs are only cosidered when they are inside the project dir.
// Needs to be called with the current working directory set to the tasks working directory.
func (t *Task) sanitizeInput(f string) error {

	if strings.Contains(f, "../") {
		return fmt.Errorf("'../' not allowed in file path %q", f)
	}

	if strings.HasPrefix(f, "/") {
		return fmt.Errorf("'/' not allowed, use only inputs relative to the project root %q", f)
	}

	return nil
}

// sanitizeRebuild used to transform from dirty member to internal member
func (t *Task) sanitizeRebuild(s string) RebuildType {
	switch strings.ToLower(s) {
	case string(RebuildAlways):
		return RebuildAlways
	case string(RebuildOnChange):
		return RebuildOnChange
	default:
		return RebuildOnChange
	}
}
