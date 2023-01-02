package bob

import (
	"os"
	"path/filepath"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	nixbuilder "github.com/benchkram/bob/bob/nix-builder"
	"github.com/benchkram/bob/pkg/auth"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/pkg/store/filestore"
)

func DefaultFilestore() (s store.Store, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	return Filestore(home)
}

func Filestore(dir string) (s store.Store, err error) {
	defer errz.Recover(&err)

	storeDir := filepath.Join(dir, global.BobCacheArtifactsDir)
	err = os.MkdirAll(storeDir, 0775)
	errz.Fatal(err)

	return filestore.New(storeDir), nil
}

func MustDefaultFilestore() store.Store {
	s, _ := DefaultFilestore()
	return s
}

func DefaultBuildinfoStore() (s buildinfostore.Store, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	return BuildinfoStore(home)
}

func BuildinfoStore(baseDir string) (s buildinfostore.Store, err error) {
	defer errz.Recover(&err)

	storeDir := filepath.Join(baseDir, global.BobCacheBuildinfoDir)
	err = os.MkdirAll(storeDir, 0775)
	errz.Fatal(err)

	return buildinfostore.NewProtoStore(storeDir), nil
}

func MustDefaultBuildinfoStore() buildinfostore.Store {
	s, _ := DefaultBuildinfoStore()
	return s
}

// Localstore returns the local artifact store
func (b *B) Localstore() store.Store {
	return b.local
}

func AuthStore(baseDir string) (s *auth.Store, err error) {
	defer errz.Recover(&err)

	storeDir := filepath.Join(baseDir, global.BobAuthStoreDir)
	err = os.MkdirAll(storeDir, 0775)
	errz.Fatal(err)

	return auth.New(storeDir), nil
}

func DefaultAuthStore() (s *auth.Store, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	return AuthStore(home)
}

// NixBuilder initialises a new nix builder object with the cache setup
// in the given location.
//
// It's save to use the same base dir as for BuildinfoStore(),
// Filestore() and AuthStore().
func NixBuilder(baseDir string) (_ *nixbuilder.NB, err error) {

	cacheDir := filepath.Join(baseDir, global.BobCacheNixFileName)

	err = os.MkdirAll(filepath.Dir(cacheDir), 0775)
	errz.Fatal(err)

	nixCache, err := nix.NewCacheStore(nix.WithPath(cacheDir))
	errz.Fatal(err)

	shellCache := nix.NewShellCache(filepath.Join(baseDir, global.BobCacheNixShellCacheDir))

	nb := nixbuilder.New(
		nixbuilder.WithCache(nixCache),
		nixbuilder.WithShellCache(shellCache),
	)

	return nb, nil
}

func DefaultNixBuilder() (_ *nixbuilder.NB, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	return NixBuilder(home)
}
