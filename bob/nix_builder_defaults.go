package bob

import (
	"os"
	"path/filepath"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/nix"
)

func DefaultNix() (_ *NixBuilder, err error) {
	defer errz.Recover(&err)

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	cacheDir := filepath.Join(home, global.BobCacheNixFileName)

	err = os.MkdirAll(filepath.Dir(cacheDir), 0775)
	errz.Fatal(err)

	nixCache, err := nix.NewCacheStore(nix.WithPath(cacheDir))
	errz.Fatal(err)

	shellCache := nix.NewShellCache(filepath.Join(home, global.BobCacheNixShellCacheDir))

	return NewNixBuilder(WithCache(nixCache), WithShellCache(shellCache)), nil
}
