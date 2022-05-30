package bob

import (
	"os"
	"path/filepath"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/authstore"
	"github.com/benchkram/bob/pkg/buildinfostore"
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

func BuildinfoStore(dir string) (s buildinfostore.Store, err error) {
	defer errz.Recover(&err)

	storeDir := filepath.Join(dir, global.BobCacheBuildinfoDir)
	err = os.MkdirAll(storeDir, 0775)
	errz.Fatal(err)

	return buildinfostore.New(storeDir), nil
}

func MustDefaultBuildinfoStore() buildinfostore.Store {
	s, _ := DefaultBuildinfoStore()
	return s
}

// Localstore returns the local artifact store
func (b *B) Localstore() store.Store {
	return b.local
}

func AuthStore(dir string) (s authstore.Store, err error) {
	defer errz.Recover(&err)

	storeDir := filepath.Join(dir, global.BobAuthStoreDir)
	err = os.MkdirAll(storeDir, 0775)
	errz.Fatal(err)

	return authstore.New(storeDir), nil
}

func DefaultAuthStore() (s authstore.Store, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	return AuthStore(home)
}
