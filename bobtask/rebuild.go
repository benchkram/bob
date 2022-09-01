package bobtask

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/errz"
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

	if hashIn == nil {
		*hashIn, err = t.HashIn()
		errz.Fatal(err)
	}

	if !t.buildInfoStore.BuildInfoExists(hashIn.String()) {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			boblog.Log.V(4).Info(fmt.Sprintf("%s, Searching for input hash %s failed", t.name, hashIn.String()))
			return true, nil
		}
		errz.Fatal(err)
	}
	boblog.Log.V(4).Info(fmt.Sprintf("%s, Searching for input hash %s succeeded", t.name, hashIn.String()))

	return false, nil
}
