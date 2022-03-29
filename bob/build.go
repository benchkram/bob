package bob

import (
	"context"
	"errors"

	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependecies.
func (b *B) Build(ctx context.Context, taskname string) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(aggregate)

	playbook, err := aggregate.Playbook(
		taskname,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
