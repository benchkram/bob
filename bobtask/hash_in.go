package bobtask

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/filehash"
	"gopkg.in/yaml.v2"
)

var hashCacheMutex = sync.Mutex{}
var hashCache = make(map[string][]byte, 10000)

// HashIn computes a hash containing inputs, environment and the task description.
func (t *Task) HashIn() (taskHash hash.In, _ error) {
	if t.hashIn != nil {
		return *t.hashIn, nil
	}

	aggregatedHashes := bytes.NewBuffer([]byte{})

	// Hash input files
	for _, f := range t.inputs {
		var h []byte
		hashCacheMutex.Lock()
		hash, ok := hashCache[f]
		hashCacheMutex.Unlock()
		if ok {
			// reuse hash
			h = hash
		} else {
			// recompute hash
			var errr error
			h, errr = filehash.Hash(f)
			if errr != nil {
				if errors.Is(errr, os.ErrPermission) {
					t.addToSkippedInputs(f)
					continue
				} else {
					return taskHash, fmt.Errorf("failed to hash file %q: %w", f, errr)
				}
			}
			hashCacheMutex.Lock()
			hashCache[f] = h
			hashCacheMutex.Unlock()
		}

		_, err := aggregatedHashes.Write(h)
		if err != nil {
			return taskHash, fmt.Errorf("failed to write file hash to aggregated hash %q: %w", f, err)
		}
	}

	// Hash the public task description
	description, err := yaml.Marshal(t)
	if err != nil {
		return taskHash, fmt.Errorf("failed to marshal task: %w", err)
	}
	descriptionHash, err := filehash.HashBytes(bytes.NewBuffer(description))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(descriptionHash)
	if err != nil {
		return taskHash, fmt.Errorf("failed to write task description to aggregated hash: %w", err)
	}

	// Hash the project name
	projectNameHash, err := filehash.HashBytes(bytes.NewBuffer([]byte(t.project)))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write project name hash: %w", err)
	}
	_, err = aggregatedHashes.Write(projectNameHash)
	if err != nil {
		return taskHash, fmt.Errorf("failed to write task description to aggregated hash: %w", err)
	}

	// Hash the environment
	sort.Strings(t.env)
	environment := strings.Join(t.env, ",")
	environmentHash, err := filehash.HashBytes(bytes.NewBufferString(environment))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}
	_, err = aggregatedHashes.Write(environmentHash)
	if err != nil {
		return taskHash, fmt.Errorf("failed to write task environment to aggregated hash: %w", err)
	}

	// Summarize
	h, err := filehash.HashBytes(aggregatedHashes)
	if err != nil {
		return taskHash, fmt.Errorf("failed to write aggregated hash: %w", err)
	}

	hashIn := hash.In(hex.EncodeToString(h))

	// store hash for reuse
	t.hashIn = &hashIn

	return hashIn, nil
}
