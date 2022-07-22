package bobsync

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/versionedsync/localsyncstore"
	"github.com/benchkram/bob/pkg/versionedsync/remotesyncstore"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"path/filepath"
)

const (
	hashCachePath = ".bob.hashcache"
)

// Sync is a collection of versioned synced files
type Sync struct {
	name string

	Path string `yaml:"path"`

	Version string `yaml:"version"`

	remoteCollectionId string

	cache *HashCache
}

func (s *Sync) GetName() string {
	return s.name
}

func (s *Sync) SetName(name string) {
	s.name = name
}

func (s *Sync) Push(ctx context.Context, remoteStore remotesyncstore.S, localStore localsyncstore.S, bobDir string) (err error) {
	defer errz.Recover(&err)

	//var collectionMustBeCreated bool
	// get collectionId ready

	s.remoteCollectionId, err = remoteStore.CollectionIdByName(ctx, s.name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		//collectionMustBeCreated = true
		s.remoteCollectionId, err = remoteStore.CollectionCreate(ctx, s.name, s.Version, s.Path)
		errz.Fatal(err)
	default:
		errz.Fatal(err)
	}

	absHashCachePath := filepath.Join(bobDir, hashCachePath)
	if s.cache == nil {
		s.cache, err = FromFileOrNew(absHashCachePath)
		errz.Fatal(err)
	}

	// TODO: run list local and remote in parallel
	err = s.cache.Update(filepath.Join(bobDir, s.Path))
	errz.Fatal(err)
	err = s.cache.SaveToFile(absHashCachePath)
	errz.Fatal(err)

	remoteCollection, err := remoteStore.Collection(ctx, s.remoteCollectionId)
	errz.Fatal(err)

	// create the delta
	delta := NewDelta(*s.cache, *remoteCollection)

	fmt.Printf("Local-Remote delta for %s\n", aurora.Bold(s.name))
	fmt.Println(delta.PushOverview())

	// TODO: prompt user and seek confirmation

	for _, f := range delta.LocalFilesMissingOnRemote {
		srcReader, err := localStore.ReadFile(bobDir, s.Path, f.LocalPath)
		errz.Fatal(err)
		err = remoteStore.FileUpload(ctx, s.remoteCollectionId, f.LocalPath, srcReader)
		errz.Fatal(err)
	}
	for _, f := range delta.RemoteFilesMissingOnLocal {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not delete from remote")
		}
		err = remoteStore.FileDelete(ctx, s.remoteCollectionId, *f.ID)
		errz.Fatal(err)
	}
	for _, f := range delta.ToBeUpdated {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not update on remote")
		}
		srcReader, err := localStore.ReadFile(bobDir, s.Path, f.LocalPath)
		errz.Fatal(err)
		err = remoteStore.FileUpdate(ctx, s.remoteCollectionId, *f.ID, srcReader)
		errz.Fatal(err)
	}

	return nil
}

func (s *Sync) Pull(ctx context.Context, remoteStore remotesyncstore.S, localStore localsyncstore.S, bobDir string) (err error) {
	defer errz.Recover(&err)

	//var collectionMustBeCreated bool
	// get collectionId ready
	s.remoteCollectionId, err = remoteStore.CollectionIdByName(ctx, s.name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		//collectionMustBeCreated = true
		s.remoteCollectionId, err = remoteStore.CollectionCreate(ctx, s.name, s.Version, s.Path)
		errz.Fatal(err)
	default:
		errz.Fatal(err)
	}

	absHashCachePath := filepath.Join(bobDir, hashCachePath)
	if s.cache == nil {
		s.cache, err = FromFileOrNew(absHashCachePath)
		errz.Fatal(err)
	}

	// TODO: run list local and remote in parallel
	err = s.cache.Update(filepath.Join(bobDir, s.Path))
	errz.Fatal(err)
	err = s.cache.SaveToFile(absHashCachePath)
	errz.Fatal(err)

	remoteCollection, err := remoteStore.Collection(ctx, s.remoteCollectionId)
	errz.Fatal(err)

	// create the delta
	delta := NewDelta(*s.cache, *remoteCollection)

	fmt.Printf("Local-Remote delta for %s\n", aurora.Bold(s.name))
	fmt.Println(delta.PullOverview())

	// TODO: prompt user and seek confirmation

	for _, f := range delta.LocalFilesMissingOnRemote {
		err := localStore.DeleteFile(bobDir, s.Path, f.LocalPath)
		errz.Fatal(err)
	}
	for _, f := range delta.RemoteFilesMissingOnLocal {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not downlaod from remote")
		}
		srcReader, err := remoteStore.File(ctx, s.remoteCollectionId, *f.ID)
		errz.Fatal(err)
		err = localStore.WriteFile(bobDir, s.Path, f.LocalPath, srcReader)
		errz.Fatal(err)
	}
	for _, f := range delta.ToBeUpdated {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not downlaod from remote")
		}
		srcReader, err := remoteStore.File(ctx, s.remoteCollectionId, *f.ID)
		errz.Fatal(err)
		err = localStore.WriteFile(bobDir, s.Path, f.LocalPath, srcReader)
		errz.Fatal(err)
	}

	return nil
}

func (s *Sync) ListLocal(bobDir string) (err error) {
	defer errz.Recover(&err)

	absHashCachPath := filepath.Join(bobDir, hashCachePath)
	if s.cache == nil {
		s.cache, err = FromFileOrNew(absHashCachPath)
		errz.Fatal(err)
	}

	err = s.cache.Update(s.Path)
	errz.Fatal(err)
	err = s.cache.SaveToFile(absHashCachPath)
	errz.Fatal(err)

	fmt.Printf("%s@%s (./%s)\n", aurora.Bold(s.name), aurora.Italic(s.Version), s.Path)
	for p := range *s.cache {
		fmt.Printf("\t%s\n", p)
	}

	return nil

}

func (s *Sync) ListRemote(ctx context.Context, store remotesyncstore.S) (err error) {
	defer errz.Recover(&err)

	// get collectionId ready
	s.remoteCollectionId, err = store.CollectionIdByName(ctx, s.name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		fmt.Printf("Sync collection %s with version %s does not exist on the server.\n",
			aurora.Bold(s.name), aurora.Bold(s.Version))
	default:
		errz.Fatal(err)
	}
	collections, err := store.Collections(ctx)
	errz.Fatal(err)

	for _, c := range collections {
		fmt.Printf("%s@%s (./%s)", aurora.Bold(c.Name), aurora.Italic(c.Version), c.LocalPath)
		if c.ID == s.remoteCollectionId {
			fmt.Printf(" [synced to local]")
		}
		fmt.Println()
		for _, f := range c.Files {
			fmt.Printf("\t%s\n", f.LocalPath)
		}
	}
	return nil
}
