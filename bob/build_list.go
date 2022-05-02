package bob

import (
	"sort"

	"github.com/benchkram/errz"
)

func (b *B) GetBuildTasks() (tasks []string, err error) {
	defer errz.Recover(&err)

	omitRunTasks := true
	aggregate, err := b.AggregateSparse(omitRunTasks)
	errz.Fatal(err)

	keys := make([]string, 0, len(aggregate.BTasks))
	for _, task := range aggregate.BTasks {
		keys = append(keys, task.Name())
	}
	sort.Strings(keys)

	return keys, nil
}
