package bob

import (
	"fmt"
	"sort"

	"github.com/benchkram/errz"
)

func (b *B) RunList() (err error) {
	defer errz.Recover(&err)

	keys, err := b.GetRunList()
	errz.Fatal(err)

	for _, k := range keys {
		fmt.Println(k)
	}

	return nil
}

func (b *B) GetRunList() (tasks []string, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	keys := make([]string, 0, len(aggregate.RTasks))
	for _, task := range aggregate.RTasks {
		keys = append(keys, task.Name())
	}
	sort.Strings(keys)

	return keys, nil
}
