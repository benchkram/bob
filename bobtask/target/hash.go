package target

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/pkg/filehash"
)

// Hash creates a hash for the entire target
func (t *T) Hash() (empty string, _ error) {
	aggregatedHashes := bytes.NewBuffer([]byte{})
	for _, f := range t.Paths {
		target := filepath.Join(t.dir, f)

		if !file.Exists(target) {
			return empty, fmt.Errorf("target doe not exist %q", f)
		}
		fi, err := os.Stat(target)
		if err != nil {
			return empty, fmt.Errorf("failed to get file info %q: %w", f, err)
		}

		if fi.IsDir() {
			if err := filepath.WalkDir(target, func(p string, fi fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if fi.IsDir() {
					return nil
				}

				h, err := filehash.Hash(p)
				if err != nil {
					return fmt.Errorf("failed to hash target %q: %w", f, err)
				}

				_, err = aggregatedHashes.Write(h)
				if err != nil {
					return fmt.Errorf("failed to write target hash to aggregated hash %q: %w", f, err)
				}

				return nil
			}); err != nil {
				return empty, fmt.Errorf("failed to walk dir %q: %w", target, err)
			}
			// TODO: what happens on a empty dir?
		} else {
			h, err := filehash.Hash(target)
			if err != nil {
				return empty, fmt.Errorf("failed to hash target %q: %w", f, err)
			}

			_, err = aggregatedHashes.Write(h)
			if err != nil {
				return empty, fmt.Errorf("failed to write target hash to aggregated hash %q: %w", f, err)
			}
		}
	}

	h, err := filehash.HashBytes(aggregatedHashes)
	if err != nil {
		return empty, fmt.Errorf("failed to write aggregated target hash: %w", err)
	}

	return hex.EncodeToString(h), nil
}
