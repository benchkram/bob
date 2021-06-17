package bob

import (
	"github.com/Benchkram/errz"
)

func (b *B) Tree(taskname string) (err error) {
	defer errz.Recover(&err)

	// // TODO: if taskname == "" print all

	// aggregate, err := b.Aggregate()
	// errz.Fatal(err)

	// task, ok := aggregate.Tasks[taskname]
	// if !ok {
	// 	return ErrTaskDoesNotExist
	// }

	// keys := make([]string, 0, len(aggregate.Tasks))
	// for k := range aggregate.Tasks {
	// 	keys = append(keys, k)
	// }
	// sort.Strings(keys)

	// for _, k := range keys {
	// 	fmt.Println(k)
	// }

	return nil
}
