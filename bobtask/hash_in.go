package bobtask

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/filehash"
)

// HashIn computes a hash containing inputs, environment and the task description.
func (t *Task) HashIn() (taskHash hash.In, err error) {
	if t.hashIn != nil {
		return *t.hashIn, nil
	}

	h := filehash.New()

	// Hash input files
	for _, f := range t.inputs {
		err = h.AddFile(f)
		if err != nil {
			if errors.Is(err, os.ErrPermission) {
				t.addToSkippedInputs(f)
				continue
			} else {
				return taskHash, fmt.Errorf("failed to hash file %q: %w", f, err)
			}
		}
	}

	// Hash the public task description
	description, err := yaml.Marshal(t)
	if err != nil {
		return taskHash, fmt.Errorf("failed to marshal task: %w", err)
	}
	err = h.AddBytes(bytes.NewBuffer(description))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}

	// Hash the project name
	err = h.AddBytes(bytes.NewBuffer([]byte(t.project)))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write project name hash: %w", err)
	}

	// Hash the environment
	sort.Strings(t.env)
	environment := strings.Join(t.env, ",")
	err = h.AddBytes(bytes.NewBufferString(environment))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}

	// Hash store paths
	err = h.AddBytes(bytes.NewBufferString(strings.Join(t.storePaths, "")))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write store paths hash: %w", err)
	}

	hashIn := hash.In(hex.EncodeToString(h.Sum()))

	// store hash for reuse
	t.hashIn = &hashIn

	return hashIn, nil
}
