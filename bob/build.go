package bob

import (
	"context"
	"errors"

	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/store"
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

	// sync artifacts from current build with remote store
	remotestore := aggregate.Remotestore()
	if remotestore != nil {
		artifactIds := []hash.In{}
		for _, t := range playbook.Tasks {
			if t.TargetExists() {
				h, _ := t.HashIn()
				artifactIds = append(artifactIds, h)
			}
		}
		for _, a := range artifactIds {
			err = store.Sync(ctx, b.local, remotestore, a.String())
			errz.Fatal(err)
		}

		// wait for the remote store to finish all processings
		remotestore.Done()
	}

	return err
}
