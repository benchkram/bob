package target

import "github.com/Benchkram/errz"

// Verify existence and integrity of targets.
// Returns true when no targets defined.
func (t *T) Verify() bool {
	switch t.Type {
	case File:
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
		errz.Log(err)
		return false
	}

	return hash == groundTruth
}

func (t *T) verifyDocker(groundTruth string) bool {
	return true
}
