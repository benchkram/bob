package bob

import (
	"context"
	"errors"

	"github.com/Benchkram/errz"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependecies.
func (b *B) Build(ctx context.Context, taskname string) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	playbook, err := aggregate.Playbook(taskname)
	errz.Fatal(err)

	return playbook.Build(ctx)
}
