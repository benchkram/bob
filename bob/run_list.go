package bob

import (
	"sort"

	"github.com/benchkram/errz"
)

func (b *B) GetRunTasks() (tasks []string, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.AggregateSparse()
	errz.Fatal(err)

	keys := make([]string, 0, len(aggregate.RTasks))
	for _, task := range aggregate.RTasks {
		keys = append(keys, task.Name())
	}
	sort.Strings(keys)

	return keys, nil
}
