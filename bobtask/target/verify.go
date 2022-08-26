package target

import (
	"github.com/benchkram/bob/pkg/boblog"
)

// Verify existence and integrity of targets.
// Returns true when no targets defined.
func (t *T) Verify() bool {
	switch t.TypeSerialize {
	case Path:
		return t.verifyFile(t.expectedHash)
	case Docker:
		return t.verifyDocker(t.expectedHash)
	default:
		return t.verifyFile(t.expectedHash)
	}
}

func (t *T) verifyFile(groundTruth string) bool {
	if len(t.PathsSerialize) == 0 {
		return true
	}

	if t.expectedHash == "" {
		return true
	}

	// check plain existence
	if !t.existsFile() {
		return false
	}

	// check integrity by comparing hash
	hash, err := t.Hash()
	if err != nil {
		boblog.Log.Error(err, "Unable to create target hash")
		return false
	}

	return hash == groundTruth
}

func (t *T) verifyDocker(groundTruth string) bool {

	if len(t.PathsSerialize) == 0 {
		return true
	}

	if t.expectedHash == "" {
		return true
	}

	// check plain existence
	if !t.existsDocker() {
		return false
	}

	// check integrity by comparing hash
	hash, err := t.Hash()
	if err != nil {
		boblog.Log.Error(err, "Unable to verify target docker image hash")
		return false
	}

	return hash == groundTruth
}
