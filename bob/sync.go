package bob

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/versionedsync/localsyncstore"
	"github.com/logrusorgru/aurora"
	"os"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/errz"
)

func (b *B) SyncPush(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	_, err = bobfile.BobfileRead(wd)
	errz.Fatal(err)
	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()
	localStore := localsyncstore.New()

	if remoteStore == nil {
		fmt.Println(aurora.Red("No remote project configured can not push"))
	} else {
		for _, sync := range aggregate.SyncCollections {
			err = sync.Push(ctx, *remoteStore, *localStore, aggregate.Dir())
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [collection: %s@%s]", sync.GetName(), sync.Version))

			}
		}
	}

	return nil
}

func (b *B) SyncPull(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	_, err = bobfile.BobfileRead(wd)
	errz.Fatal(err)
	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()
	localStore := localsyncstore.New()

	if remoteStore == nil {
		fmt.Println(aurora.Red("no remote project configured can not pull"))
	} else {
		for _, sync := range aggregate.SyncCollections {
			err = sync.Pull(ctx, *remoteStore, *localStore, aggregate.Dir())
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from remote to local [collection: %s@%s]", sync.GetName(), sync.Version))
			}
		}
	}

	return nil
}

func (b *B) SyncListLocal(_ context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	aggregate, err := bobfile.BobfileRead(wd)
	errz.Fatal(err)

	fmt.Printf("bob sync ls: displaying all files ready to by synced\n\n")

	for _, sync := range aggregate.SyncCollections {
		err = sync.ListLocal(aggregate.Dir())
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed list local [collection: %s@%s]", sync.GetName(), sync.Version))
		}
	}

	return nil
}

func (b *B) SyncListRemote(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	_, err = bobfile.BobfileRead(wd)
	errz.Fatal(err)
	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()

	if remoteStore == nil {
		fmt.Println(aurora.Red("No remote project configured can not list remote"))
	} else {
		// FIXME: list remote should only be run once
		for _, sync := range aggregate.SyncCollections {
			err = sync.ListRemote(ctx, *remoteStore)
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed list remote [collection: %s@%s]", sync.GetName(), sync.Version))
			}
		}
	}

	return nil
}
