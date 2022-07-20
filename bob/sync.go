package bob

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/versionedsync/localsyncstore"
	"github.com/logrusorgru/aurora"
	"os"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/errz"
)

func (b *B) SyncPush(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	aggregate, err := bobfile.BobfileRead(wd)
	aggregate, err = b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()
	localStore := localsyncstore.New()

	if remoteStore == nil {
		fmt.Println(aurora.Red("No remote project configured can not push"))
	} else {
		for _, sync := range aggregate.SyncCollections {
			err = sync.Push(ctx, *remoteStore, *localStore, aggregate.Dir())
			errz.Fatal(err)
		}
	}

	return nil
}

func (b *B) SyncPull(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	aggregate, err := bobfile.BobfileRead(wd)
	aggregate, err = b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()
	localStore := localsyncstore.New()

	if remoteStore == nil {
		fmt.Println(aurora.Red("no remote project configured can not pull"))
	} else {
		for _, sync := range aggregate.SyncCollections {
			err = sync.Pull(ctx, *remoteStore, *localStore, aggregate.Dir())
			errz.Fatal(err)
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
		errz.Fatal(err)
	}

	return nil
}

func (b *B) SyncListRemote(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	aggregate, err := bobfile.BobfileRead(wd)
	errz.Fatal(err)
	aggregate, err = b.Aggregate()
	errz.Fatal(err)

	remoteStore := aggregate.VersionedSyncStore()

	if remoteStore == nil {
		fmt.Println(aurora.Red("No remote project configured can not list remote"))
	} else {
		for _, sync := range aggregate.SyncCollections {
			err = sync.ListRemote(ctx, *remoteStore)
			errz.Fatal(err)
		}
	}

	return nil
}
