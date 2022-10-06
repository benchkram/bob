package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/store"
)

func (p *Playbook) downloadArtifact(ctx context.Context, a hash.In, taskName string) {
	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		syncFromRemoteToLocal(ctx, p.remoteStore, p.localStore, a, fmt.Sprintf("Download artifact of task: %s", taskName))
	}
}

func (p *Playbook) pushArtifacts(ctx context.Context, a []hash.In, taskName string) {
	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		syncFromLocalToRemote(ctx, p.localStore, p.remoteStore, a, fmt.Sprintf("Upload artifact of task: %s", taskName))
	}
}

// syncFromRemoteToLocal syncs the artifact from the remote store to the local store.
func syncFromRemoteToLocal(ctx context.Context, remote store.Store, local store.Store, a hash.In, msg string) {
	err := store.Sync(ctx, remote, local, a.String(), msg)
	if errors.Is(err, store.ErrArtifactAlreadyExists) {
		boblog.Log.V(5).Info(fmt.Sprintf("artifact already exists locally [artifactId: %s]. skipping...", a.String()))
	} else if errors.Is(err, store.ErrArtifactNotFoundinSrc) {
		boblog.Log.V(5).Info(fmt.Sprintf("failed to sync from remote to local [artifactId: %s]", a.String()))
	} else if err != nil {
		boblog.Log.V(5).Error(err, fmt.Sprintf("failed to sync from remote to local [artifactId: %s]", a.String()))
	}

	boblog.Log.V(5).Info(fmt.Sprintf("synced from remote to local [artifactId: %s]", a.String()))
}

// syncFromLocalToRemote syncs the artifacts from the local store to the remote store.
func syncFromLocalToRemote(ctx context.Context, local store.Store, remote store.Store, artifactIds []hash.In, msg string) {
	for _, a := range artifactIds {
		err := store.Sync(ctx, local, remote, a.String(), msg)
		if errors.Is(err, store.ErrArtifactAlreadyExists) {
			boblog.Log.V(5).Info(fmt.Sprintf("artifact already exists on the remote [artifactId: %s]. skipping...", a.String()))
			continue
		} else if err != nil {
			boblog.Log.V(5).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		// wait for the remote store to finish uploading this artifact. can be moved outside the for loop, but then
		// we don't know which artifacts failed to upload.
		err = remote.Done()
		if err != nil {
			boblog.Log.V(5).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
			continue
		}

		boblog.Log.V(5).Info(fmt.Sprintf("synced from local to remote [artifactId: %s]", a.String()))
	}
}
