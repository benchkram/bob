package playbook

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/store"
	"github.com/logrusorgru/aurora"
)

// TaskKey is key for context values passed to client for upload/download output formatting
type TaskKey string

func (p *Playbook) downloadArtifact(ctx context.Context, a hash.In, task *bobtask.Task, ignoreLocal bool) {
	if p.enablePull && p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		description := fmt.Sprintf("%-*s\t  %s", p.namePad, task.ColoredName(), aurora.Faint("pulling artifact "+a.String()))
		ctx = context.WithValue(ctx, TaskKey("description"), description)
		syncFromRemoteToLocal(ctx, p.remoteStore, p.localStore, a, task.Name(), ignoreLocal)
	}
}

func (p *Playbook) pushArtifact(ctx context.Context, a hash.In, taskName string) {
	if p.enableCaching && p.remoteStore != nil && p.localStore != nil {
		description := fmt.Sprintf("  %-*s\t%s", p.namePad, taskName, aurora.Faint("pushing artifact "+a.String()))
		ctx = context.WithValue(ctx, TaskKey("description"), description)
		syncFromLocalToRemote(ctx, p.localStore, p.remoteStore, a, taskName, p.namePad)
	}
}

// syncFromRemoteToLocal syncs the artifact from the remote store to the local store.
// if ignoreAlreadyExists is true it will ignore local artifact and perform a fresh download
func syncFromRemoteToLocal(ctx context.Context, remote store.Store, local store.Store, a hash.In, taskName string, ignoreAlreadyExists bool) {
	err := store.Sync(ctx, remote, local, a.String(), ignoreAlreadyExists)
	if errors.Is(err, store.ErrArtifactAlreadyExists) {
		boblog.Log.V(5).Info(fmt.Sprintf("artifact already exists locally [artifactId: %s]. skipping...", a.String()))
	} else if errors.Is(err, store.ErrArtifactNotFoundinSrc) {
		boblog.Log.V(5).Info(fmt.Sprintf("failed to sync from remote to local [artifactId: %s]", a.String()))
	} else if err != nil {
		boblog.Log.V(5).Error(err, fmt.Sprintf("%s: failed to sync from remote to local [artifactId: %s]", taskName, a.String()))
	}

	boblog.Log.V(5).Info(fmt.Sprintf("synced from remote to local [artifactId: %s]", a.String()))
}

// syncFromLocalToRemote syncs the artifact from the local store to the remote store.
func syncFromLocalToRemote(ctx context.Context, local store.Store, remote store.Store, a hash.In, taskName string, namePad int) {
	err := store.Sync(ctx, local, remote, a.String(), false)
	if errors.Is(err, store.ErrArtifactAlreadyExists) {
		boblog.Log.V(5).Info(fmt.Sprintf("artifact already exists on the remote [artifactId: %s]. skipping...", a.String()))
		return
	} else if err != nil {
		boblog.Log.V(5).Error(err, fmt.Sprintf("  %-*s\tfailed to sync from local to remote [artifactId: %s]", namePad, taskName, a.String()))
		return
	}

	// wait for the remote store to finish uploading this artifact. can be moved outside the for loop, but then
	// we don't know which artifacts failed to upload.
	err = remote.Done()
	if err != nil {
		boblog.Log.V(5).Error(err, fmt.Sprintf("failed to sync from local to remote [artifactId: %s]", a.String()))
		return
	}
	boblog.Log.V(5).Info(fmt.Sprintf("synced from local to remote [artifactId: %s]", a.String()))
}
