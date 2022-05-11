package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/errz"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependencies.
func (b *B) Build(ctx context.Context, taskName string) (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(ag)

	err = BuildNix(ag, taskName)
	errz.Fatal(err)

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	remotestore := ag.Remotestore()

	artifactIds := []hash.In{}
	for _, t := range playbook.Tasks {

		h, err := t.HashIn()
		if err != nil {
			continue
		}

		artifactIds = append(artifactIds, h)
	}

	for _, a := range artifactIds {
		err := store.Sync(ctx, remotestore, b.local, a.String())
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from remote to local [artifactId: %s]", a.String()))
			continue
		}

		boblog.Log.V(1).Info(fmt.Sprintf("synced from remote to local [artifactId: %s]", a.String()))
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	// sync artifacts from current build with remote store
	if remotestore != nil {
		artifactIds := []hash.In{}
		for _, t := range playbook.Tasks {
			if t.TargetExists() {
				h, _ := t.HashIn()
				artifactIds = append(artifactIds, h)
			}
		}

		for _, a := range artifactIds {
			err := store.Sync(ctx, b.local, remotestore, a.String())
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
				continue
			}

			//wait for the remote store to finish uploading this artifact. can be moved outside of the for loop but then
			// we don't know which artifacts failed to upload.
			err = remotestore.Done()
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
				continue
			}

			boblog.Log.V(1).Info(fmt.Sprintf("synced from local to remote [artifactId: %s]", a.String()))
		}
	}

	return nil
}
