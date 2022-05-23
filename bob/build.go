package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/store"
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

	fmt.Println("Building nix dependencies...")
	err = BuildNixDependenciesInPipeline(ag, taskName)
	errz.Fatal(err)
	fmt.Println("Succeded building nix dependencies")

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	// HINT: It's not easily possible to parallelize hash computation
	// with building nix dependecies.. as the storePaths computed by
	// BuildNixDependenciesInPipeline are considered in the task input hash.
	err = playbook.PreComputeInputHashes()
	errz.Fatal(err)

	remote := ag.Remotestore()

	if b.enableCaching && remote != nil {
		// populate the local cache with any pre-compiled artifacts
		syncFromRemoteToLocal(ctx, remote, b.local, getArtifactIds(playbook, false))
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	if b.enableCaching && remote != nil {
		// sync any newly generated artifacts with the remote store
		syncFromLocalToRemote(ctx, b.local, remote, getArtifactIds(playbook, true))
	}

	return nil
}

// getArtifactIds returns the artifact ids of the given playbook (and optionally checks if the target exists first)
func getArtifactIds(pbook *playbook.Playbook, checkForTarget bool) []hash.In {
	artifactIds := []hash.In{}
	for _, t := range pbook.Tasks {
		if checkForTarget && !t.TargetExists() {
			continue
		}

		h, err := t.HashIn()
		if err != nil {
			continue
		}

		artifactIds = append(artifactIds, h)
	}
	return artifactIds
}

// syncFromRemoteToLocal syncs the artifacts from the remote store to the local store.
func syncFromRemoteToLocal(ctx context.Context, remote store.Store, local store.Store, artifactIds []hash.In) {
	for _, a := range artifactIds {
		err := store.Sync(ctx, remote, local, a.String())
		if errors.Is(err, store.ErrArtifactAlreadyExists) {
			boblog.Log.V(1).Info(fmt.Sprintf("artifact already exists locally [artifactId: %s]. skipping...", a.String()))
			continue
		} else if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from remote to local [artifactId: %s]", a.String()))
			continue
		}

		boblog.Log.V(1).Info(fmt.Sprintf("synced from remote to local [artifactId: %s]", a.String()))
	}
}

// syncFromLocalToRemote syncs the artifacts from the local store to the remote store.
func syncFromLocalToRemote(ctx context.Context, local store.Store, remote store.Store, artifactIds []hash.In) {
	for _, a := range artifactIds {
		err := store.Sync(ctx, local, remote, a.String())
		if errors.Is(err, store.ErrArtifactAlreadyExists) {
			boblog.Log.V(1).Info(fmt.Sprintf("artifact already exists on the remote [artifactId: %s]. skipping...", a.String()))
			continue
		} else if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		// wait for the remote store to finish uploading this artifact. can be moved outside of the for loop but then
		// we don't know which artifacts failed to upload.
		err = remote.Done()
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		boblog.Log.V(1).Info(fmt.Sprintf("synced from local to remote [artifactId: %s]", a.String()))
	}
}
