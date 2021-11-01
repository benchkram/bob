package bobtask

import (
	"errors"
	"fmt"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/errz"
)

type RebuildOptions struct {
	Hash *hash.Task
}

// NeedsRebuild
func (t *Task) NeedsRebuild(options *RebuildOptions) (rebuildRequired bool, err error) {
	defer errz.Recover(&err)

	var hash *hash.Task
	if options != nil {
		if options.Hash != nil {
			hash = options.Hash
		}
	}

	if hash == nil {
		hash, err = t.Hash()
		errz.Fatal(err)
	}

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
		rebuildRequired = hash.Input != storedHash.Input
	}

	return rebuildRequired, nil
}
