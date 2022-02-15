package target

import "github.com/Benchkram/bob/pkg/boblog"

// Verify existence and integrity of targets.
// Returns true when no targets defined.
func (t *T) Verify() bool {
	switch t.Type {
	case Path:
		return t.verifyFile(t.hash)
	case Docker:
		return t.verifyDocker(t.hash)
	default:
		return t.verifyFile(t.hash)
	}
}

func (t *T) verifyFile(groundTruth string) bool {
	if len(t.Paths) == 0 {
		return true
	}

	if t.hash == "" {
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
	if len(t.Paths) == 0 {
		return true
	}

	if t.hash == "" {
		return true
	}

	// check plain existence
	if !t.existsDocker() {
		return false
	}

	hash, err := t.Hash()
	if err != nil {
		boblog.Log.Error(err, "Unable to verify target docker image hash")
		return false
	}

	return groundTruth == hash
}
