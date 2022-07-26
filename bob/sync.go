package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bobsync"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/bob/pkg/versionedsync/localsyncstore"
	"github.com/benchkram/bob/pkg/versionedsync/remotesyncstore"
	"github.com/logrusorgru/aurora"
	"os"
	"sort"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/errz"
)

func (b *B) SyncCreatePush(ctx context.Context, collectionName, version, path string, dry bool) (err error) {
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
		// create only sync object and check if exists and is inside bobdir
		// TODO: check if it conflicts with any git tracked file
		sync, err := bobsync.NewSync(collectionName, version, path, aggregate.Dir())
		errz.Fatal(err)
		// check if path is the same as in another sync in bob.yaml
		err = bobsync.CheckForConflicts(aggregate.SyncCollections, *sync)
		if errors.Is(err, bobsync.ErrSyncPathTaken) {
			errz.Fatal(usererror.Wrap(err))
		} else {
			errz.Fatal(err)
		}
		// create collection on remote, if name-version exists on remote => fail
		err = sync.CreateOnRemote(ctx, *remoteStore, dry)
		errz.Fatal(err)
		// TODO: automatically add sync to bob file
		// add it to the bobfile and check for conflicts
		// 		if path exists => replace sync entry in bobfile
		// err = aggregate.AddSync(sync)
		// errz.Fatal(err)
		fmt.Printf("Add that to your bob file so that others can use it:\n"+
			"syncCollections:\n  - name: %s\n    version: %s\n    path: %s\n", sync.Name, sync.Version, sync.Path)

		err = sync.Push(ctx, *remoteStore, *localStore, aggregate.Dir(), dry)
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from local to remote [collection: %s@%s]", sync.Name, sync.Version))

		}
	}

	return nil
}

func (b *B) SyncPull(ctx context.Context, force bool) (err error) {
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
			err = sync.Pull(ctx, *remoteStore, *localStore, aggregate.Dir(), force)
			if err != nil {
				boblog.Log.V(1).Error(err, fmt.Sprintf("failed to sync from remote to local [collection: %s@%s]", sync.Name, sync.Version))
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

	// call validate separate since aggregate is not called
	for _, sync := range aggregate.SyncCollections {
		err = sync.Validate(aggregate.Dir())
		if err != nil {
			errz.Fatal(usererror.Wrapm(err, "can not validate defined sync"))
		}
	}

	for _, sync := range aggregate.SyncCollections {
		err = sync.ListLocal(aggregate.Dir())
		if err != nil {
			boblog.Log.V(1).Error(err, fmt.Sprintf("failed list local [collection: %s@%s]", sync.Name, sync.Version))
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
		for i := range aggregate.SyncCollections {
			err = aggregate.SyncCollections[i].FillRemoteId(ctx, *remoteStore)
			if err != nil {
				if errors.Is(err, remotesyncstore.ErrCollectionNotFound) {
					boblog.Log.V(1).Error(err, fmt.Sprintf("sync collection %s@%s does not exist on the server",
						aggregate.SyncCollections[i].Name, aggregate.SyncCollections[i].Version))
				} else {
					errz.Fatal(err)
				}
			}
		}

		collections, err := (*remoteStore).Collections(ctx)
		errz.Fatal(err)
		for _, c := range collections {
			sort.Sort(c)
			fmt.Printf("%s@%s", aurora.Bold(c.Name), aurora.Italic(c.Version))
			referenced, path := func() (bool, string) {
				for _, sync := range aggregate.SyncCollections {
					if c.ID == sync.GetRemoteId() {
						return true, sync.Path
					}
				}
				return false, ""
			}()
			if referenced {
				fmt.Printf(" [referenced in bob.yaml] (./%s)", path)
			}
			fmt.Println()
			for _, f := range c.Files {
				fmt.Printf("\t%s\n", f.LocalPath)
			}

		}

	}

	return nil
}
