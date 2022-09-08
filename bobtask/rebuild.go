package bobtask

import (
	"fmt"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

type RebuildOptions struct {
	HashIn *hash.In
}

// didTaskChange returns true if the `In` hash does not exist in the buildinfo storage.
// This indicates that there has not been a previous build for this combination of inputs.
func (t *Task) DidTaskChange() (_ bool, err error) {
	defer errz.Recover(&err)

	hashIn, err := t.HashIn()
	errz.Fatal(err)

	if t.buildInfoStore.BuildInfoExists(hashIn.String()) {
		boblog.Log.V(4).Info(fmt.Sprintf("%s, Searching for input hash %s succeeded", t.name, hashIn.String()))
		return false, nil
	}

	boblog.Log.V(4).Info(fmt.Sprintf("%s, Searching for input hash %s failed", t.name, hashIn.String()))
	return true, nil
}
