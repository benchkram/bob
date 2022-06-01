package bobrun

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/benchkram/bob/pkg/multilinecmd"
	"github.com/benchkram/errz"
)

type RunMap map[string]*Run

func (rm RunMap) String() string {
	description := bytes.NewBufferString("")

	fmt.Fprint(description, "RunMap:\n")

	keys := make([]string, 0, len(rm))
	for k := range rm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		task := rm[k]
		fmt.Fprintf(description, "  %s(%s): -\n", k, task.name)
	}

	return description.String()
}

// Sanitize run map and write filtered & sanitized
// propertys from dirty members to plain (e.g. dirtyInit -> init)
func (rm RunMap) Sanitize() (err error) {
	defer errz.Recover(&err)

	for key, task := range rm {

		sanitizedScript := multilinecmd.Split(task.ScriptDirty)
		task.script = sanitizedScript

		rm[key] = task
	}

	return nil
}
