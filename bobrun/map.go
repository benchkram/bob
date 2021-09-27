package bobrun

import (
	"bytes"
	"fmt"
	"sort"
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
