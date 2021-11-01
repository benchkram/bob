package target

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/filehash"
)

// Hash creates a hash for the rntire target
func (t *T) Hash() (empty string, _ error) {
	aggregatedHashes := bytes.NewBuffer([]byte{})
	for _, f := range t.Paths {
		target := filepath.Join(t.dir, f)
		h, err := filehash.Hash(target)
		if err != nil {
			return empty, fmt.Errorf("failed to hash target %q: %w", f, err)
		}

		_, err = aggregatedHashes.Write(h)
		if err != nil {
			return empty, fmt.Errorf("failed to write target hash to aggregated hash %q: %w", f, err)
		}
	}

	h, err := filehash.HashBytes(aggregatedHashes)
	if err != nil {
		return empty, fmt.Errorf("failed to write aggregated target hash: %w", err)
	}

	return hex.EncodeToString(h), nil
}
