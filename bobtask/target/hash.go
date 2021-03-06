package target

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/bob/pkg/usererror"

	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
)

// Hash creates a hash for the entire target
func (t *T) Hash() (empty string, _ error) {
	switch t.Type {
	case Path:
		return t.filepathHash()
	case Docker:
		return t.dockerImagesHash()
	default:
		return t.filepathHash()
	}
}

func (t *T) filepathHash() (empty string, _ error) {
	aggregatedHashes := bytes.NewBuffer([]byte{})
	for _, f := range t.Paths {
		target := filepath.Join(t.dir, f)

		if !file.Exists(target) {
			return empty, usererror.Wrapm(fmt.Errorf("target does not exist %q", f), "failed to hash target")
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

func (t *T) dockerImagesHash() (string, error) {

	var hash string

	for _, image := range t.Paths {
		h, err := t.dockerRegistryClient.ImageHash(image)
		if err != nil {
			if errors.Is(err, dockermobyutil.ErrImageNotFound) {
				return "", usererror.Wrapm(err, "failed to fetch docker image hash")
			} else {
				return "", fmt.Errorf("failed to get docker image hash info %q: %w", image, err)
			}
		}
		hash = hash + h

	}

	return hash, nil
}
