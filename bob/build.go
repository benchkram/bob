package bob

import (
	"context"
	"errors"

	"github.com/Benchkram/errz"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

func (b *B) Build(ctx context.Context, taskname string) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)
	println(aggregate.Tasks.String())
	println(aggregate.Runs.String())

	playbook, err := aggregate.Playbook(taskname)
	errz.Fatal(err)
	println(playbook.String())

	return playbook.Build(ctx)
}
