package bobtask

import (
	"errors"
	"fmt"

	"github.com/Benchkram/errz"
)

// NeedsRebuild
func (t *Task) NeedsRebuild() (rebuildRequired bool, err error) {
	defer errz.Recover(&err)

	hash, err := t.Hash()
	errz.Fatal(err)

	storedHashes, err := t.ReadHashes()
	if err != nil {
		if errors.Is(err, ErrHashesFileDoesNotExist) {
			return true, nil
		} else if errors.Is(err, ErrTaskHashDoesNotExist) {
			return true, nil
		} else {
			return true, fmt.Errorf("failed to read file hashes: %w", err)
		}
	}

	rebuildRequired = true

	storedHash, ok := storedHashes[t.name]
	if ok {
		rebuildRequired = hash != storedHash
	}

	return rebuildRequired, nil
}
