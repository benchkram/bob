package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/store"
)

// TaskKey is key for context values passed to client for upload/download output formatting
type TaskKey string

func (p *Playbook) downloadArtifact(ctx context.Context, a hash.In, taskName string) {
	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		ctx = context.WithValue(ctx, TaskKey("taskName"), taskName)
		ctx = context.WithValue(ctx, TaskKey("namePad"), p.namePad)
		syncFromRemoteToLocal(ctx, p.remoteStore, p.localStore, a)
	}
}

func (p *Playbook) pushArtifacts(ctx context.Context, a []hash.In, taskName string) {
	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		ctx = context.WithValue(ctx, TaskKey("taskName"), taskName)
		ctx = context.WithValue(ctx, TaskKey("namePad"), p.namePad)
		syncFromLocalToRemote(ctx, p.localStore, p.remoteStore, a)
	}
}

// syncFromRemoteToLocal syncs the artifact from the remote store to the local store.
func syncFromRemoteToLocal(ctx context.Context, remote store.Store, local store.Store, a hash.In) {
	err := store.Sync(ctx, remote, local, a.String())
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
func syncFromLocalToRemote(ctx context.Context, local store.Store, remote store.Store, artifactIds []hash.In) {
	for _, a := range artifactIds {
		err := store.Sync(ctx, local, remote, a.String())
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
