package bobsync

import (
	"context"
	"fmt"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/bob/pkg/userprompt"
	"github.com/benchkram/bob/pkg/versionedsync/collection"
	"github.com/benchkram/bob/pkg/versionedsync/localsyncstore"
	"github.com/benchkram/bob/pkg/versionedsync/remotesyncstore"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"path/filepath"
	"sort"
	"strings"
)

const (
	hashCachePath = ".bob.hashcache"
)

// Sync is a collection of versioned synced files
type Sync struct {
	Name string `yaml:"name"`

	Path string `yaml:"path"`

	Version string `yaml:"version"`

	remoteCollectionId string

	cache *HashCache
}

func NewSync(collectionName, version, path, bobDir string) (s *Sync, err error) {
	s = &Sync{
		Name:    collectionName,
		Version: version,
		Path:    path,
	}
	return s, s.Validate(bobDir)
}

func (s *Sync) GetRemoteId() string {
	return s.remoteCollectionId
}

// Validate checks if the Sync definition is ok
// path must be inside the bobDir.
func (s *Sync) Validate(bobDir string) (err error) {
	defer errz.Recover(&err)
	if s.Path == "" {
		return usererror.Wrapm(fmt.Errorf("path is empty"), "invalid collection path")
	}
	// fileInfo, err := os.Stat(filepath.Join(bobDir, s.Path))
	// if err != nil {
	//  	return usererror.Wrapm(err, "invalid collection path")
	//}
	//if !fileInfo.IsDir() {
	//	return usererror.Wrapm(fmt.Errorf("%s is not a directory", s.Path), "invalid collection path")
	//}
	rel, err := filepath.Rel(bobDir, filepath.Join(bobDir, s.Path))
	errz.Fatal(err)
	if strings.HasPrefix(rel, "..") {
		return usererror.Wrapm(fmt.Errorf("%s is not in the bob directory", s.Path), "invalid collection path")
	}

	if strings.Contains(s.Name, collection.Divider) {
		return usererror.Wrapm(ErrInvalidCollectionName, fmt.Sprintf("can not contain \"%s\"", collection.Divider))
	}
	if strings.Contains(s.Version, collection.Divider) {
		return usererror.Wrapm(ErrInvalidCollectionVersion, fmt.Sprintf("can not contain \"%s\"", collection.Divider))
	}

	return nil
}

func (s *Sync) CreateOnRemote(ctx context.Context, remoteStore remotesyncstore.S, dry bool) (err error) {
	defer errz.Recover(&err)

	_, err = remoteStore.CollectionIdByName(ctx, s.Name, s.Version)
	switch err {
	case nil:
		return usererror.Wrapm(ErrCollectionVersionExists, fmt.Sprintf("%s@%s can not be created on remote", s.Name, s.Version))
	case remotesyncstore.ErrCollectionNotFound:
	default:
		errz.Fatal(err)
	}

	if dry {
		return nil
	}
	s.remoteCollectionId, err = remoteStore.CollectionCreate(ctx, s.Name, s.Version)
	errz.Fatal(err)

	return nil
}

func (s *Sync) Push(ctx context.Context, remoteStore remotesyncstore.S, localStore localsyncstore.S, bobDir string, dry bool) (err error) {
	defer errz.Recover(&err)

	//var collectionMustBeCreated bool
	// get collectionId ready

	s.remoteCollectionId, err = remoteStore.CollectionIdByName(ctx, s.Name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		//collectionMustBeCreated = true
		s.remoteCollectionId, err = remoteStore.CollectionCreate(ctx, s.Name, s.Version)
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
	// TODO: no symlinks allowed
	// TODO: collection root and all underneath is in .gitignore [could be allowed with a warning]
	err = s.cache.Update(filepath.Join(bobDir, s.Path))
	errz.Fatal(err)

	remoteCollection, err := remoteStore.Collection(ctx, s.remoteCollectionId)
	errz.Fatal(err)

	// create the delta
	delta := NewDelta(*s.cache, *remoteCollection)

	if !dry {
		fmt.Printf("Creating collection %s@%s (%s) on remote\n", aurora.Bold(s.Name), aurora.Italic(s.Version), s.Path)
	} else {
		fmt.Printf("Simulated collection %s@%s (%s) on remote\n", aurora.Bold(s.Name), aurora.Italic(s.Version), s.Path)
	}
	fmt.Println(delta.PushOverview())

	if dry {
		return nil
	}

	for _, f := range delta.LocalFilesMissingOnRemote {
		if !f.IsDirectory {
			srcReader, err := localStore.ReadFile(filepath.Join(bobDir, s.Path, f.LocalPath))
			errz.Fatal(err)
			err = remoteStore.FileUpload(ctx, s.remoteCollectionId, f.LocalPath, srcReader)
			errz.Fatal(err)
		} else {
			err = remoteStore.MakeDir(ctx, s.remoteCollectionId, f.LocalPath)
			errz.Fatal(err)
		}
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
		if !f.IsDirectory {
			srcReader, err := localStore.ReadFile(filepath.Join(bobDir, s.Path, f.LocalPath))
			errz.Fatal(err)
			err = remoteStore.FileUpdate(ctx, s.remoteCollectionId, *f.ID, false, srcReader)
			errz.Fatal(err)
		} else {
			err = remoteStore.FileUpdate(ctx, s.remoteCollectionId, *f.ID, true, nil)
			errz.Fatal(err)
		}
	}

	err = s.cache.SaveToFile(absHashCachePath)
	errz.Fatal(err)

	return nil
}

func (s *Sync) Pull(ctx context.Context, remoteStore remotesyncstore.S, localStore localsyncstore.S, bobDir string, force bool) (err error) {
	defer errz.Recover(&err)

	//var collectionMustBeCreated bool
	// get collectionId ready
	s.remoteCollectionId, err = remoteStore.CollectionIdByName(ctx, s.Name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		//collectionMustBeCreated = true
		errz.Fatal(usererror.Wrapm(err, fmt.Sprintf("can not sync: %s@%s does not exist on the server", s.Name, s.Version)))
	default:
		errz.Fatal(err)
	}

	absHashCachePath := filepath.Join(bobDir, hashCachePath)
	if s.cache == nil {
		s.cache, err = FromFileOrNew(absHashCachePath)
		errz.Fatal(err)
	}

	// create collection dir if it does not exist
	absCollectionPath := filepath.Join(bobDir, s.Path)
	if !file.Exists(absCollectionPath) {
		fmt.Printf("creating sync collection directory: %s\n", absCollectionPath)
		err = localStore.MakeDir(absCollectionPath)
		errz.Fatal(err)
	}

	// TODO: run list local and remote in parallel
	err = s.cache.Update(absCollectionPath)
	errz.Fatal(err)

	remoteCollection, err := remoteStore.Collection(ctx, s.remoteCollectionId)
	errz.Fatal(err)

	// create the delta
	delta := NewDelta(*s.cache, *remoteCollection)

	fmt.Printf("Sync %s@%s from remote to %s\n", aurora.Bold(s.Name), aurora.Italic(s.Version), absCollectionPath)
	fmt.Println(delta.PullOverview())

	if !force {
		confirm, err := userprompt.Confirm()
		errz.Fatal(err)
		if !confirm {
			return nil
		}
	}

	for _, f := range delta.LocalFilesMissingOnRemote {
		err := localStore.DeleteFile(filepath.Join(bobDir, s.Path, f.LocalPath))
		errz.Fatal(err)
	}
	for _, f := range delta.RemoteFilesMissingOnLocal {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not downlaod from remote")
		}
		if !f.IsDirectory {
			srcReader, err := remoteStore.File(ctx, s.remoteCollectionId, *f.ID)
			errz.Fatal(err)
			err = localStore.WriteFile(filepath.Join(bobDir, s.Path, f.LocalPath), srcReader)
			errz.Fatal(err)
		} else {
			err = localStore.MakeDir(filepath.Join(bobDir, s.Path, f.LocalPath))
			errz.Fatal(err)
		}
	}
	for _, f := range delta.ToBeUpdated {
		if f.ID == nil {
			return fmt.Errorf("ID not available can not downlaod from remote")
		}
		if !f.IsDirectory {
			srcReader, err := remoteStore.File(ctx, s.remoteCollectionId, *f.ID)
			errz.Fatal(err)
			err = localStore.WriteFile(filepath.Join(bobDir, s.Path, f.LocalPath), srcReader)
			errz.Fatal(err)
		} else {
			err = localStore.MakeDir(filepath.Join(bobDir, s.Path, f.LocalPath))
			errz.Fatal(err)
		}
	}

	err = s.cache.Update(filepath.Join(bobDir, s.Path))
	errz.Fatal(err)
	err = s.cache.SaveToFile(absHashCachePath)
	errz.Fatal(err)

	return nil
}

func (s *Sync) ListLocal(bobDir string) (err error) {
	defer errz.Recover(&err)

	absHashCachePath := filepath.Join(bobDir, hashCachePath)
	if s.cache == nil {
		s.cache, err = FromFileOrNew(absHashCachePath)
		errz.Fatal(err)
	}

	err = s.cache.Update(s.Path)
	errz.Fatal(err)
	err = s.cache.SaveToFile(absHashCachePath)
	errz.Fatal(err)

	fmt.Printf("%s@%s (./%s)\n", aurora.Bold(s.Name), aurora.Italic(s.Version), s.Path)
	for _, k := range (*s.cache).SortedKeys() {
		var suffix string
		if (*s.cache)[k].IsDir {
			suffix = "/"
		}

		fmt.Printf("\t%s%s\n", k, suffix)
	}

	return nil

}

func (s *Sync) ListRemote(ctx context.Context, store remotesyncstore.S) (err error) {
	defer errz.Recover(&err)

	// get collectionId ready
	s.remoteCollectionId, err = store.CollectionIdByName(ctx, s.Name, s.Version)

	// check if collections exists
	switch err {
	case nil:
	case remotesyncstore.ErrCollectionNotFound:
		fmt.Printf("Sync collection %s@%s does not exist on the server.\n",
			aurora.Bold(s.Name), aurora.Italic(s.Version))
	default:
		errz.Fatal(err)
	}
	collections, err := store.Collections(ctx)
	errz.Fatal(err)

	for _, c := range collections {
		sort.Sort(c)
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

func (s *Sync) FillRemoteId(ctx context.Context, store remotesyncstore.S) (err error) {
	defer errz.Recover(&err)

	// get collectionId ready
	s.remoteCollectionId, err = store.CollectionIdByName(ctx, s.Name, s.Version)

	errz.Fatal(err)
	return nil
}
