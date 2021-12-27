package bobtask

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/pkg/filehash"
	"gopkg.in/yaml.v2"
)

// HashIn computes a hash containing inputs, environment and the task description.
func (t *Task) HashIn() (taskHash hash.In, _ error) {
	if t.hashIn != nil {
		return *t.hashIn, nil
	}

	aggregatedHashes := bytes.NewBuffer([]byte{})

	// Hash input files
	for _, f := range t.inputs {
		h, err := filehash.Hash(f)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				t.addToSkippedInputs(f)
				continue
			} else {
				return taskHash, fmt.Errorf("failed to hash file %q: %w", f, err)
			}
		}

		_, err = aggregatedHashes.Write(h)
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
