package nixtest

import (
	"io/ioutil"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"
)

func Bob() (*bob.B, error) {
	nixBuilder, err := NixBuilder()
	if err != nil {
		return nil, err
	}
	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
		bob.WithNixBuilder(nixBuilder),
	)
}

// tmpFiles tracks temporarily created files in these tests
// to be cleaned up at the end.
var tmpFiles []string

func NixBuilder() (*bob.NixBuilder, error) {
	file, err := ioutil.TempFile("", ".nix_cache*")
	if err != nil {
		return nil, err
	}
	name := file.Name()
	file.Close()

	tmpFiles = append(tmpFiles, name)

	cache, err := nix.NewCacheStore(nix.WithPath(name))
	if err != nil {
		return nil, err
	}

	return bob.NewNixBuilder(bob.WithCache(cache)), nil
}
