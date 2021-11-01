package hash

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/Benchkram/bob/pkg/filehash"
)

type Hashes map[string]Task
type Targets map[string]string
type H struct {
	Path string
	Hash string
}

// Task is a context aware object to
// handle hashes of a task
type Task struct {
	// Input holds a hash  related to the input
	// files, task description and environment
	Input string

	// Targets hold hash values on all related
	// targets in the build chain.
	Targets Targets
}

func HashFiles(files []string) []H {
	fhs := make([]H, 0, len(files))
	for _, f := range files {
		h, err := filehash.Hash(f)
		if err != nil {
			log.Printf("failed to hash file %q: %v\n", f, err)
			continue
		}

		fhs = append(fhs, H{
			Path: f,
			Hash: hex.EncodeToString(h),
		})
	}
	return fhs
}

func FileHashesDiffer(fhs1, fhs2 []H) error {
	for _, fh1 := range fhs1 {
		var found bool
		for _, fh2 := range fhs2 {
			if fh1.Path == fh2.Path {
				found = true
				if fh1.Hash != fh2.Hash {
					return fmt.Errorf("hashes of file %q differ: %q != %q", fh1.Path, fh1.Hash, fh2.Hash)
				}
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in second slice", fh1.Path)
		}
	}

	for _, fh2 := range fhs2 {
		var found bool
		for _, fh1 := range fhs1 {
			if fh1.Path == fh2.Path {
				found = true
				if fh1.Hash != fh2.Hash {
					return fmt.Errorf("hashes of file %q differ: %q != %q", fh2.Path, fh2.Hash, fh1.Hash)
				}
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in first slice", fh2.Path)
		}
	}

	return nil
}
