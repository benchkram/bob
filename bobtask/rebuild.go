package bobtask

import (
	"errors"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/pkg/buildinfostore"
	"github.com/Benchkram/errz"
)

type RebuildOptions struct {
	HashIn *hash.In
}

// NeedsRebuild returns true if the `In` hash does not exist in the hash storage
func (t *Task) NeedsRebuild(options *RebuildOptions) (_ bool, err error) {
	defer errz.Recover(&err)

	// returns true if rebuild strategy set to `always`
	if t.rebuild == ALWAYS {
		return true, nil
	}

	var hashIn *hash.In
	if options != nil {
		if options.HashIn != nil {
			hashIn = options.HashIn
		}
	}

	if hashIn == nil {
		*hashIn, err = t.HashIn()
		errz.Fatal(err)
	}

	_, err = t.buildInfoStore.GetBuildInfo(hashIn.String())
	if err != nil {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			return true, nil
		}
		errz.Fatal(err)
	}

	// storedHashes, err := t.ReadHashes()
	// if err != nil {
	// 	if errors.Is(err, ErrHashesFileDoesNotExist) {
	// 		return true, nil
	// 	} else if errors.Is(err, ErrTaskHashDoesNotExist) {
	// 		return true, nil
	// 	} else {
	// 		return true, fmt.Errorf("failed to read file hashes: %w", err)
	// 	}
	// }

	// _, ok := storedHashes[*hashIn]
	// return !ok, nil

	return false, nil
}
