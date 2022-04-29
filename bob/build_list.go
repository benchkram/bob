package bob

import (
	"fmt"
	"sort"

	"github.com/benchkram/errz"
)

func (b *B) List() (err error) {
	defer errz.Recover(&err)

	keys, err := b.GetList()
	errz.Fatal(err)

	for _, k := range keys {
		fmt.Println(k)
	}

	return nil
}

func (b *B) GetList() (tasks []string, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.AggregateSparse()
	errz.Fatal(err)

	keys := make([]string, 0, len(aggregate.BTasks))
	for _, task := range aggregate.BTasks {
		keys = append(keys, task.Name())
	}
	sort.Strings(keys)

	return keys, nil
}
