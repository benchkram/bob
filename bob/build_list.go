package bob

import (
	"fmt"
	"sort"

	"github.com/Benchkram/errz"
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

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	keys := make([]string, 0, len(aggregate.BTasks))
	for k := range aggregate.BTasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, nil
}
