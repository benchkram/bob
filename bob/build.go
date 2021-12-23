package bob

import (
	"context"
	"errors"

	"github.com/Benchkram/bob/bob/playbook"
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

	b.PrintVersionCompatibility(aggregate)

	playbook, err := aggregate.Playbook(
		taskname,
		playbook.WithArtifactsDisabled(b.disableCache),
	)
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
