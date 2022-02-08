package bobtask

import (
	"errors"
	"fmt"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/bob/pkg/buildinfostore"
	"github.com/Benchkram/bob/pkg/dockermoby"
	"github.com/Benchkram/errz"
)

type RebuildOptions struct {
	HashIn *hash.In
}

// NeedsRebuild returns true if the `In` hash does not exist in the hash storage
func (t *Task) NeedsRebuild(options *RebuildOptions) (_ bool, err error) {
	defer errz.Recover(&err)

	var hashIn *hash.In
	if options != nil {
		if options.HashIn != nil {
			hashIn = options.HashIn
		}
	}

	if t.target.Type == target.Docker {
		registry, err := dockermoby.New()
		errz.Fatal(err)

		fmt.Println(t.target.Paths[0])
		imHash, err := registry.FetchImageHash(t.target.Paths[0])
		errz.Fatal(err)

		fmt.Println(imHash)

		// fmt.Println(imHash)
		// _, err = registry.SaveImage(imHash, t.target.Paths[0])
		// errz.Fatal(err)
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
