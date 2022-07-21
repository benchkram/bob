package bobrun

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/multilinecmd"
	"github.com/benchkram/bob/pkg/nix"
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
// properties from dirty members to plain (e.g. dirtyInit -> init)
func (rm RunMap) Sanitize() (err error) {
	defer errz.Recover(&err)

	for key, task := range rm {
		task.init = multilinecmd.Split(task.InitDirty)
		task.initOnce = multilinecmd.Split(task.InitOnceDirty)
		rm[key] = task
	}

	return nil
}

// CollectNixDependenciesForTasks will collect all nix dependencies for task taskName
// in nixDependencies slice
func (rm RunMap) CollectNixDependenciesForTasks(whitelist []string) ([]nix.Dependency, error) {
	var nixDependencies []nix.Dependency
	for _, taskFromMap := range rm {
		if !taskFromMap.UseNix() {
			continue
		}

		// only add dependencies of whitelisted tasks.
		for _, taskName := range whitelist {
			if taskFromMap.Name() == taskName {
				nixDependencies = append(nixDependencies, taskFromMap.Dependencies()...)
			}
		}
	}

	return nixDependencies, nil
}
