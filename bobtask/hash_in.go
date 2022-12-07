package bobtask

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/bob/pkg/sliceutil"
)

// HashInAlways computes the input hash without using a cached value
func (t *Task) HashInAlways() (taskHash hash.In, err error) {
	return t.computeInputHash()
}

// HashIn computes the input hash
func (t *Task) HashIn() (taskHash hash.In, err error) {
	if t.hashIn != nil {
		boblog.Log.V(4).Info(fmt.Sprintf("Reusing hash for task %s, using %d input files ", t.Name(), len(t.inputs)))
		return *t.hashIn, nil
	}
	return t.computeInputHash()
}

// computeInputHash computes a hash containing inputs, environment and the task description.
func (t *Task) computeInputHash() (taskHash hash.In, err error) {
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
	err = h.AddBytes(strings.NewReader(t.Description()))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}

	// Hash the project name
	err = h.AddBytes(bytes.NewBuffer([]byte(t.project)))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write project name hash: %w", err)
	}

	// Hash the environment
	env := filterEnvOfIgnores(t.env)
	sort.Strings(env)
	environment := strings.Join(env, ",")
	err = h.AddBytes(bytes.NewBufferString(environment))
	if err != nil {
		return taskHash, fmt.Errorf("failed to write description hash: %w", err)
	}

	hashIn := hash.In(hex.EncodeToString(h.Sum()))

	// store hash for reuse
	t.hashIn = &hashIn

	boblog.Log.V(4).Info(fmt.Sprintf("Computed hash [h: %s] for task [t: %s], using [inputs:%d] input files ", t.hashIn.String(), t.Name(), len(t.inputs)))

	return hashIn, nil
}

func filterEnvOfIgnores(env []string) []string {
	ignore := []string{"buildCommandPath", "SHLVL"}

	var result []string
	for _, v := range env {
		pair := strings.SplitN(v, "=", 2)
		if sliceutil.Contains(ignore, pair[0]) {
			continue
		}
		result = append(result, v)
	}
	return result
}
